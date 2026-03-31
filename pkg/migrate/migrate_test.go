package migrate

import (
	"context"
	"testing"
	"testing/fstest"
	"time"

	atlasmigrate "ariga.io/atlas/sql/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	return pool, func() {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
	}
}

// makeMigrationFS creates an fstest.MapFS with migration files and a valid atlas.sum.
func makeMigrationFS(t *testing.T, files map[string]string) fstest.MapFS {
	t.Helper()

	var migrationFiles []atlasmigrate.File
	for name, content := range files {
		migrationFiles = append(migrationFiles, atlasmigrate.NewLocalFile(name, []byte(content)))
	}

	hashFile, err := atlasmigrate.NewHashFile(migrationFiles)
	require.NoError(t, err)

	sumBytes, err := hashFile.MarshalText()
	require.NoError(t, err)

	fs := fstest.MapFS{}
	for name, content := range files {
		fs[name] = &fstest.MapFile{Data: []byte(content)}
	}
	fs["atlas.sum"] = &fstest.MapFile{Data: sumBytes}

	return fs
}

// TestMigrateIdempotent verifies that calling Migrate twice with the same
// migrations on the same database does not fail.
func TestMigrateIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	lg := zap.NewNop()

	migrations := makeMigrationFS(t, map[string]string{
		"20260101000000_init.sql": `CREATE TABLE test_users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);`,
	})

	// First call — should apply the migration.
	err := Migrate(pool, lg, "", migrations)
	require.NoError(t, err, "first Migrate call should succeed")

	// Verify table was created.
	var exists bool
	err = pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'test_users')`).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "test_users table should exist after first migration")

	// Second call — should be idempotent (no pending files).
	err = Migrate(pool, lg, "", migrations)
	require.NoError(t, err, "second Migrate call should succeed (idempotent)")
}

// TestMigrateRetryAfterPartialFailure reproduces the crash from production:
// a migration with multiple statements fails partway through (e.g. table
// already exists), then a subsequent Migrate call panics with
// "index out of range [0] with length 0" because PartialHashes was stored
// as NULL instead of a proper array.
func TestMigrateRetryAfterPartialFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	lg := zap.NewNop()
	ctx := context.Background()

	// Pre-create a table that will conflict with the migration.
	_, err := pool.Exec(ctx, `CREATE TABLE conflicting_table (id SERIAL PRIMARY KEY)`)
	require.NoError(t, err)

	// Migration has 3 statements: first two succeed, third conflicts.
	migrations := makeMigrationFS(t, map[string]string{
		"20260101000000_init.sql": "CREATE TABLE aaa_first (id SERIAL PRIMARY KEY);\n" +
			"CREATE TABLE bbb_second (id SERIAL PRIMARY KEY);\n" +
			"CREATE TABLE conflicting_table (id SERIAL PRIMARY KEY);\n",
	})

	// First call — should fail on the third statement.
	err = Migrate(pool, lg, "", migrations)
	require.Error(t, err, "first Migrate should fail due to conflict")
	assert.Contains(t, err.Error(), "conflicting_table")

	// Verify first two tables were created.
	for _, tbl := range []string{"aaa_first", "bbb_second"} {
		var exists bool
		err = pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`, tbl).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "table %s should exist after partial migration", tbl)
	}

	// Drop the conflicting table so the retry can succeed.
	_, err = pool.Exec(ctx, `DROP TABLE conflicting_table`)
	require.NoError(t, err)

	// Second call — should resume and apply the remaining statement.
	// Before the fix, this panicked: "index out of range [0] with length 0"
	// because PartialHashes was NULL in the database.
	require.NotPanics(t, func() {
		err = Migrate(pool, lg, "", migrations)
	}, "second Migrate should not panic")
	require.NoError(t, err, "second Migrate should succeed after removing conflict")

	// Verify the previously-conflicting table now exists (re-created by migration).
	var exists bool
	err = pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'conflicting_table')`).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "conflicting_table should exist after successful retry")
}

// TestMigrateMultipleFS verifies that passing multiple embed.FS (each with its
// own atlas.sum) works correctly — simulates the komeet-backend pattern where
// main migrations and outbox migrations are separate packages.
func TestMigrateMultipleFS(t *testing.T) {
	if testing.Short() {
		t.Skip("requires docker")
	}

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	lg := zap.NewNop()

	// First FS — main migrations (has its own atlas.sum).
	mainFS := makeMigrationFS(t, map[string]string{
		"20260101000000_schema.sql": `CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT NOT NULL);`,
	})

	// Second FS — outbox migrations (has its own atlas.sum).
	outboxFS := makeMigrationFS(t, map[string]string{
		"20240726143237_outbox.sql": `CREATE TABLE outbox (id UUID PRIMARY KEY, message BYTEA NOT NULL);`,
	})

	// Single Migrate call with both FS merged — should apply all files.
	err := Migrate(pool, lg, "", mainFS, outboxFS)
	require.NoError(t, err, "Migrate with multiple FS should succeed")

	// Verify both tables created.
	ctx := context.Background()
	for _, tbl := range []string{"users", "outbox"} {
		var exists bool
		err = pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`, tbl).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "table %s should exist", tbl)
	}

	// Second call — idempotent.
	err = Migrate(pool, lg, "", mainFS, outboxFS)
	require.NoError(t, err, "second Migrate with multiple FS should succeed (idempotent)")
}
