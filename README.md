<p align="center">
  <img src="website/static/img/ratel.png" alt="Ratel" width="180"/>
</p>

<h1 align="center">RATEL</h1>

<p align="center">
  <strong>Fearless, type-safe PostgreSQL ORM for Go.</strong><br/>
  Compile-time checked queries. Zero reflection. Maximum performance.
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/github/go-mod/go-version/yaroher/ratel" alt="Go Version"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"/></a>
  <a href="https://github.com/yaroher/ratel/actions/workflows/ci.yml"><img src="https://github.com/yaroher/ratel/actions/workflows/ci.yml/badge.svg" alt="CI"/></a>
</p>

---

```bash
go install github.com/yaroher/ratel/cmd/ratel@latest
```

```go
// Type-safe queries — compiler catches typos
users, err := Users.Query(ctx, db,
    Users.SelectAll().
        Where(Users.Email.Eq("john@example.com")).
        OrderBy(Users.CreatedAt.Desc()).
        Limit(10),
)
```

## Why Ratel

**Type Safety** — Column names and types checked at compile time. No magic strings. No runtime reflection.

**Relations** — HasMany, BelongsTo, ManyToMany with eager loading. Type-safe relation options per query.

**Code Generation** — Generate from Protobuf or SQL schema. Full DDL, DML, scanners, and repositories.

**PostgreSQL Native** — Built on [pgx](https://github.com/jackc/pgx). JSONB, arrays, partial indexes, schemas, RLS — first-class support.

**Migrations** — [Atlas](https://atlasgo.io/)-powered schema diffing. Auto-generate migration files from model changes.

**Zero Overhead** — Go generics, no `interface{}` boxing. Direct struct scanning without reflection.

## Write less. Catch more.

<table>
<tr>
<th>Raw SQL + pgx</th>
<th>Ratel</th>
</tr>
<tr>
<td>

```go
rows, err := db.Query(ctx,
  "SELECT id, email, name
   FROM users
   WHERE is_active = $1", true)

// Manual scanning...
// No compile-time checks
// Typos found at runtime
```

</td>
<td>

```go
users, err := Users.Query(ctx, db,
  Users.SelectAll().
    Where(Users.IsActive.Eq(true)),
)

// Auto scanning
// Compile-time column checks
// Type-safe predicates
```

</td>
</tr>
</table>

## Quick Start

### Proto-first

Define your schema in `.proto` files, generate everything:

```protobuf
message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
    indexes: [{ columns: ["email"], unique: true }]
  };

  int64 id = 1 [(ratel.column) = { constraints: { primary_key: true } }];
  string email = 2 [(ratel.column) = { constraints: { unique: true } }];
  string name = 3;
  google.protobuf.Timestamp created_at = 4 [(ratel.column) = {
    constraints: { default_value: "now()" }
  }];
}
```

```bash
protoc --go_out=. --plain_out=. --ratel_out=. schema.proto
```

### SQL-first

Define models in Go, generate schema:

```bash
ratel schema -p github.com/myapp/models --discover -o schema.sql
ratel diff -p github.com/myapp/models --discover -d migrations -n init
```

### Query

```go
// SELECT
users, err := Users.Query(ctx, db,
    Users.SelectAll().
        Where(Users.Email.Like("%@example.com")).
        OrderByDESC(Users.CreatedAt).
        Limit(20),
)

// INSERT with RETURNING
inserted, err := Users.QueryRow(ctx, db,
    Users.Insert().
        Columns(Users.Email, Users.Name).
        Values("alice@example.com", "Alice").
        Returning(Users.ID),
)

// UPDATE
Users.QueryRow(ctx, db,
    Users.Update().
        Set(Users.Name.Set("Alice Smith")).
        Where(Users.ID.Eq(inserted.ID)).
        Returning(Users.ID),
)

// DELETE
Users.Delete().Where(Users.ID.Eq(1))
```

## CLI

| Command | Description |
|---------|-------------|
| `ratel schema` | Generate SQL from Go models |
| `ratel diff` | Generate migration diffs (Atlas) |
| `ratel generate` | Generate Go models from SQL |
| `protoc-gen-ratel` | Protoc plugin for code generation |

## Project Structure

```
ratel/
├── cmd/
│   ├── ratel/              # CLI tool
│   └── protoc-gen-ratel/   # Protobuf plugin
├── pkg/
│   ├── ddl/                # DDL (CREATE TABLE, indexes, constraints)
│   ├── dml/                # DML (SELECT, INSERT, UPDATE, DELETE)
│   ├── schema/             # Table and column definitions
│   ├── exec/               # Query execution and scanning
│   └── repository/         # Repository pattern abstractions
├── examples/
│   ├── store/              # Hand-written models example
│   └── proto/              # Protobuf-based example
├── ratelproto/             # Protobuf option definitions
└── website/                # Documentation (Docusaurus)
```

## Requirements

- Go 1.22+
- PostgreSQL 14+

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
