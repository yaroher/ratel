---
sidebar_position: 2
title: Quick Start (Proto)
---

# Quick Start — Proto-First

Define your schema in `.proto` files. Ratel generates type-safe Go code with DDL, DML, scanners, and relation loaders.

## 1. Create a Proto File

```protobuf title="schema/store.proto"
syntax = "proto3";
package store;
option go_package = "myapp/storepb";

import "google/protobuf/timestamp.proto";
import "ratelproto/ratelproto.proto";

message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
    indexes: [
      { columns: ["email"], unique: true }
    ]
  };

  int64 id = 1 [(ratel.column) = {
    constraints: { primary_key: true }
  }];

  string email = 2 [(ratel.column) = {
    constraints: { unique: true }
  }];

  string name = 3;

  google.protobuf.Timestamp created_at = 4 [(ratel.column) = {
    constraints: { default_value: "now()" }
  }];
}
```

## 2. Generate Go Code

```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --plain_out=. --plain_opt=paths=source_relative \
  --ratel_out=. --ratel_opt=paths=source_relative \
  schema/store.proto
```

This generates:
- `store.pb.go` — standard protobuf code
- `store.plain.go` — plain Go structs (no protobuf runtime)
- `store_ratel.pb.go` — table definitions, scanners, DDL, DML

## 3. Create the Schema

```go
package main

import (
    "context"
    "fmt"
    "myapp/storepb"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/yaroher/ratel/pkg/ddl"
)

func main() {
    ctx := context.Background()
    pool, _ := pgxpool.New(ctx, "postgres://localhost:5432/mydb")
    defer pool.Close()

    // Generate and execute CREATE TABLE statements
    stmts, _ := ddl.SchemaSortedStatements(storepb.Users)
    for _, stmt := range stmts {
        pool.Exec(ctx, stmt)
    }
}
```

## 4. Query with Type Safety

```go
// SELECT — compiler checks column names and types
users, err := storepb.Users.Query(ctx, pool,
    storepb.Users.SelectAll().
        Where(storepb.Users.Email.Eq("john@example.com")).
        OrderByDESC(storepb.Users.CreatedAt).
        Limit(10),
)

for _, u := range users {
    fmt.Println(u.Email, u.Name)
}

// INSERT with RETURNING
inserted, err := storepb.Users.QueryRow(ctx, pool,
    storepb.Users.Insert().
        Columns(storepb.Users.Email, storepb.Users.Name).
        Values("jane@example.com", "Jane Doe").
        Returning(storepb.Users.Id),
)
fmt.Println("New ID:", inserted.Id)

// UPDATE
updated, err := storepb.Users.QueryRow(ctx, pool,
    storepb.Users.Update().
        Set(storepb.Users.Name.Set("Jane Smith")).
        Where(storepb.Users.Id.Eq(1)).
        Returning(storepb.Users.Id),
)
```

Every column name and type is checked at compile time. Typos and type mismatches are caught before your code runs.

## Next Steps

- [Defining Tables](/docs/guides/defining-tables) — table options, embedded messages, schemas
- [Column Types](/docs/guides/column-types) — full type mapping reference
- [Relations](/docs/guides/relations) — HasMany, BelongsTo, ManyToMany
