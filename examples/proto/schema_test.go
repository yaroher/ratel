package storepb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	pgEngine "github.com/yaroher/ratel/internal/engine/postgres"
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/migrate"
)

// allTables returns all proto-generated table sqlers.
func allTables() []ddl.SchemaSqler {
	return []ddl.SchemaSqler{
		Currencys,
		Users,
		Profiles,
		Categorys,
		Tags,
		Products,
		Orders,
		OrderItems,
	}
}

func TestSchemaIncludesAdditionalSQL(t *testing.T) {
	sqlers := allTables()
	sqlers = append(sqlers, AdditionalSQL...)

	sql, err := ddl.SchemaSortedSQL(sqlers...)
	if err != nil {
		t.Fatalf("SchemaSortedSQL: %v", err)
	}

	expected := []string{
		"CREATE EXTENSION IF NOT EXISTS pg_trgm",
		"CREATE OR REPLACE FUNCTION store.set_updated_at()",
	}
	for _, s := range expected {
		if !strings.Contains(sql, s) {
			t.Errorf("expected AdditionalSQL statement not found: %s", s)
		}
	}
}

func TestAdditionalSQLProducesMigrationChanges(t *testing.T) {
	current := &migrate.SchemaState{
		Schemas: []migrate.Schema{
			{Name: "public"},
			{Name: "store"},
		},
	}

	desired := &migrate.SchemaState{
		Schemas: []migrate.Schema{
			{
				Name: "public",
				Extensions: []migrate.Extension{
					{Name: "pg_trgm", Schema: "public"},
				},
			},
			{
				Name: "store",
				Functions: []migrate.Function{
					{
						Name:       "set_updated_at",
						Schema:     "store",
						ReturnType: "trigger",
						Language:   "plpgsql",
						Body:       "BEGIN NEW.updated_at = now(); RETURN NEW; END;",
					},
				},
			},
		},
	}

	differ := &pgEngine.Differ{}
	changes, err := differ.Diff(current, desired)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	var hasExtension, hasFunction bool
	for _, c := range changes {
		switch ch := c.(type) {
		case migrate.AddExtension:
			if ch.E.Name == "pg_trgm" {
				hasExtension = true
			}
		case migrate.AddFunction:
			if ch.F.Name == "set_updated_at" {
				hasFunction = true
			}
		}
	}

	if !hasExtension {
		t.Error("migration should include AddExtension for pg_trgm")
	}
	if !hasFunction {
		t.Error("migration should include AddFunction for set_updated_at")
	}
}

// TestAdditionalSQLMigrationEndToEnd reproduces the full ratel diff flow:
// generate SQL with AdditionalSQL → apply to postgres → inspect → diff → plan
// → verify migration contains CREATE EXTENSION and CREATE FUNCTION.
func TestAdditionalSQLMigrationEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	ctx := context.Background()

	// 1. Generate SQL from proto models + AdditionalSQL (same as ratel schema).
	sqlers := allTables()
	sqlers = append(sqlers, AdditionalSQL...)
	stmts, err := ddl.SchemaSortedStatements(sqlers...)
	if err != nil {
		t.Fatalf("SchemaSortedStatements: %v", err)
	}
	desiredSQL := strings.Join(stmts, ";\n") + ";"

	// 2. Start postgres container.
	container, err := pgContainer.Run(ctx, "postgres:16-alpine",
		pgContainer.WithDatabase("test"),
		pgContainer.WithUsername("test"),
		pgContainer.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start container: %v", err)
	}
	defer func() { _ = container.Terminate(ctx) }()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	migrator := pgEngine.NewMigrator(pool)

	// 3. Inspect empty state (current — no migrations yet).
	currentState, err := migrator.InspectRealm(ctx)
	if err != nil {
		t.Fatalf("InspectRealm (current): %v", err)
	}

	// 4. Apply desired SQL to the database.
	if _, err := pool.Exec(ctx, desiredSQL); err != nil {
		t.Fatalf("apply desired SQL: %v", err)
	}

	// 5. Inspect desired state.
	desiredState, err := migrator.InspectRealm(ctx)
	if err != nil {
		t.Fatalf("InspectRealm (desired): %v", err)
	}

	// 6. Verify inspector sees the extension and function.
	var extFound, funcFound bool
	for _, s := range desiredState.Schemas {
		for _, ext := range s.Extensions {
			if ext.Name == "pg_trgm" {
				extFound = true
			}
		}
		for _, fn := range s.Functions {
			if fn.Name == "set_updated_at" {
				funcFound = true
			}
		}
	}
	if !extFound {
		t.Error("inspector should detect pg_trgm extension")
	}
	if !funcFound {
		t.Error("inspector should detect set_updated_at function")
	}

	// 7. Diff current → desired.
	changes, err := migrator.Diff(currentState, desiredState)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if len(changes) == 0 {
		t.Fatal("diff should produce changes")
	}

	var addExt, addFunc bool
	for _, c := range changes {
		switch ch := c.(type) {
		case migrate.AddExtension:
			if ch.E.Name == "pg_trgm" {
				addExt = true
			}
		case migrate.AddFunction:
			if ch.F.Name == "set_updated_at" {
				addFunc = true
			}
		}
	}
	if !addExt {
		t.Error("diff should include AddExtension for pg_trgm")
	}
	if !addFunc {
		t.Error("diff should include AddFunction for set_updated_at")
	}

	// 8. Plan → verify migration SQL.
	plan, err := migrator.Plan(ctx, "init", changes)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	var migrationSQL strings.Builder
	for _, c := range plan.Changes {
		migrationSQL.WriteString(c.SQL)
		migrationSQL.WriteString("\n")
	}
	sql := migrationSQL.String()

	if !strings.Contains(sql, "pg_trgm") {
		t.Errorf("migration SQL should contain pg_trgm extension, got:\n%s", sql)
	}
	if !strings.Contains(sql, "set_updated_at") {
		t.Errorf("migration SQL should contain set_updated_at function, got:\n%s", sql)
	}

	t.Logf("Migration SQL (%d statements):\n%s", len(plan.Changes), sql)
}
