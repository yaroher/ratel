---
sidebar_position: 2
title: pkg/ddl
---

# pkg/ddl

The DDL package generates SQL schema statements (CREATE TABLE, CREATE INDEX, etc.).

## TableDDL

```go
import "github.com/yaroher/ratel/pkg/ddl"

table := ddl.NewTableDDL[UsersAlias, UsersColumnAlias](
    UsersAliasName,
    columns,
    ddl.WithSchema[UsersAlias, UsersColumnAlias]("store"),
    ddl.WithIndexes(indexes...),
    ddl.WithPostStatements("ALTER TABLE {table} ENABLE RLS"),
)

// Generate SQL
stmts := table.SchemaSql() // []string
```

### Table Options

| Function | Description |
|----------|-------------|
| `WithSchema(name)` | Set PostgreSQL schema |
| `WithIndexes(idx...)` | Add indexes |
| `WithUniqueColumns(cols...)` | Composite UNIQUE |
| `WithPrimaryKeyColumns(cols...)` | Composite PRIMARY KEY |
| `WithTableCheckConstraint(name, expr)` | Named CHECK |
| `WithPostStatements(sql...)` | Raw SQL after CREATE TABLE |

## ColumnDDL

```go
col := ddl.NewColumnDDL[UsersColumnAlias](
    UsersColEmail,
    ddl.TEXT,
    ddl.WithNotNull[UsersColumnAlias](),
    ddl.WithUnique[UsersColumnAlias](),
)
```

### Column Options

| Function | Description |
|----------|-------------|
| `WithNotNull()` | NOT NULL |
| `WithNullable()` | Explicit NULL |
| `WithUnique()` | UNIQUE |
| `WithPrimaryKey()` | PRIMARY KEY |
| `WithDefault(val)` | DEFAULT expression |
| `WithCheck(expr)` | CHECK constraint |
| `WithReferences(table, col)` | REFERENCES |
| `WithOnDelete(action)` | ON DELETE |
| `WithOnUpdate(action)` | ON UPDATE |

## Data Types

| Constant | SQL |
|----------|-----|
| `ddl.SMALLINT` | SMALLINT |
| `ddl.INTEGER` | INTEGER |
| `ddl.BIGINT` | BIGINT |
| `ddl.SERIAL` | SERIAL |
| `ddl.SMALLSERIAL` | SMALLSERIAL |
| `ddl.BIGSERIAL` | BIGSERIAL |
| `ddl.REAL` | REAL |
| `ddl.DOUBLE` | DOUBLE PRECISION |
| `ddl.TEXT` | TEXT |
| `ddl.BOOLEAN` | BOOLEAN |
| `ddl.DATE` | DATE |
| `ddl.TIME` | TIME |
| `ddl.TIMESTAMP` | TIMESTAMP |
| `ddl.TIMESTAMPTZ` | TIMESTAMPTZ |
| `ddl.INTERVAL` | INTERVAL |
| `ddl.UUID` | UUID |
| `ddl.JSON` | JSON |
| `ddl.JSONB` | JSONB |
| `ddl.BYTEA` | BYTEA |

Functions: `ddl.Varchar(n)`, `ddl.Char(n)`, `ddl.Numeric(p, s)`, `ddl.Array(type)`

## Index

```go
idx := ddl.NewIndex[UsersAlias, UsersColumnAlias](
    "idx_users_email",
    UsersAliasName,
).OnColumns(UsersColEmail).
  Unique().
  Where("is_deleted = false").
  Using("btree")
```

## RawSQL

Wraps an arbitrary SQL statement for use with `SchemaSQL`. Use this for extensions, functions, grants, or any SQL not tied to a table definition:

```go
ddl.SchemaSQL(
    ddl.RawSQL("CREATE EXTENSION IF NOT EXISTS pg_trgm"),
    ddl.RawSQL("CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger ..."),
    models.Users,
    models.Orders,
)
```

`RawSQL` implements `SchemaSqler`, so it works anywhere tables do.

## SchemaSortedStatements

Generate all CREATE statements in dependency order:

```go
stmts, err := ddl.SchemaSortedStatements(
    Users, Orders, Products, Categories,
)
// Returns CREATE SCHEMA + CREATE TABLE in topological order
```
