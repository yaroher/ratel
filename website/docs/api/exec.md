---
sidebar_position: 5
title: pkg/exec
---

# pkg/exec

The exec package handles query execution and row scanning.

## DB Interface

```go
type DB interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

Satisfied by `pgx.Conn`, `*pgxpool.Pool`, and `pgx.Tx`.

## Scanner Interface

Your scanner struct must implement:

```go
type Scanner[C ColumnAlias] interface {
    GetTarget(col string) func() any
}
```

`GetTarget` returns a function that provides a pointer to the struct field for scanning. The executor calls this for each column in the result set.

## Query Methods

These methods are available on every `Table`:

### Query (multiple rows)

```go
scanners, err := Users.Query(ctx, db, selectQuery)
// Returns []*UsersScanner
```

### QueryRow (single row)

```go
scanner, err := Users.QueryRow(ctx, db, selectQuery)
// Returns *UsersScanner
```

### With Relations

```go
scanners, err := Users.Query(ctx, db, query,
    exec.WithRelations(relationLoader1, relationLoader2),
)
```

## Execution Flow

1. Build SQL from query
2. Execute via `db.Query()` or `db.QueryRow()`
3. For each row, create a new Scanner via the factory function
4. Call `GetTarget(col)` for each column to get scan destinations
5. Scan row into the scanner
6. After all rows, load relations (if any)
7. Return populated scanners
