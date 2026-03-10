---
sidebar_position: 3
title: Constraints
---

# Constraints

## Primary Key

### Single Column

```protobuf
int64 id = 1 [(ratel.column) = {
  constraints: { primary_key: true }
}];
```

### Composite

```protobuf
option (ratel.table) = {
  generate: true
  primary_key: { columns: ["order_id", "line_no"] }
};
```

## Unique

### Single Column

```protobuf
string email = 2 [(ratel.column) = {
  constraints: { unique: true }
}];
```

### Composite

```protobuf
option (ratel.table) = {
  generate: true
  unique: [
    { columns: ["order_id", "product_id"] }
  ]
};
```

## Default Values

```protobuf
google.protobuf.Timestamp created_at = 4 [(ratel.column) = {
  constraints: { default_value: "now()" }
}];

bool is_active = 5 [(ratel.column) = {
  constraints: { default_value: "true" }
}];
```

## CHECK Constraints

### Column-Level

```protobuf
string code = 1 [(ratel.column) = {
  constraints: { check: "code <> ''" }
}];
```

### Table-Level

```protobuf
option (ratel.table) = {
  generate: true
  constraints: [
    "CHECK (status IN ('NEW', 'PAID', 'CANCELLED'))"
  ]
};
```

## Foreign Key References

Reference another table's column:

```protobuf
int64 user_id = 2 [(ratel.column) = {
  constraints: {
    references_table: "users"
    references_column: "id"
    on_delete: CASCADE
  }
}];
```

### Cross-Schema References

Use `references_schema` for tables in a different schema:

```protobuf
int64 user_id = 2 [(ratel.column) = {
  constraints: {
    references_schema: "auth"
    references_table: "users"
    references_column: "id"
    on_delete: CASCADE
  }
}];
```

Generates: `REFERENCES "auth"."users"(id) ON DELETE CASCADE`

### Reference Actions

Available actions for `on_delete` and `on_update`:

| Action | Description |
|--------|-------------|
| `NO_ACTION` | Default — raise error if referenced row exists |
| `CASCADE` | Delete/update dependent rows |
| `SET_NULL` | Set FK column to NULL |
| `SET_DEFAULT` | Set FK column to its default value |
| `RESTRICT` | Like NO_ACTION but not deferrable |

```protobuf
int64 order_id = 1 [(ratel.column) = {
  constraints: {
    references_schema: "store"
    references_table: "orders"
    references_column: "id"
    on_delete: CASCADE
    on_update: NO_ACTION
  }
}];
```

## Raw Constraints

For anything not covered by structured options:

```protobuf
string code = 1 [(ratel.column) = {
  constraints: {
    raw: "COLLATE \"C\" NOT NULL"
  }
}];
```

## Go Approach

```go
col := schema.BigIntColumn(UsersColID,
    ddl.WithPrimaryKey[UsersColumnAlias](),
    ddl.WithNotNull[UsersColumnAlias](),
)

fkCol := schema.BigIntColumn(OrdersColUserID,
    ddl.WithNotNull[UsersColumnAlias](),
    ddl.WithReferences[UsersColumnAlias]("users", "id"),
    ddl.WithOnDelete[UsersColumnAlias]("CASCADE"),
)
```
