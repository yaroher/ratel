package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yaroher/ratel/pkg/migrate"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	return pool, func() {
		pool.Close()
		_ = container.Terminate(ctx)
	}
}

func exec(t *testing.T, pool *pgxpool.Pool, sql string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), sql)
	require.NoError(t, err)
}

// TestInspectTables verifies that tables, columns, indexes, foreign keys and
// check constraints are fully populated by the inspector.
func TestInspectTables(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	exec(t, pool, `
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT,
    age INTEGER CHECK (age > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_active ON users (email) WHERE age > 18;

CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT
);
CREATE INDEX idx_posts_user ON posts (user_id);
`)

	ins := NewInspector(pool)
	schema, err := ins.InspectSchema(context.Background(), "public")
	require.NoError(t, err)

	// 2 tables
	require.Len(t, schema.Tables, 2)

	// find tables by name
	tableMap := make(map[string]*struct{ idx int })
	for i := range schema.Tables {
		tableMap[schema.Tables[i].Name] = &struct{ idx int }{i}
	}
	require.Contains(t, tableMap, "users")
	require.Contains(t, tableMap, "posts")

	// ---- users table ----
	var users, posts *migrate.Table // will resolve below
	for i := range schema.Tables {
		switch schema.Tables[i].Name {
		case "users":
			users = &schema.Tables[i]
		case "posts":
			posts = &schema.Tables[i]
		}
	}
	require.NotNil(t, users)
	require.NotNil(t, posts)

	// 5 columns
	assert.Len(t, users.Columns, 5)

	colMap := make(map[string]migrate.Column)
	for _, c := range users.Columns {
		colMap[c.Name] = c
	}

	// id: serial -> int4, not nullable
	idCol, ok := colMap["id"]
	require.True(t, ok, "column 'id' not found")
	assert.Equal(t, "int4", idCol.Type)
	assert.False(t, idCol.Nullable)

	// email: text, not nullable
	emailCol, ok := colMap["email"]
	require.True(t, ok, "column 'email' not found")
	assert.Equal(t, "text", emailCol.Type)
	assert.False(t, emailCol.Nullable)

	// name: text, nullable
	nameCol, ok := colMap["name"]
	require.True(t, ok, "column 'name' not found")
	assert.Equal(t, "text", nameCol.Type)
	assert.True(t, nameCol.Nullable)

	// age: int4, nullable (no NOT NULL constraint)
	ageCol, ok := colMap["age"]
	require.True(t, ok, "column 'age' not found")
	assert.Equal(t, "int4", ageCol.Type)
	assert.True(t, ageCol.Nullable)

	// created_at: timestamptz, not nullable, has default now()
	createdAtCol, ok := colMap["created_at"]
	require.True(t, ok, "column 'created_at' not found")
	assert.Equal(t, "timestamptz", createdAtCol.Type)
	assert.False(t, createdAtCol.Nullable)
	assert.NotEmpty(t, createdAtCol.Default, "created_at should have a default")
	assert.Contains(t, strings.ToLower(createdAtCol.Default), "now")

	// Primary key on "id"
	require.NotNil(t, users.PrimaryKey)
	require.Len(t, users.PrimaryKey.Columns, 1)
	assert.Equal(t, "id", users.PrimaryKey.Columns[0].Name)

	// 2 secondary indexes: idx_users_email and idx_users_active
	assert.Len(t, users.Indexes, 2)
	idxMap := make(map[string]migrate.Index)
	for _, idx := range users.Indexes {
		idxMap[idx.Name] = idx
	}

	assert.Contains(t, idxMap, "idx_users_email")
	assert.Contains(t, idxMap, "idx_users_active")

	activeIdx := idxMap["idx_users_active"]
	assert.NotEmpty(t, activeIdx.Where, "partial index should have a WHERE clause")
	assert.Contains(t, activeIdx.Where, "18")

	// 1 check constraint (age > 0)
	require.Len(t, users.Checks, 1)
	assert.Contains(t, users.Checks[0].Expr, "age")
	assert.Contains(t, users.Checks[0].Expr, "0")

	// ---- posts table ----
	// FK to users with CASCADE delete
	require.Len(t, posts.ForeignKeys, 1)
	fk := posts.ForeignKeys[0]
	assert.Equal(t, "users", fk.RefTable)
	assert.Equal(t, "CASCADE", fk.OnDelete)
	assert.Contains(t, fk.Columns, "user_id")
	assert.Contains(t, fk.RefColumns, "id")
}

