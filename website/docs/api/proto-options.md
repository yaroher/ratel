---
sidebar_position: 1
title: Proto Options
---

# Proto Options Reference

Import the Ratel proto options:

```protobuf
import "ratelproto/ratelproto.proto";
```

## ratel.table (Message Option)

Applied to a protobuf message to configure table generation.

```protobuf
message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
    schema: "store"
    indexes: [...]
    unique: [...]
    primary_key: { columns: [...] }
    constraints: [...]
    post_statements: [...]
  };
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `generate` | bool | false | Enable code generation |
| `table_name` | string | snake_case(MessageName) | SQL table name |
| `schema` | string | "" (public) | PostgreSQL schema |
| `indexes` | repeated Index | [] | Table indexes |
| `unique` | repeated UniqueColumns | [] | Composite unique constraints |
| `primary_key` | PrimaryKeyColumns | - | Composite primary key |
| `constraints` | repeated string | [] | Table-level CHECK constraints |
| `post_statements` | repeated string | [] | Raw SQL after CREATE TABLE |

### Index

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Index name (auto-generated if empty) |
| `columns` | repeated string | Column names |
| `unique` | bool | UNIQUE index |
| `where` | string | Partial index WHERE clause |
| `using` | string | Index method (btree, hash, gin, gist, brin) |
| `include` | repeated string | INCLUDE columns (covering index) |
| `expressions` | repeated string | Expression-based index |
| `concurrent` | bool | CREATE INDEX CONCURRENTLY |

### UniqueColumns

| Field | Type | Description |
|-------|------|-------------|
| `columns` | repeated string | Column names for composite unique |

### PrimaryKeyColumns

| Field | Type | Description |
|-------|------|-------------|
| `columns` | repeated string | Column names for composite PK |

## ratel.column (Field Option)

Applied to a protobuf field to configure column generation.

```protobuf
int64 id = 1 [(ratel.column) = {
  constraints: { primary_key: true }
  skip: false
}];
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `constraints` | Constraint | - | Column constraints |
| `skip` | bool | false | Skip this field in DDL |

### Constraint

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `unique` | bool | false | UNIQUE constraint |
| `primary_key` | bool | false | PRIMARY KEY |
| `default_value` | string | "" | DEFAULT expression |
| `raw` | string | "" | Raw constraint SQL |
| `check` | string | "" | CHECK expression |
| `references_table` | string | "" | FK target table |
| `references_column` | string | "" | FK target column |
| `references_schema` | string | "" | FK target schema |
| `on_delete` | ReferenceAction | NO_ACTION | ON DELETE action |
| `on_update` | ReferenceAction | NO_ACTION | ON UPDATE action |

### ReferenceAction

```protobuf
enum ReferenceAction {
  NO_ACTION = 0;
  CASCADE = 1;
  SET_NULL = 2;
  SET_DEFAULT = 3;
  RESTRICT = 4;
}
```

## ratel.relation (Field Option)

Applied to a protobuf field to define a relation.

```protobuf
repeated Order orders = 10 [(ratel.relation) = {
  one_to_many: { ref_name: "user_id" }
}];
```

### OneToMany / HasOne

| Field | Type | Description |
|-------|------|-------------|
| `ref_name` | string | FK column name in child table |
| `on_delete_cascade` | bool | Add ON DELETE CASCADE |
| `on_delete_set_null` | bool | Add ON DELETE SET NULL |
| `existed_field` | bool | FK field exists in child message |

### BelongsTo

| Field | Type | Description |
|-------|------|-------------|
| `foreign_key` | string | FK column in current table |
| `owner_key` | string | PK in parent table (default: id) |

### ManyToMany

| Field | Type | Description |
|-------|------|-------------|
| `ref_on_delete_cascade` | bool | CASCADE on FK to this table |
| `back_ref_on_delete_cascade` | bool | CASCADE on FK to related table |
| `pivot_table` | PivotTable | Custom pivot table config |
