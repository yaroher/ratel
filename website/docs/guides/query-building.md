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

// Subqueries (non-correlated)
Users.ID.EqOf(subquery)     // = (SELECT ...)
Users.ID.InOf(subquery)     // IN (SELECT ...)
Users.ID.AnyOf(subquery)    // = ANY(SELECT ...)

// EXISTS (non-correlated)
Users.Table.ExistsOf(subquery)      // EXISTS (SELECT ...)
Users.Table.NotExistsOf(subquery)   // NOT EXISTS (SELECT ...)

// Raw SQL
Users.Email.EqRaw("current_user")
Users.Table.Raw("age > $1", 18)
```

### Correlated Subqueries

Use `Table.Ref(column)` to reference a column from an outer query inside a subquery.
This enables correlated `EXISTS`, `IN`, and other subquery patterns where the inner
query depends on the outer row.

```go
// Find active users who have at least one paid order:
//
// SELECT * FROM users
// WHERE users.deleted_at IS NULL
//   AND EXISTS (
//     SELECT 1 FROM orders
//     WHERE orders.user_id = users.id
//       AND orders.status = $1
//   )
subquery := Orders.Select1().Where(
    Orders.UserID.EqRef(Users.Table.Ref(UserColumnID)),  // orders.user_id = users.id
    Orders.Status.Eq("PAID"),
)

query := Users.SelectAll().Where(
    Users.DeletedAt.IsNull(),
    Users.Table.ExistsOf(subquery),
)
```

`Ref` returns a column reference that writes `table.column` directly into SQL
without a parameter placeholder. It works with any comparison method:

```go
Col.EqRef(OtherTable.Ref(col))   // col = other_table.col
Col.NeqRef(OtherTable.Ref(col))  // col != other_table.col
```
