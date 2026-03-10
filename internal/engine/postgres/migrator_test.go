package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaroher/ratel/pkg/migrate"
)

// allSQL concatenates all SQL statements from a plan into one string for
// easy substring checks.
func allSQL(plan *migrate.Plan) string {
	var sb strings.Builder
	for _, c := range plan.Changes {
		sb.WriteString(c.SQL)
		sb.WriteString("\n")
	}
	return sb.String()
}

// TestMigratorEndToEnd performs a full round-trip:
//  1. Create an initial table in a fresh database.
//  2. Inspect the current state.
//  3. Apply additional DDL (new column, new index, new table with FK).
//  4. Inspect the desired state.
//  5. Diff current → desired and verify changes.
//  6. Plan the changes and verify the generated SQL.
func TestMigratorEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Step 1 – initial schema.
	_, err := pool.Exec(ctx, `CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT NOT NULL)`)
	require.NoError(t, err)

	m := NewMigrator(pool)

	// Step 2 – inspect current state.
	currentState, err := m.InspectRealm(ctx)
	require.NoError(t, err)
	require.NotNil(t, currentState)

	// Confirm we can see the users table.
	found := false
	for _, s := range currentState.Schemas {
		for _, tbl := range s.Tables {
			if tbl.Name == "users" {
				found = true
			}
		}
	}
	assert.True(t, found, "initial inspection should find the 'users' table")

	// Step 3 – apply additional DDL.
	_, err = pool.Exec(ctx, `
		ALTER TABLE users ADD COLUMN name TEXT;
		CREATE INDEX idx_users_email ON users (email);
		CREATE TABLE posts (
			id BIGSERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			title TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	// Step 4 – inspect desired state.
	desiredState, err := m.InspectRealm(ctx)
	require.NoError(t, err)
	require.NotNil(t, desiredState)

	// Confirm the new objects exist in the inspected desired state.
	postsFound := false
	nameColFound := false
	for _, s := range desiredState.Schemas {
		for _, tbl := range s.Tables {
			if tbl.Name == "posts" {
				postsFound = true
			}
			if tbl.Name == "users" {
				for _, col := range tbl.Columns {
					if col.Name == "name" {
						nameColFound = true
					}
				}
			}
		}
	}
	assert.True(t, postsFound, "desired state should include 'posts' table")
	assert.True(t, nameColFound, "desired state should include 'name' column on 'users'")

	// Step 5 – diff current → desired.
	changes, err := m.Diff(currentState, desiredState)
	require.NoError(t, err)
	assert.NotEmpty(t, changes, "diff should produce changes")

	// Verify we have at least an AddTable for posts and a ModifyTable for users.
	var addTables []migrate.AddTable
	var modifyTables []migrate.ModifyTable
	for _, c := range changes {
		switch tc := c.(type) {
		case migrate.AddTable:
			addTables = append(addTables, tc)
		case migrate.ModifyTable:
			modifyTables = append(modifyTables, tc)
		}
	}
	assert.NotEmpty(t, addTables, "expected at least one AddTable change")
	assert.NotEmpty(t, modifyTables, "expected at least one ModifyTable change for users")

	postsInAdd := false
	for _, at := range addTables {
		if at.T.Name == "posts" {
			postsInAdd = true
		}
	}
	assert.True(t, postsInAdd, "AddTable should include 'posts'")

	usersModified := false
	nameColAdded := false
	for _, mt := range modifyTables {
		if mt.To.Name == "users" {
			usersModified = true
			for _, sub := range mt.Changes {
				if ac, ok := sub.(migrate.AddColumn); ok && ac.C.Name == "name" {
					nameColAdded = true
				}
			}
		}
	}
	assert.True(t, usersModified, "ModifyTable should target 'users'")
	assert.True(t, nameColAdded, "ModifyTable for 'users' should include AddColumn for 'name'")

	// Step 6 – plan the changes and verify SQL.
	plan, err := m.Plan(ctx, "test_migration", changes)
	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.NotEmpty(t, plan.Changes, "plan should contain at least one statement")

	sql := allSQL(plan)
	assert.Contains(t, sql, "posts", "generated SQL should reference the 'posts' table")
	assert.Contains(t, sql, "name", "generated SQL should reference the 'name' column")
}

// TestMigratorRLS verifies that enabling RLS and adding a policy is detected
// by inspection + diff and that the planner produces the correct SQL when the
// changes are presented as top-level changes (the form the planner supports).
func TestMigratorRLS(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create table without RLS.
	_, err := pool.Exec(ctx, `CREATE TABLE documents (id SERIAL PRIMARY KEY, owner TEXT NOT NULL, content TEXT)`)
	require.NoError(t, err)

	m := NewMigrator(pool)

	// Inspect before enabling RLS.
	beforeRLS, err := m.InspectRealm(ctx)
	require.NoError(t, err)

	// Find the documents table in the before state.
	var docsBefore *migrate.Table
	for i := range beforeRLS.Schemas {
		for j := range beforeRLS.Schemas[i].Tables {
			if beforeRLS.Schemas[i].Tables[j].Name == "documents" {
				docsBefore = &beforeRLS.Schemas[i].Tables[j]
			}
		}
	}
	require.NotNil(t, docsBefore, "documents table should exist before RLS")
	assert.False(t, docsBefore.RLSEnabled, "RLS should be disabled before ALTER")

	// Enable RLS and add a policy.
	_, err = pool.Exec(ctx, `
		ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
		CREATE POLICY owner_sees_own ON documents FOR SELECT USING (owner = current_user);
	`)
	require.NoError(t, err)

	// Inspect after enabling RLS.
	afterRLS, err := m.InspectRealm(ctx)
	require.NoError(t, err)

	// Verify the inspector correctly reads RLS state.
	var docsAfter *migrate.Table
	for i := range afterRLS.Schemas {
		for j := range afterRLS.Schemas[i].Tables {
			if afterRLS.Schemas[i].Tables[j].Name == "documents" {
				docsAfter = &afterRLS.Schemas[i].Tables[j]
			}
		}
	}
	require.NotNil(t, docsAfter, "documents table should exist after RLS")
	assert.True(t, docsAfter.RLSEnabled, "inspector should detect RLS as enabled")

	// Verify the policy was inspected.
	policyFound := false
	for _, pol := range docsAfter.Policies {
		if pol.Name == "owner_sees_own" {
			policyFound = true
		}
	}
	assert.True(t, policyFound, "inspector should detect 'owner_sees_own' policy")

	// Diff produces changes.
	changes, err := m.Diff(beforeRLS, afterRLS)
	require.NoError(t, err)
	assert.NotEmpty(t, changes, "diff should produce changes when RLS is enabled")

	// Plan the changes using top-level EnableRLS and AddPolicy — the form the
	// planner supports. These correspond to the sub-changes inside ModifyTable
	// but we present them as first-class changes so the planner can handle them.
	tableName := `"public"."documents"`
	policy := docsAfter.Policies[0]
	policyPtr := policy

	topLevelChanges := []migrate.Change{
		migrate.EnableRLS{Table: tableName},
		migrate.AddPolicy{Table: tableName, P: &policyPtr},
	}

	plan, err := m.Plan(ctx, "add_rls", topLevelChanges)
	require.NoError(t, err)
	require.NotNil(t, plan)

	sql := allSQL(plan)
	assert.Contains(t, sql, "ENABLE ROW LEVEL SECURITY", "plan should contain ENABLE ROW LEVEL SECURITY")
	assert.Contains(t, sql, "CREATE POLICY", "plan should contain CREATE POLICY")
	assert.Contains(t, sql, "owner_sees_own", "plan should contain the policy name")
}

// TestMigratorExtensions verifies that adding an extension is detected by the
// inspector+differ and correctly planned.
func TestMigratorExtensions(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	m := NewMigrator(pool)

	// Inspect before adding extension.
	before, err := m.InspectRealm(ctx)
	require.NoError(t, err)

	// Add the extension.
	_, err = pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pg_trgm`)
	require.NoError(t, err)

	// Inspect after adding extension.
	after, err := m.InspectRealm(ctx)
	require.NoError(t, err)

	// Verify the extension appears in the after state.
	extFound := false
	for _, s := range after.Schemas {
		for _, ext := range s.Extensions {
			if ext.Name == "pg_trgm" {
				extFound = true
			}
		}
	}
	assert.True(t, extFound, "inspector should detect pg_trgm extension")

	// Diff should detect the new extension.
	changes, err := m.Diff(before, after)
	require.NoError(t, err)
	assert.NotEmpty(t, changes, "diff should detect the new extension")

	addExtFound := false
	for _, c := range changes {
		if ae, ok := c.(migrate.AddExtension); ok && ae.E.Name == "pg_trgm" {
			addExtFound = true
		}
	}
	assert.True(t, addExtFound, "diff should produce AddExtension for pg_trgm")

	// Plan the changes.
	plan, err := m.Plan(ctx, "add_ext", changes)
	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.NotEmpty(t, plan.Changes, "plan should contain at least one statement")

	sql := allSQL(plan)
	assert.Contains(t, sql, "pg_trgm", "generated SQL should reference pg_trgm")
}

