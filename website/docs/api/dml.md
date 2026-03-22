---
sidebar_position: 3
title: pkg/dml
---

# pkg/dml

The DML package builds type-safe SQL queries (SELECT, INSERT, UPDATE, DELETE).

## SelectQuery

```go
query := Users.SelectAll().
    Where(Users.Email.Eq("john@example.com")).
    OrderByDESC(Users.CreatedAt).
    Limit(10)
```

### Methods

| Method | Description |
|--------|-------------|
| `Fields(cols...)` | Select specific columns |
| `Distinct()` | DISTINCT |
| `Where(clause...)` | WHERE conditions |
| `GroupBy(cols...)` | GROUP BY |
| `Having(clause...)` | HAVING |
| `OrderByASC(cols...)` | ORDER BY ASC |
| `OrderByDESC(cols...)` | ORDER BY DESC |
| `OrderByRaw(sql...)` | Raw ORDER BY |
| `Limit(n)` | LIMIT |
| `Offset(n)` | OFFSET |
| `ForUpdate()` | FOR UPDATE |
| `InnerJoin(table, alias, on...)` | INNER JOIN |
| `LeftJoin(table, alias, on...)` | LEFT JOIN |
| `RightJoin(table, alias, on...)` | RIGHT JOIN |
| `FullJoin(table, alias, on...)` | FULL JOIN |

## InsertQuery

```go
query := Users.Insert().
    Columns(Users.Email, Users.Name).
    Values("john@example.com", "John").
    Returning(Users.ID)
```

## UpdateQuery

```go
query := Users.Update().
    Set(Users.Name.Set("Jane")).
    Set(Users.IsActive.Set(true)).
    Where(Users.ID.Eq(1)).
    Returning(Users.ID)
```

## DeleteQuery

```go
query := Users.Delete().
    Where(Users.ID.Eq(1))
```

## Clause Operators

### Comparison

| Operator | SQL |
|----------|-----|
| `Col.Eq(val)` | `= val` |
| `Col.Neq(val)` | `!= val` |
| `Col.Gt(val)` | `> val` |
| `Col.Gte(val)` | `>= val` |
| `Col.Lt(val)` | `< val` |
| `Col.Lte(val)` | `<= val` |

### Set Operations

| Operator | SQL |
|----------|-----|
| `Col.In(vals...)` | `IN (...)` |
| `Col.NotIn(vals...)` | `NOT IN (...)` |
| `Col.InOf(subquery)` | `IN (SELECT ...)` |
| `Col.Between(a, b)` | `BETWEEN a AND b` |
| `Col.NotBetween(a, b)` | `NOT BETWEEN` |

### Text

| Operator | SQL |
|----------|-----|
| `Col.Like(pattern)` | `LIKE` |
| `Col.ILike(pattern)` | `ILIKE` |
| `Col.NotLike(pattern)` | `NOT LIKE` |

### Null

| Operator | SQL |
|----------|-----|
| `Col.IsNull()` | `IS NULL` |
| `Col.IsNotNull()` | `IS NOT NULL` |

### Array (PostgreSQL)

| Operator | SQL |
|----------|-----|
| `Col.ARRAYContains(vals...)` | `@>` |
| `Col.ARRAYContainedBy(vals...)` | `<@` |
| `Col.ARRAYOverlap(vals...)` | `&&` |

### JSON

| Operator | SQL |
|----------|-----|
| `Col.JSONContains(json)` | `@>` |
| `Col.JSONHasKey(key)` | `?` |
| `Col.JSONHasAnyKey(keys...)` | `?|` |
| `Col.JSONHasAllKeys(keys...)` | `?&` |

### Subqueries

| Operator | SQL |
|----------|-----|
| `Col.EqOf(subquery)` | `= (SELECT ...)` |
| `Col.InOf(subquery)` | `IN (SELECT ...)` |
| `Col.AnyOf(subquery)` | `= ANY(SELECT ...)` |
| `Table.ExistsOf(subquery)` | `EXISTS (...)` |
| `Table.NotExistsOf(subquery)` | `NOT EXISTS (...)` |

### Correlated Subqueries

Use `Table.Ref(column)` to reference an outer table's column inside a subquery:

| Operator | SQL |
|----------|-----|
| `Table.Ref(col)` | Returns a column reference (`table.col`) |
| `Col.EqRef(ref)` | `col = other_table.col` |
| `Col.NeqRef(ref)` | `col != other_table.col` |

```go
// SELECT * FROM users WHERE EXISTS (
//   SELECT 1 FROM orders WHERE orders.user_id = users.id AND orders.status = $1)
subquery := Orders.Select1().Where(
    Orders.UserID.EqRef(Users.Table.Ref(UserColumnID)),
    Orders.Status.Eq("PAID"),
)
query := Users.SelectAll().Where(Users.Table.ExistsOf(subquery))
```

### Logical

| Operator | SQL |
|----------|-----|
| `Table.And(clauses...)` | `(a AND b)` |
| `Table.Or(clauses...)` | `(a OR b)` |
| `Table.Raw(sql, args...)` | Raw SQL |
