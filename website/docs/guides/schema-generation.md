---
sidebar_position: 5
title: Schema Generation
---

# Schema Generation

Ratel generates DDL (CREATE TABLE, CREATE INDEX, etc.) from your table definitions.

## Generate SQL Programmatically

```go
import "github.com/yaroher/ratel/pkg/ddl"

// Single table
stmts := Users.Table.SchemaSql()
for _, stmt := range stmts {
    fmt.Println(stmt)
}

// Multiple tables — topologically sorted
stmts, err := ddl.SchemaSortedStatements(
    Users, Orders, Products, Categories,
)
```

`SchemaSortedStatements` analyzes foreign key dependencies and outputs tables in the correct order. It also generates `CREATE SCHEMA` statements when needed.

## Generate SQL via CLI

```bash
ratel schema -p github.com/myapp/models --discover -o schema.sql
```

Or specify tables explicitly:

```bash
ratel schema -p github.com/myapp/models -t Users,Orders,Products -o schema.sql
```

## Output Example

```sql
CREATE SCHEMA IF NOT EXISTS "store";

CREATE TABLE "store"."users" (
    "id" BIGSERIAL PRIMARY KEY,
    "email" TEXT NOT NULL UNIQUE,
    "name" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX "idx_users_email" ON "store"."users" ("email");

CREATE TABLE "store"."orders" (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "store"."users"("id") ON DELETE CASCADE,
    "status" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);

CHECK (status IN ('NEW', 'PAID', 'CANCELLED'));
```

## Topological Sorting

When tables have foreign key dependencies, `SchemaSortedStatements` ensures:

1. `CREATE SCHEMA` statements come first
2. Referenced tables are created before referencing tables
3. Circular dependencies are detected and reported as errors

## Schema Qualification

Tables with a `schema` option generate fully qualified names:

```protobuf
option (ratel.table) = {
  generate: true
  table_name: "users"
  schema: "store"
};
```

Generates: `CREATE TABLE "store"."users" (...)`

Cross-schema FK references use the `references_schema` field:

```protobuf
constraints: {
  references_schema: "auth"
  references_table: "users"
}
```

Generates: `REFERENCES "auth"."users"("id")`
