---
sidebar_position: 3
title: Quick Start (SQL)
---

# Quick Start — Go Models

Define your models directly in Go without protobuf. Full control over table definitions, scanners, and query builders.

## 1. Define a Table

```go title="models/users.go"
package models

import (
    "time"
    "github.com/yaroher/ratel/pkg/ddl"
    "github.com/yaroher/ratel/pkg/schema"
)

// Table and column aliases
type UsersAlias string
func (u UsersAlias) String() string { return string(u) }
const UsersAliasName UsersAlias = "users"

type UsersColumnAlias string
func (u UsersColumnAlias) String() string { return string(u) }

const (
    UsersColID        UsersColumnAlias = "id"
    UsersColEmail     UsersColumnAlias = "email"
    UsersColName      UsersColumnAlias = "name"
    UsersColCreatedAt UsersColumnAlias = "created_at"
)

// Scanner — maps rows to Go struct
type UsersScanner struct {
    ID        int64
    Email     string
    Name      string
    CreatedAt time.Time
}

func (u *UsersScanner) GetTarget(col string) func() any {
    switch UsersColumnAlias(col) {
    case UsersColID:
        return func() any { return &u.ID }
    case UsersColEmail:
        return func() any { return &u.Email }
    case UsersColName:
        return func() any { return &u.Name }
    case UsersColCreatedAt:
        return func() any { return &u.CreatedAt }
    default:
        panic("unknown column: " + col)
    }
}

// Table definition
type UsersTable struct {
    *schema.Table[UsersAlias, UsersColumnAlias, *UsersScanner]
    ID        schema.BigSerialColumnI[UsersColumnAlias]
    Email     schema.TextColumnI[UsersColumnAlias]
    Name      schema.TextColumnI[UsersColumnAlias]
    CreatedAt schema.TimestamptzColumnI[UsersColumnAlias]
}

var Users = func() UsersTable {
    idCol := schema.BigSerialColumn(UsersColID,
        ddl.WithPrimaryKey[UsersColumnAlias](),
    )
    emailCol := schema.TextColumn(UsersColEmail,
        ddl.WithNotNull[UsersColumnAlias](),
        ddl.WithUnique[UsersColumnAlias](),
    )
    nameCol := schema.TextColumn(UsersColName,
        ddl.WithNotNull[UsersColumnAlias](),
    )
    createdAtCol := schema.TimestamptzColumn(UsersColCreatedAt,
        ddl.WithNotNull[UsersColumnAlias](),
        ddl.WithDefault[UsersColumnAlias]("now()"),
    )

    return UsersTable{
        Table: schema.NewTable[UsersAlias, UsersColumnAlias, *UsersScanner](
            UsersAliasName,
            func() *UsersScanner { return &UsersScanner{} },
            []*ddl.ColumnDDL[UsersColumnAlias]{
                idCol.DDL(), emailCol.DDL(),
                nameCol.DDL(), createdAtCol.DDL(),
            },
        ),
        ID: idCol, Email: emailCol,
        Name: nameCol, CreatedAt: createdAtCol,
    }
}()
```

## 2. Generate Schema SQL

```bash
ratel schema -p myapp/models --discover -o schema.sql
```

Or programmatically:

```go
stmts, _ := ddl.SchemaSortedStatements(models.Users)
for _, stmt := range stmts {
    fmt.Println(stmt)
}
```

## 3. Query

```go
// SELECT
users, err := models.Users.Query(ctx, pool,
    models.Users.SelectAll().
        Where(models.Users.Email.Like("%@example.com")).
        OrderByDESC(models.Users.CreatedAt).
        Limit(20),
)

// INSERT
inserted, err := models.Users.QueryRow(ctx, pool,
    models.Users.Insert().
        Columns(models.Users.Email, models.Users.Name).
        Values("alice@example.com", "Alice").
        Returning(models.Users.ID),
)

// UPDATE
_, err = models.Users.QueryRow(ctx, pool,
    models.Users.Update().
        Set(models.Users.Name.Set("Alice Smith")).
        Where(models.Users.ID.Eq(inserted.ID)).
        Returning(models.Users.ID),
)

// DELETE
_, err = pool.Exec(ctx,
    models.Users.Delete().
        Where(models.Users.ID.Eq(1)).
        Build(),
)
```

## Proto vs SQL Approach

| | Proto-First | Go Models |
|---|---|---|
| **Schema definition** | `.proto` files | Go code |
| **Code generation** | Automatic (protoc) | Manual |
| **Boilerplate** | Minimal | More verbose |
| **Flexibility** | Proto options | Full Go control |
| **Best for** | New projects | Existing schemas |

Both approaches produce the same runtime behavior and type safety.

## Next Steps

- [Defining Tables](/docs/guides/defining-tables) — advanced table options
- [Query Building](/docs/guides/query-building) — full query API
- [Repository Pattern](/docs/guides/repository-pattern) — higher-level CRUD
