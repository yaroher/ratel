---
sidebar_position: 4
title: pkg/schema
---

# pkg/schema

The schema package defines typed column constructors and relation helpers.

## Table

```go
table := schema.NewTable[TableAlias, ColumnAlias, *Scanner](
    aliasName,
    func() *Scanner { return &Scanner{} },
    columns,
    tableOptions...,
)
```

A `Table` combines DDL generation, DML query building, and query execution in one type-safe struct.

## Column Constructors

Every column type is a function that returns a typed column interface:

```go
schema.BigSerialColumn[C](alias, opts...)   // BIGSERIAL
schema.BigIntColumn[C](alias, opts...)      // BIGINT
schema.IntegerColumn[C](alias, opts...)     // INTEGER
schema.TextColumn[C](alias, opts...)        // TEXT
schema.BooleanColumn[C](alias, opts...)     // BOOLEAN
schema.TimestamptzColumn[C](alias, opts...) // TIMESTAMPTZ
schema.JSONBColumn[C](alias, opts...)       // JSONB
schema.UUIDColumn[C](alias, opts...)        // UUID
// ... and all other types
```

Each column provides:
- `.DDL()` — returns `*ddl.ColumnDDL` for schema generation
- `.Eq()`, `.Gt()`, `.Like()`, etc. — clause operators for WHERE
- `.Set()` — value setter for UPDATE

Nullable variants (`NullBigIntColumn`, `NullTextColumn`, etc.) add `.IsNull()` and `.IsNotNull()` operators.

## Relation Helpers

### HasMany

```go
rel := schema.HasMany[
    ParentAlias, ParentCol, *ParentScanner,
    ChildAlias, ChildCol, *ChildScanner,
](parentAlias, childRef, childFK, parentPK)
```

### HasOne

```go
rel := schema.HasOne[
    ParentAlias, ParentCol, *ParentScanner,
    ChildAlias, ChildCol, *ChildScanner,
](parentAlias, childRef, childFK, parentPK)
```

### BelongsTo

```go
rel := schema.BelongsTo[
    ChildAlias, ChildCol, *ChildScanner,
    ParentAlias, ParentCol, *ParentScanner,
](childAlias, parentRef, childFK, parentPK)
```

### Relation Loading

```go
loader := schema.HasManyLoad(
    relation,
    childTable,
    foreignKeyCol,
    func(parent *ParentScanner, children []*ChildScanner) {
        parent.Children = children
    },
)
```