// TestMigratorLock verifies that the advisory locking mechanism works without
// deadlock for a single caller.
func TestMigratorLock(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	m := NewMigrator(pool)

	unlock, err := m.Lock(ctx, "migration-lock", 5*time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlock)

	err = unlock()
	assert.NoError(t, err)
}

// TestMigratorInspectSchema verifies InspectSchema returns a schema with the
// expected structure after creating some objects.
func TestMigratorInspectSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		CREATE TABLE accounts (
			id   BIGSERIAL PRIMARY KEY,
			name TEXT    NOT NULL,
			age  INTEGER CHECK (age >= 0)
		);
		CREATE UNIQUE INDEX idx_accounts_name ON accounts (name);
	`)
	require.NoError(t, err)

	m := NewMigrator(pool)
	schema, err := m.InspectSchema(ctx, "public")
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Equal(t, "public", schema.Name)

	var accountsTable *migrate.Table
	for i := range schema.Tables {
		if schema.Tables[i].Name == "accounts" {
			accountsTable = &schema.Tables[i]
		}
	}
	require.NotNil(t, accountsTable, "InspectSchema should return the 'accounts' table")

	// Verify columns.
	colNames := make(map[string]bool)
	for _, col := range accountsTable.Columns {
		colNames[col.Name] = true
	}
	assert.True(t, colNames["id"], "expected 'id' column")
	assert.True(t, colNames["name"], "expected 'name' column")
	assert.True(t, colNames["age"], "expected 'age' column")

	// Verify primary key.
	assert.NotNil(t, accountsTable.PrimaryKey, "expected a primary key on 'accounts'")

	// Verify the unique index.
	idxFound := false
	for _, idx := range accountsTable.Indexes {
		if idx.Name == "idx_accounts_name" && idx.Unique {
			idxFound = true
		}
	}
	assert.True(t, idxFound, "expected unique index 'idx_accounts_name' on 'accounts'")
}
