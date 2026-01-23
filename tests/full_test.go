package tests

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yaroher/ratel/ddl"
	"github.com/yaroher/ratel/dml/set"
	"github.com/yaroher/ratel/schema"
)

// Table and Column aliases
type TableAlias string

func (t TableAlias) String() string { return string(t) }

type ColumnAlias string

func (c ColumnAlias) String() string { return string(c) }

const (
	UsersTable TableAlias = "users"
)

const (
	ColID        ColumnAlias = "id"
	ColName      ColumnAlias = "name"
	ColEmail     ColumnAlias = "email"
	ColAge       ColumnAlias = "age"
	ColCreatedAt ColumnAlias = "created_at"
)

// User represents a row in the users table
type User struct {
	ID        int32
	Name      string
	Email     string
	Age       *int32
	CreatedAt time.Time
}

// GetTarget implements exec.Scanner interface
func (u *User) GetTarget(field string) func() any {
	switch ColumnAlias(field) {
	case ColID:
		return func() any { return &u.ID }
	case ColName:
		return func() any { return &u.Name }
	case ColEmail:
		return func() any { return &u.Email }
	case ColAge:
		return func() any { return &u.Age }
	case ColCreatedAt:
		return func() any { return &u.CreatedAt }
	default:
		panic("unknown field: " + field)
	}
}

// GetSetter implements exec.Scanner interface
func (u *User) GetSetter(field ColumnAlias) func() set.ValueSetter[ColumnAlias] {
	switch field {
	case ColID:
		return func() set.ValueSetter[ColumnAlias] { return set.NewSetter(field, u.ID) }
	case ColName:
		return func() set.ValueSetter[ColumnAlias] { return set.NewSetter(field, u.Name) }
	case ColEmail:
		return func() set.ValueSetter[ColumnAlias] { return set.NewSetter(field, u.Email) }
	case ColAge:
		return func() set.ValueSetter[ColumnAlias] { return set.NewSetter(field, u.Age) }
	case ColCreatedAt:
		return func() set.ValueSetter[ColumnAlias] { return set.NewSetter(field, u.CreatedAt) }
	default:
		panic("unknown field: " + string(field))
	}
}

// GetValue implements exec.Scanner interface
func (u *User) GetValue(field ColumnAlias) func() any {
	switch field {
	case ColID:
		return func() any { return u.ID }
	case ColName:
		return func() any { return u.Name }
	case ColEmail:
		return func() any { return u.Email }
	case ColAge:
		return func() any { return u.Age }
	case ColCreatedAt:
		return func() any { return u.CreatedAt }
	default:
		panic("unknown field: " + string(field))
	}
}

// setupPostgres creates a PostgreSQL container and returns the connection pool
func setupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	cleanup := func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

// createUsersTable creates the users table schema
func createUsersTable() *schema.Table[TableAlias, ColumnAlias, *User] {
	return schema.NewTable[TableAlias, ColumnAlias, *User](
		UsersTable,
		func() *User { return &User{} },
		ddl.NewColumnDDL(
			ColID,
			ddl.SERIAL,
			ddl.WithPrimaryKey[ColumnAlias](),
		),
		ddl.NewColumnDDL(
			ColName,
			ddl.TEXT,
			ddl.WithNotNull[ColumnAlias](),
		),
		ddl.NewColumnDDL(
			ColEmail,
			ddl.TEXT,
			ddl.WithNotNull[ColumnAlias](),
		),
		ddl.NewColumnDDL(
			ColAge,
			ddl.INTEGER,
			ddl.WithNullable[ColumnAlias](),
		),
		ddl.NewColumnDDL(
			ColCreatedAt,
			ddl.TIMESTAMPTZ,
			ddl.WithNotNull[ColumnAlias](),
			ddl.WithDefault[ColumnAlias]("CURRENT_TIMESTAMP"),
		),
	)
}

