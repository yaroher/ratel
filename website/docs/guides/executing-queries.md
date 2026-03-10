---
sidebar_position: 7
title: Executing Queries
---

# Executing Queries

## DB Interface

Ratel works with any type implementing the `DB` interface:

```go
type DB interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

This is satisfied by `pgx.Conn`, `*pgxpool.Pool`, and `pgx.Tx`.

## Query Multiple Rows

```go
users, err := Users.Query(ctx, db,
    Users.SelectAll().
        Where(Users.IsActive.Eq(true)).
        Limit(10),
)
// users is []*UsersScanner
```

## Query Single Row

```go
user, err := Users.QueryRow(ctx, db,
    Users.SelectAll().
        Where(Users.ID.Eq(1)),
)
// user is *UsersScanner
```

## Execute (No Result)

For DELETE and UPDATE without RETURNING:

```go
tag, err := db.Exec(ctx,
    Users.Delete().
        Where(Users.ID.Eq(1)).
        Build(),
)
fmt.Println(tag.RowsAffected())
```

## INSERT with RETURNING

```go
inserted, err := Users.QueryRow(ctx, db,
    Users.Insert().
        Columns(Users.Email, Users.Name).
        Values("john@example.com", "John").
        Returning(Users.ID, Users.CreatedAt),
)
fmt.Println(inserted.ID, inserted.CreatedAt)
```

## Scanner

Every table has a Scanner struct that maps database rows to Go fields. The scanner implements `GetTarget` which returns a pointer to each field for row scanning:

```go
type UsersScanner struct {
    ID        int64
    Email     string
    Name      string
    CreatedAt time.Time
    // Relation fields
    Orders    []*OrdersScanner
    Profile   *ProfilesScanner
}

func (u *UsersScanner) GetTarget(col string) func() any {
    switch UsersColumnAlias(col) {
    case UsersColID:
        return func() any { return &u.ID }
    case UsersColEmail:
        return func() any { return &u.Email }
    // ...
    }
}
```

When using proto-first approach, scanners are generated automatically.

## Relations Loading

Relations defined on the scanner are loaded automatically after the main query:

```go
users, err := Users.Query(ctx, db,
    Users.SelectAll().Where(Users.IsActive.Eq(true)),
)
// users[0].Orders — loaded via OneToMany
// users[0].Profile — loaded via HasOne
```

The executor calls the scanner's `Relations()` method to get configured loaders, then runs additional queries to populate related records.
