---
sidebar_position: 4
title: Relations
---

# Relations

Ratel supports four relation types, all with type-safe eager loading.

## OneToMany

Parent has many children. Example: User has many Orders.

```protobuf
message User {
  option (ratel.table) = { generate: true, table_name: "users" };
  int64 id = 1 [(ratel.column) = { constraints: { primary_key: true } }];

  repeated Order orders = 10 [(ratel.relation) = {
    one_to_many: {
      ref_name: "user_id"
      on_delete_cascade: true
    }
  }];
}
```

| Option | Description |
|--------|-------------|
| `ref_name` | FK column name in the child table |
| `on_delete_cascade` | Add ON DELETE CASCADE to the FK |
| `existed_field` | FK field already exists in child message |

## HasOne

Parent has exactly one child. Example: User has one Profile.

```protobuf
message User {
  Profile profile = 11 [(ratel.relation) = {
    has_one: {
      ref_name: "user_id"
      on_delete_cascade: true
    }
  }];
}
```

Same options as OneToMany, but loads a single record instead of a slice.

## BelongsTo

Child references its parent. Example: Order belongs to User.

```protobuf
message Order {
  option (ratel.table) = { generate: true, table_name: "orders" };
  int64 user_id = 2;

  User user = 10 [(ratel.relation) = {
    belongs_to: {
      foreign_key: "user_id"
      owner_key: "id"
    }
  }];
}
```

| Option | Description |
|--------|-------------|
| `foreign_key` | FK column in the current table |
| `owner_key` | PK column in the parent table (default: `id`) |

## ManyToMany

Both sides have many. Ratel auto-generates a pivot table.

```protobuf
message Product {
  option (ratel.table) = { generate: true, table_name: "products" };

  repeated Category categories = 10 [(ratel.relation) = {
    many_to_many: {
      ref_on_delete_cascade: true
      back_ref_on_delete_cascade: true
    }
  }];
}
```

| Option | Description |
|--------|-------------|
| `ref_on_delete_cascade` | CASCADE on the FK to this table |
| `back_ref_on_delete_cascade` | CASCADE on the FK to the related table |
| `pivot_table.table_name` | Custom pivot table name |

## Loading Relations

Relations are loaded eagerly when querying:

```go
users, err := Users.Query(ctx, db,
    Users.SelectAll().Where(Users.IsActive.Eq(true)),
)
// users[0].Orders is populated
// users[0].Profile is populated
```

Each scanner struct has a `Relations()` method that returns configured relation loaders. The executor uses these to automatically load related records after the main query.

## Relation Direction Summary

| Relation | FK Location | Result Type |
|----------|-------------|-------------|
| OneToMany | Child table | `[]*ChildScanner` |
| HasOne | Child table | `*ChildScanner` |
| BelongsTo | Current table | `*ParentScanner` |
| ManyToMany | Pivot table | `[]*RelatedScanner` |