// TestInspectRLS verifies that row-level security flags and policies are
// correctly read from the database.
func TestInspectRLS(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	exec(t, pool, `
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    owner TEXT NOT NULL,
    content TEXT
);
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents FORCE ROW LEVEL SECURITY;
CREATE POLICY owner_policy ON documents
    AS PERMISSIVE
    FOR ALL
    TO PUBLIC
    USING (owner = current_user);
CREATE POLICY insert_policy ON documents
    FOR INSERT
    WITH CHECK (owner = current_user);
`)

	ins := NewInspector(pool)
	schema, err := ins.InspectSchema(context.Background(), "public")
	require.NoError(t, err)

	require.Len(t, schema.Tables, 1)
	docs := schema.Tables[0]
	assert.Equal(t, "documents", docs.Name)

	assert.True(t, docs.RLSEnabled, "RLSEnabled should be true")
	assert.True(t, docs.RLSForced, "RLSForced should be true")

	require.Len(t, docs.Policies, 2)

	polMap := make(map[string]migrate.Policy)
	for _, p := range docs.Policies {
		polMap[p.Name] = p
	}

	// owner_policy: PERMISSIVE, ALL, has USING
	ownerPol, ok := polMap["owner_policy"]
	require.True(t, ok, "owner_policy not found")
	assert.True(t, ownerPol.Permissive)
	assert.Equal(t, "ALL", ownerPol.Command)
	assert.NotEmpty(t, ownerPol.Using)
	assert.Contains(t, strings.ToLower(ownerPol.Using), "current_user")

	// insert_policy: INSERT, has WITH CHECK
	insertPol, ok := polMap["insert_policy"]
	require.True(t, ok, "insert_policy not found")
	assert.Equal(t, "INSERT", insertPol.Command)
	assert.NotEmpty(t, insertPol.WithCheck)
	assert.Contains(t, strings.ToLower(insertPol.WithCheck), "current_user")
}

// TestInspectExtensions verifies that installed extensions are listed.
func TestInspectExtensions(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	exec(t, pool, `CREATE EXTENSION IF NOT EXISTS pg_trgm;`)

	ins := NewInspector(pool)
	schema, err := ins.InspectSchema(context.Background(), "public")
	require.NoError(t, err)

	found := false
	for _, ext := range schema.Extensions {
		if ext.Name == "pg_trgm" {
			found = true
			assert.NotEmpty(t, ext.Version)
			break
		}
	}
	assert.True(t, found, "pg_trgm extension not found in inspected extensions")
}

// TestInspectFunctions verifies that user-defined functions are correctly
// populated with name, return type, language, and volatility.
func TestInspectFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	exec(t, pool, `
CREATE FUNCTION add_numbers(a INTEGER, b INTEGER) RETURNS INTEGER
LANGUAGE plpgsql IMMUTABLE AS $$
BEGIN
    RETURN a + b;
END;
$$;
`)

	ins := NewInspector(pool)
	schema, err := ins.InspectSchema(context.Background(), "public")
	require.NoError(t, err)

	var fn *migrate.Function
	for i := range schema.Functions {
		if schema.Functions[i].Name == "add_numbers" {
			fn = &schema.Functions[i]
			break
		}
	}
	require.NotNil(t, fn, "function 'add_numbers' not found")

	assert.Equal(t, "add_numbers", fn.Name)
	assert.Equal(t, "integer", fn.ReturnType)
	assert.Equal(t, "plpgsql", fn.Language)
	assert.Equal(t, "IMMUTABLE", fn.Volatility)

	// 2 arguments
	require.Len(t, fn.Args, 2)
	assert.Equal(t, "a", fn.Args[0].Name)
	assert.Equal(t, "integer", fn.Args[0].Type)
	assert.Equal(t, "b", fn.Args[1].Name)
	assert.Equal(t, "integer", fn.Args[1].Type)
}

// TestInspectTriggers verifies that triggers are correctly populated with
// timing, events, and the associated function name.
func TestInspectTriggers(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	exec(t, pool, `
CREATE TABLE audit_log (id SERIAL PRIMARY KEY, action TEXT);
CREATE TABLE items (id SERIAL PRIMARY KEY, name TEXT);

CREATE FUNCTION log_changes() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO audit_log (action) VALUES (TG_OP);
    RETURN NEW;
END;
$$;

CREATE TRIGGER items_audit
    AFTER INSERT OR UPDATE ON items
    FOR EACH ROW EXECUTE FUNCTION log_changes();
`)

	ins := NewInspector(pool)
	schema, err := ins.InspectSchema(context.Background(), "public")
	require.NoError(t, err)

	// Find the items table
	var itemsTable *migrate.Table
	for i := range schema.Tables {
		if schema.Tables[i].Name == "items" {
			itemsTable = &schema.Tables[i]
			break
		}
	}
	require.NotNil(t, itemsTable, "table 'items' not found")

	require.Len(t, itemsTable.Triggers, 1)
	trig := itemsTable.Triggers[0]

	assert.Equal(t, "items_audit", trig.Name)
	assert.Equal(t, "AFTER", trig.Timing)
	assert.True(t, trig.ForEachRow)
	assert.Equal(t, "log_changes", trig.Function)

	// Events: INSERT and UPDATE (order: INSERT, DELETE, UPDATE as appended by bitmask)
	assert.Len(t, trig.Events, 2)
	eventSet := make(map[string]bool)
	for _, e := range trig.Events {
		eventSet[e] = true
	}
	assert.True(t, eventSet["INSERT"], "INSERT event expected")
	assert.True(t, eventSet["UPDATE"], "UPDATE event expected")
}