func TestFullIntegration(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()

	// Create table schema
	users := createUsersTable()

	// Add unique constraint on email
	users.TableDDL.Unique([]ColumnAlias{ColEmail})

	// Create table in database
	for _, sql := range users.TableDDL.SchemaSql() {
		t.Logf("Executing DDL: %s", sql)
		if _, err := pool.Exec(ctx, sql); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
	}

	t.Run("Insert", func(t *testing.T) {
		// Define columns for semantic API
		name := schema.TextColumn[ColumnAlias](ColName)
		email := schema.TextColumn[ColumnAlias](ColEmail)
		age := schema.NullIntegerColumn[ColumnAlias](ColAge)

		// Insert a user
		age30 := int32(30)
		insertQuery := users.Insert().From(
			name.Set("John Doe"),
			email.Set("john@example.com"),
			age.Set(&age30),
		).Returning(ColID, ColName, ColEmail, ColAge, ColCreatedAt)

		sql, args := insertQuery.Build()
		t.Logf("Insert SQL: %s, Args: %v", sql, args)

		user, err := users.QueryRow(ctx, pool, insertQuery)
		if err != nil {
			t.Fatalf("failed to insert user: %v", err)
		}

		if user.Name != "John Doe" {
			t.Errorf("expected name 'John Doe', got '%s'", user.Name)
		}
		if user.Email != "john@example.com" {
			t.Errorf("expected email 'john@example.com', got '%s'", user.Email)
		}
		if user.Age == nil || *user.Age != 30 {
			t.Errorf("expected age 30, got %v", user.Age)
		}
		if user.ID == 0 {
			t.Error("expected non-zero ID")
		}

		t.Logf("Inserted user: ID=%d, Name=%s, Email=%s, Age=%d, CreatedAt=%v",
			user.ID, user.Name, user.Email, *user.Age, user.CreatedAt)
	})

	t.Run("Insert Multiple Users", func(t *testing.T) {
		name := schema.TextColumn[ColumnAlias](ColName)
		email := schema.TextColumn[ColumnAlias](ColEmail)
		age := schema.NullIntegerColumn[ColumnAlias](ColAge)

		users := []struct {
			name  string
			email string
			age   *int32
		}{
			{"Alice Smith", "alice@example.com", ptr(int32(25))},
			{"Bob Johnson", "bob@example.com", ptr(int32(35))},
			{"Charlie Brown", "charlie@example.com", nil},
		}

		for _, u := range users {
			query := createUsersTable().Insert().From(
				name.Set(u.name),
				email.Set(u.email),
				age.Set(u.age),
			)
			sql, args := query.Build()
			t.Logf("Insert SQL: %s, Args: %v", sql, args)

			_, err := pool.Exec(ctx, sql, args...)
			if err != nil {
				t.Fatalf("failed to insert user %s: %v", u.name, err)
			}
		}
	})

	t.Run("Select All", func(t *testing.T) {
		selectQuery := users.SelectAll().OrderByASC(ColID)

		sql, args := selectQuery.Build()
		t.Logf("Select SQL: %s, Args: %v", sql, args)

		userList, err := users.Query(ctx, pool, selectQuery)
		if err != nil {
			t.Fatalf("failed to select users: %v", err)
		}

		if len(userList) != 4 {
			t.Errorf("expected 4 users, got %d", len(userList))
		}

		for i, u := range userList {
			t.Logf("User[%d]: ID=%d, Name=%s, Email=%s, Age=%v",
				i, u.ID, u.Name, u.Email, u.Age)
		}
	})

	t.Run("Select With Where", func(t *testing.T) {
		email := schema.TextColumn[ColumnAlias](ColEmail)

		selectQuery := users.SelectAll().
			Where(email.Eq("alice@example.com"))

		sql, args := selectQuery.Build()
		t.Logf("Select Where SQL: %s, Args: %v", sql, args)

		user, err := users.QueryRow(ctx, pool, selectQuery)
		if err != nil {
			t.Fatalf("failed to select user: %v", err)
		}

		if user.Name != "Alice Smith" {
			t.Errorf("expected name 'Alice Smith', got '%s'", user.Name)
		}

		t.Logf("Found user: Name=%s, Email=%s", user.Name, user.Email)
	})

	t.Run("Select With Complex Where", func(t *testing.T) {
		age := schema.NullIntegerColumn[ColumnAlias](ColAge)
		name := schema.TextColumn[ColumnAlias](ColName)

		selectQuery := users.SelectAll().
			Where(
				users.And(
					age.IsNotNull(),
					age.Gte(ptr(int32(30))),
					name.Like("%John%"),
				),
			).
			OrderByDESC(ColAge)

		sql, args := selectQuery.Build()
		t.Logf("Complex Select SQL: %s, Args: %v", sql, args)

		userList, err := users.Query(ctx, pool, selectQuery)
		if err != nil {
			t.Fatalf("failed to select users: %v", err)
		}

		for _, u := range userList {
			t.Logf("Found user: Name=%s, Age=%v", u.Name, u.Age)
			if u.Age == nil || *u.Age < 30 {
				t.Errorf("expected age >= 30, got %v", u.Age)
			}
		}
	})

	t.Run("Update", func(t *testing.T) {
		email := schema.TextColumn[ColumnAlias](ColEmail)
		age := schema.NullIntegerColumn[ColumnAlias](ColAge)

		newAge := int32(26)
		updateQuery := users.Update().
			Set(age.Set(&newAge)).
			Where(email.Eq("alice@example.com")).
			ReturningAll()

		sql, args := updateQuery.Build()
		t.Logf("Update SQL: %s, Args: %v", sql, args)

		user, err := users.QueryRow(ctx, pool, updateQuery)
		if err != nil {
			t.Fatalf("failed to update user: %v", err)
		}

		if user.Age == nil || *user.Age != 26 {
			t.Errorf("expected age 26, got %v", user.Age)
		}

		t.Logf("Updated user: Name=%s, Age=%d", user.Name, *user.Age)
	})

	t.Run("Delete", func(t *testing.T) {
		email := schema.TextColumn[ColumnAlias](ColEmail)

		deleteQuery := users.Delete().
			Where(email.Eq("charlie@example.com"))

		sql, args := deleteQuery.Build()
		t.Logf("Delete SQL: %s, Args: %v", sql, args)

		affected, err := users.Execute(ctx, pool, deleteQuery)
		if err != nil {
			t.Fatalf("failed to delete user: %v", err)
		}

		if affected != 1 {
			t.Errorf("expected 1 row affected, got %d", affected)
		}

		t.Logf("Deleted %d user(s)", affected)

		// Verify deletion
		selectQuery := users.Select1().Where(email.Eq("charlie@example.com"))
		_, err = users.QueryRow(ctx, pool, selectQuery)
		if err == nil {
			t.Error("expected error when selecting deleted user")
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		selectQuery := users.SelectAll().
			OrderByASC(ColID).
			Limit(2).
			Offset(1)

		sql, args := selectQuery.Build()
		t.Logf("Pagination SQL: %s, Args: %v", sql, args)

		userList, err := users.Query(ctx, pool, selectQuery)
		if err != nil {
			t.Fatalf("failed to select users with pagination: %v", err)
		}

		if len(userList) != 2 {
			t.Errorf("expected 2 users, got %d", len(userList))
		}

		for _, u := range userList {
			t.Logf("Paginated user: Name=%s", u.Name)
		}
	})

	t.Run("Insert On Conflict Do Nothing", func(t *testing.T) {
		name := schema.TextColumn[ColumnAlias](ColName)
		email := schema.TextColumn[ColumnAlias](ColEmail)

		// Try to insert duplicate email
		insertQuery := users.Insert().From(
			name.Set("John Duplicate"),
			email.Set("john@example.com"),
		).OnConflict(ColEmail).DoNothing()

		sql, args := insertQuery.Build()
		t.Logf("Insert On Conflict SQL: %s, Args: %v", sql, args)

		_, err := pool.Exec(ctx, sql, args...)
		if err != nil {
			t.Fatalf("failed to execute insert on conflict: %v", err)
		}

		// Verify original user is unchanged
		selectQuery := users.SelectAll().Where(
			schema.TextColumn[ColumnAlias](ColEmail).Eq("john@example.com"),
		)
		user, err := users.QueryRow(ctx, pool, selectQuery)
		if err != nil {
			t.Fatalf("failed to select user: %v", err)
		}

		if user.Name != "John Doe" {
			t.Errorf("expected name 'John Doe', got '%s'", user.Name)
		}
	})
}

// Helper function to create pointer
func ptr[T any](v T) *T {
	return &v
}
