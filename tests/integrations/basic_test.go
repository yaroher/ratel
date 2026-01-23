package integrations

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yaroher/ratel/dml"
	"github.com/yaroher/ratel/next"
)

// userField is a column alias type for the users table
type userField string

func (f userField) String() string { return string(f) }

const (
	userFieldID    userField = "id"
	userFieldName  userField = "name"
	userFieldEmail userField = "email"
	userFieldAge   userField = "age"
)

// TestUser demonstrates basic CRUD operations with testcontainers
func TestUser(t *testing.T) {
	ctx := context.Background()

	// Создаем PostgreSQL контейнер
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Получаем connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	// Подключаемся к БД
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// Создаем тестовую таблицу
	_, err = pool.Exec(ctx, `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем колонки
	nameCol := next.StringColumn[userField](userFieldName)
	emailCol := next.StringColumn[userField](userFieldEmail)
	ageCol := next.Nullable(next.IntegerColumn[userField](userFieldAge))

	t.Run("Insert", func(t *testing.T) {
		// Вставляем пользователя
		insertQuery := &dml.InsertQuery[userField]{
			BaseQuery: dml.BaseQuery[userField]{
				Ta: "users",
			},
		}
		insertQuery.From(
			nameCol.Set("John Doe"),
			emailCol.Set("john@example.com"),
			ageCol.Set(int32(30)),
		)

		sql, args := insertQuery.Build()
		t.Logf("SQL: %s", sql)
		t.Logf("Args: %v", args)

		tag, err := pool.Exec(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}

		if tag.RowsAffected() != 1 {
			t.Fatalf("expected 1 row affected, got %d", tag.RowsAffected())
		}
	})

	t.Run("Select", func(t *testing.T) {
		// Выбираем пользователя
		selectQuery := &dml.SelectQuery[userField]{
			BaseQuery: dml.BaseQuery[userField]{
				Ta:          "users",
				UsingFields: []userField{userFieldID, userFieldName, userFieldEmail, userFieldAge},
			},
		}
		selectQuery.Where(nameCol.Eq("John Doe"))

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)
		t.Logf("Args: %v", args)

		var id, age int32
		var name, email string

		err := pool.QueryRow(ctx, sql, args...).Scan(&id, &name, &email, &age)
		if err != nil {
			t.Fatal(err)
		}

		if name != "John Doe" {
			t.Fatalf("expected name 'John Doe', got '%s'", name)
		}
		if email != "john@example.com" {
			t.Fatalf("expected email 'john@example.com', got '%s'", email)
		}
		if age != 30 {
			t.Fatalf("expected age 30, got %d", age)
		}

		t.Logf("User: id=%d, name=%s, email=%s, age=%d", id, name, email, age)
	})

	t.Run("Update", func(t *testing.T) {
		// Обновляем возраст
		updateQuery := &dml.UpdateQuery[userField]{
			BaseQuery: dml.BaseQuery[userField]{
				Ta: "users",
			},
		}
		updateQuery.Set(ageCol.Set(int32(31))).Where(emailCol.Eq("john@example.com"))

		sql, args := updateQuery.Build()
		t.Logf("SQL: %s", sql)
		t.Logf("Args: %v", args)

		tag, err := pool.Exec(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}

		if tag.RowsAffected() != 1 {
			t.Fatalf("expected 1 row affected, got %d", tag.RowsAffected())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Удаляем пользователя
		deleteQuery := &dml.DeleteQuery[userField]{
			BaseQuery: dml.BaseQuery[userField]{
				Ta: "users",
			},
		}
		deleteQuery.Where(emailCol.Eq("john@example.com"))

		sql, args := deleteQuery.Build()
		t.Logf("SQL: %s", sql)
		t.Logf("Args: %v", args)

		tag, err := pool.Exec(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}

		if tag.RowsAffected() != 1 {
			t.Fatalf("expected 1 row affected, got %d", tag.RowsAffected())
		}
	})
}
