<p align="center">
  <img src="website/static/img/ratel-banner.svg" alt="Ratel" width="500"/>
</p>

<p align="center">
  <strong>Fearless, type-safe PostgreSQL ORM for Go</strong>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/github/go-mod/go-version/yaroher/ratel" alt="Go Version"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"/></a>
  <a href="https://github.com/yaroher/ratel/actions/workflows/ci.yml"><img src="https://github.com/yaroher/ratel/actions/workflows/ci.yml/badge.svg" alt="CI"/></a>
</p>

---

**Ratel** is a type-safe PostgreSQL ORM and code generator for Go. Compile-time checked queries, zero reflection, maximum performance.

## Features

- **Type-safe queries** — Column names, types, and relations are checked at compile time
- **Code generation** — Generate models from SQL schema or protobuf definitions
- **Migration support** — Built-in diff tool powered by [Atlas](https://atlasgo.io/)
- **Relations** — HasMany, BelongsTo, ManyToMany with eager loading
- **pgx integration** — Built on top of [pgx](https://github.com/jackc/pgx) for optimal PostgreSQL support
- **Protobuf support** — Generate database models from protobuf definitions with `protoc-gen-ratel`

## Installation

```bash
go install github.com/yaroher/ratel/cmd/ratel@latest
go install github.com/yaroher/ratel/cmd/protoc-gen-ratel@latest
```

## Quick Start

### 1. Define your models

```go
package models

import (
    "github.com/yaroher/ratel/pkg/ddl"
    "github.com/yaroher/ratel/pkg/schema"
)

type UsersAlias string
func (u UsersAlias) String() string { return string(u) }
const UsersAliasName UsersAlias = "users"

type UsersColumnAlias string
func (u UsersColumnAlias) String() string { return string(u) }

const (
    UsersColumnUserID   UsersColumnAlias = "user_id"
    UsersColumnEmail    UsersColumnAlias = "email"
    UsersColumnFullName UsersColumnAlias = "full_name"
)

type UsersScanner struct {
    UserID   int64
    Email    string
    FullName string
}

// ... implement GetTarget, GetSetter, GetValue methods

var Users = func() UsersTable {
    userIDCol := schema.BigSerialColumn(UsersColumnUserID, ddl.WithPrimaryKey[UsersColumnAlias]())
    emailCol := schema.TextColumn(UsersColumnEmail,
        ddl.WithNotNull[UsersColumnAlias](),
        ddl.WithUnique[UsersColumnAlias](),
    )
    fullNameCol := schema.TextColumn(UsersColumnFullName, ddl.WithNotNull[UsersColumnAlias]())

    return UsersTable{
        Table: schema.NewTable[UsersAlias, UsersColumnAlias, *UsersScanner](
            UsersAliasName,
            func() *UsersScanner { return &UsersScanner{} },
            []*ddl.ColumnDDL[UsersColumnAlias]{
                userIDCol.DDL(),
                emailCol.DDL(),
                fullNameCol.DDL(),
            },
        ),
        UserID:   userIDCol,
        Email:    emailCol,
        FullName: fullNameCol,
    }
}()
```

### 2. Generate SQL schema

```bash
# Generate schema from Go models
ratel schema -p github.com/yourproject/models --discover -o schema.sql
```

### 3. Create migrations

```bash
# Generate migration diff from Go models
ratel diff -p github.com/yourproject/models --discover -d migrations -n add_users

# Or from SQL file
ratel diff -s schema.sql -d migrations -n add_users
```

### 4. Query your data

```go
import (
    "github.com/yaroher/ratel/pkg/dml"
    "github.com/yaroher/ratel/pkg/exec"
)

// Select with type-safe columns
query := dml.Select(Users.Table).
    Columns(Users.UserID, Users.Email, Users.FullName).
    Where(Users.Email.Eq("user@example.com"))

// Execute query
users, err := exec.Query(ctx, db, query)

// Insert
insert := dml.Insert(Users.Table).
    Set(Users.Email.Set("new@example.com")).
    Set(Users.FullName.Set("John Doe")).
    Returning(Users.UserID)

// Update
update := dml.Update(Users.Table).
    Set(Users.FullName.Set("Jane Doe")).
    Where(Users.UserID.Eq(1))

// Delete
del := dml.Delete(Users.Table).
    Where(Users.UserID.Eq(1))
```

## CLI Commands

### `ratel schema`

Generate SQL schema from Go models package.

```bash
ratel schema -p <package> [flags]

Flags:
  -p, --package string   Go package path containing models (required)
  -t, --tables strings   Table variable names (e.g., Users,Products)
  -d, --discover         Auto-discover tables from source files
  -o, --output string    Output SQL schema file (default "schema.sql")
```

### `ratel diff`

Generate SQL migration diff between current schema and models.

```bash
ratel diff [flags]

Flags:
  -s, --sql string       SQL schema file to compare against
  -p, --package string   Go package path containing models
  -t, --tables strings   Table variable names (e.g., Users,Products)
      --discover         Auto-discover tables from source files
  -d, --dir string       Migration directory for output (default "./migrations")
  -n, --name string      Migration name (default "migration")
  -v, --pg_version int   PostgreSQL version (default 18)
```

### `ratel generate`

Generate Go models from SQL schema.

```bash
ratel generate -i <input.sql> -o <output_dir> -p <package_name>
```

## Protobuf Integration

Ratel supports generating database models from protobuf definitions using `protoc-gen-ratel`.

```bash
protoc \
    --plugin=protoc-gen-ratel=./bin/protoc-gen-ratel \
    --go_out=. \
    --go_opt=paths=source_relative \
    --ratel_out=. \
    --ratel_opt=paths=source_relative \
    your_schema.proto
```

## Project Structure

```
ratel/
├── cmd/
│   ├── ratel/              # CLI tool
│   └── protoc-gen-ratel/   # Protobuf plugin
├── pkg/
│   ├── ddl/                # DDL (CREATE TABLE, etc.)
│   ├── dml/                # DML (SELECT, INSERT, UPDATE, DELETE)
│   ├── exec/               # Query execution
│   ├── schema/             # Table and column definitions
│   └── pgx-ext/            # pgx extensions
├── examples/
│   ├── store/              # Hand-written models example
│   └── proto/              # Protobuf-based example
└── ratelproto/             # Ratel protobuf options
```

## Examples

See the [examples](./examples) directory for complete working examples:

- **[store](./examples/store)** — E-commerce schema with users, orders, products, categories, and tags
- **[proto](./examples/proto)** — Same schema defined using protobuf

## Requirements

- Go 1.21+
- PostgreSQL 14+ (for migrations with testcontainers)
- Docker (for running migration tests)

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
