package ddl

import (
	"testing"

	"github.com/yaroher/ratel/ddl/constraint"
)

func TestCreateTableBuild(t *testing.T) {
	sql, args := CreateTableStmt("users").
		IfNotExists().
		Columns(
			Column("id", "SERIAL").PrimaryKey(),
			Column("name", "TEXT").NotNull(),
			Column("email", "TEXT").Unique(),
			Column("age", "INTEGER").Default("0"),
		).
		Constraint(constraint.NamedUnique("uq_users_name_email", "name", "email")).
		Build()

	if len(args) != 0 {
		t.Fatalf("expected no args, got %d", len(args))
	}

	expected := "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT NOT NULL, email TEXT UNIQUE, age INTEGER DEFAULT 0, CONSTRAINT uq_users_name_email UNIQUE (name, email));"
	if sql != expected {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestAlterTableBuild(t *testing.T) {
	sql, args := AlterTableStmt("posts").
		AddColumn(Column("title", "TEXT").NotNull()).
		AddColumnIfNotExists(Column("subtitle", "TEXT")).
		DropColumn("old_title", true).
		RenameColumn("content", "body").
		SetColumnType("user_id", "BIGINT").
		Build()

	if len(args) != 0 {
		t.Fatalf("expected no args, got %d", len(args))
	}

	expected := "ALTER TABLE posts ADD COLUMN title TEXT NOT NULL, ADD COLUMN IF NOT EXISTS subtitle TEXT, DROP COLUMN IF EXISTS old_title, RENAME COLUMN content TO body, ALTER COLUMN user_id TYPE BIGINT;"
	if sql != expected {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestDropTableBuild(t *testing.T) {
	sql, args := DropTableStmt("a", "b").
		IfExists().
		Cascade().
		Build()

	if len(args) != 0 {
		t.Fatalf("expected no args, got %d", len(args))
	}

	expected := "DROP TABLE IF EXISTS a, b CASCADE;"
	if sql != expected {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestCreateIndexBuild(t *testing.T) {
	sql, args := CreateIndexStmt("idx_users_email", "users").
		Unique().
		IfNotExists().
		Concurrently().
		Using("btree").
		Columns("email").
		Where("email IS NOT NULL").
		Build()

	if len(args) != 0 {
		t.Fatalf("expected no args, got %d", len(args))
	}

	expected := "CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users USING btree (email) WHERE email IS NOT NULL;"
	if sql != expected {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestDropIndexBuild(t *testing.T) {
	sql, args := DropIndexStmt("idx_users_email").
		IfExists().
		Concurrently().
		Restrict().
		Build()

	if len(args) != 0 {
		t.Fatalf("expected no args, got %d", len(args))
	}

	expected := "DROP INDEX CONCURRENTLY IF EXISTS idx_users_email RESTRICT;"
	if sql != expected {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}
