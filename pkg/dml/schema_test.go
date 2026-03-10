package dml

import (
	"testing"
)

type testAlias string

func (a testAlias) String() string { return string(a) }

type testCol string

func (c testCol) String() string { return string(c) }

func TestSelectQuery_WithSchema(t *testing.T) {
	tbl := NewTableDML[testAlias, testCol]("users", "id", "email")
	tbl.WithSchema("store")

	query := tbl.SelectAll()
	sql, _ := query.Build()

	expected := `SELECT users.id, users.email FROM "store"."users" AS users;`
	if sql != expected {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expected, sql)
	}
}

func TestSelectQuery_WithoutSchema(t *testing.T) {
	tbl := NewTableDML[testAlias, testCol]("users", "id", "email")

	query := tbl.SelectAll()
	sql, _ := query.Build()

	expected := `SELECT users.id, users.email FROM users;`
	if sql != expected {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expected, sql)
	}
}

func TestInsertQuery_WithSchema(t *testing.T) {
	tbl := NewTableDML[testAlias, testCol]("users", "id", "email")
	tbl.WithSchema("store")

	query := tbl.Insert().Columns("email").Values("test@example.com")
	sql, args := query.Build()

	expected := `INSERT INTO "store"."users" (email) VALUES ($1);`
	if sql != expected {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expected, sql)
	}
	if len(args) != 1 || args[0] != "test@example.com" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestUpdateQuery_WithSchema(t *testing.T) {
	tbl := NewTableDML[testAlias, testCol]("users", "id", "email")
	tbl.WithSchema("store")

	query := tbl.Update()
	sql, _ := query.Build()

	// UPDATE with schema uses AS alias for WHERE clause column references
	if sql != `UPDATE "store"."users" AS users SET ;` {
		t.Errorf("unexpected SQL: %s", sql)
	}
}

func TestDeleteQuery_WithSchema(t *testing.T) {
	tbl := NewTableDML[testAlias, testCol]("users", "id", "email")
	tbl.WithSchema("store")

	query := tbl.Delete()
	sql, _ := query.Build()

	expected := `DELETE FROM "store"."users" AS users;`
	if sql != expected {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expected, sql)
	}
}

func TestDeleteQuery_WithoutSchema(t *testing.T) {
	tbl := NewTableDML[testAlias, testCol]("users", "id", "email")

	query := tbl.Delete()
	sql, _ := query.Build()

	expected := `DELETE FROM users;`
	if sql != expected {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expected, sql)
	}
}
