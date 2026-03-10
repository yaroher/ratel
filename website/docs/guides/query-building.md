---
sidebar_position: 6
title: Query Building
---

# Query Building

All queries are built using type-safe column references. The compiler catches typos and type mismatches.

## SELECT

```go
// Select all columns
query := Users.SelectAll()

// Select specific columns
query := Users.Select(Users.ID, Users.Email)

// With conditions
query := Users.SelectAll().
    Where(Users.Email.Eq("john@example.com")).
    Where(Users.IsActive.Eq(true)).
    OrderByDESC(Users.CreatedAt).
    Limit(10).
    Offset(20)
```

### Available Methods

| Method | SQL |
|--------|-----|
| `.Fields(cols...)` | SELECT specific columns |
| `.Distinct()` | SELECT DISTINCT |
| `.Where(clause...)` | WHERE |
| `.GroupBy(cols...)` | GROUP BY |
| `.Having(clause...)` | HAVING |
| `.OrderByASC(cols...)` | ORDER BY ... ASC |
| `.OrderByDESC(cols...)` | ORDER BY ... DESC |
| `.OrderByRaw(sql...)` | ORDER BY (raw) |
| `.Limit(n)` | LIMIT |
| `.Offset(n)` | OFFSET |
| `.ForUpdate()` | FOR UPDATE |

### Joins

```go
query := Users.SelectAll().
    InnerJoin(Orders.Table, "o",
        Users.ID.EqCol(Orders.UserID),
    ).
    LeftJoin(Profiles.Table, "p",
        Users.ID.EqCol(Profiles.UserID),
    )
```

## INSERT

```go
// Single row
query := Users.Insert().
    Columns(Users.Email, Users.Name).
    Values("john@example.com", "John Doe").
    Returning(Users.ID)

// Multiple rows
query := Users.Insert().
    Columns(Users.Email, Users.Name).
    Values("john@example.com", "John").
    Values("jane@example.com", "Jane").
    Returning(Users.ID)
```

## UPDATE

```go
query := Users.Update().
    Set(Users.Name.Set("Jane Smith")).
    Set(Users.IsActive.Set(true)).
    Where(Users.ID.Eq(1)).
    Returning(Users.ID, Users.Name)
```

## DELETE

```go
query := Users.Delete().
    Where(Users.IsActive.Eq(false))
```

## Where Clause Operators

### Comparison

```go
Users.ID.Eq(1)         // = 1
Users.ID.Neq(1)        // != 1
Users.ID.Gt(1)         // > 1
Users.ID.Gte(1)        // >= 1
Users.ID.Lt(10)        // < 10
Users.ID.Lte(10)       // <= 10
```

### IN / NOT IN

```go
Users.ID.In(1, 2, 3)       // IN (1, 2, 3)
Users.ID.NotIn(1, 2, 3)    // NOT IN (1, 2, 3)
Users.ID.InOf(subquery)     // IN (SELECT ...)
```

### BETWEEN

```go
Users.ID.Between(1, 100)      // BETWEEN 1 AND 100
Users.ID.NotBetween(1, 100)   // NOT BETWEEN 1 AND 100
```

### LIKE / ILIKE

```go
Users.Email.Like("%@example.com")    // LIKE
Users.Email.ILike("john%")          // ILIKE (case-insensitive)
Users.Email.NotLike("%@spam.com")   // NOT LIKE
```

### NULL Checks

```go
Users.Bio.IsNull()       // IS NULL
Users.Bio.IsNotNull()    // IS NOT NULL
```

### Array Operators (PostgreSQL)

```go
Users.Tags.ARRAYContains("go", "rust")     // @> ARRAY[...]
Users.Tags.ARRAYContainedBy("go")          // <@ ARRAY[...]
Users.Tags.ARRAYOverlap("go", "rust")      // && ARRAY[...]
```

### JSON Operators

```go
Users.Data.JSONContains(`{"role":"admin"}`)   // @> '...'
Users.Data.JSONHasKey("role")                 // ? 'role'
Users.Data.JSONHasAnyKey("role", "name")      // ?| ARRAY[...]
Users.Data.JSONHasAllKeys("role", "name")     // ?& ARRAY[...]
```

### Logical Operators

```go
// AND (multiple Where calls are AND'd)
Users.SelectAll().
    Where(Users.IsActive.Eq(true)).
    Where(Users.Email.Like("%@example.com"))

// OR
Users.SelectAll().
    Where(Users.Or(
        Users.Email.Eq("admin@example.com"),
        Users.IsActive.Eq(true),
    ))

// Subqueries
Users.ID.EqOf(subquery)     // = (SELECT ...)
Users.ID.AnyOf(subquery)    // = ANY(SELECT ...)

// EXISTS
Users.ExistsOf(subquery)      // EXISTS (SELECT ...)
Users.NotExistsOf(subquery)   // NOT EXISTS (SELECT ...)

// Raw SQL
Users.Email.EqRaw("current_user")
Users.Raw("age > $1", 18)
```
