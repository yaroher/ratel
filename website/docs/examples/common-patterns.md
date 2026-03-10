---
sidebar_position: 2
title: Common Patterns
---

# Common Patterns

## Soft Deletes

Add a nullable `deleted_at` column and filter by default:

```protobuf
message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
    indexes: [
      { columns: ["email"], unique: true, where: "deleted_at IS NULL" }
    ]
  };

  google.protobuf.Timestamp deleted_at = 5;
}
```

Query only active records:

```go
users, _ := Users.Query(ctx, db,
    Users.SelectAll().
        Where(Users.DeletedAt.IsNull()),
)
```

Soft delete:

```go
Users.QueryRow(ctx, db,
    Users.Update().
        Set(Users.DeletedAt.Set(time.Now())).
        Where(Users.ID.Eq(1)).
        Returning(Users.ID),
)
```

## Composite Primary Keys

```protobuf
message OrderItem {
  option (ratel.table) = {
    generate: true
    table_name: "order_items"
    primary_key: { columns: ["order_id", "line_no"] }
  };

  int64 order_id = 1;
  int32 line_no = 2;
}
```

## Cross-Schema Foreign Keys

Reference tables in different schemas:

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

## Row-Level Security (RLS)

```protobuf
option (ratel.table) = {
  generate: true
  table_name: "documents"
  post_statements: [
    "ALTER TABLE {table} ENABLE ROW LEVEL SECURITY",
    "CREATE POLICY tenant_isolation ON {table} USING (tenant_id = current_setting('app.tenant_id')::bigint)"
  ]
};
```

## Partial Indexes

```protobuf
indexes: [
  {
    columns: ["email"]
    unique: true
    where: "is_deleted = false"
  }
]
```

## GIN Indexes for JSONB

```protobuf
option (ratel.table) = {
  generate: true
  indexes: [
    { columns: ["metadata"], using: "gin" }
  ]
};
```

Query with JSON operators:

```go
Users.SelectAll().
    Where(Users.Metadata.JSONContains(`{"role":"admin"}`))
```

## Expression Indexes

```protobuf
indexes: [
  { expressions: ["lower(email)"], unique: true }
]
```

## Covering Indexes

```protobuf
indexes: [
  {
    columns: ["user_id"]
    include: ["created_at", "status"]
  }
]
```

## FOR UPDATE Locking

```go
tx, _ := pool.Begin(ctx)

user, _ := Users.QueryRow(ctx, tx,
    Users.SelectAll().
        Where(Users.ID.Eq(1)).
        ForUpdate(),
)
// user row is locked until transaction ends

Users.QueryRow(ctx, tx,
    Users.Update().
        Set(Users.Balance.Set(user.Balance - amount)).
        Where(Users.ID.Eq(1)).
        Returning(Users.ID),
)

tx.Commit(ctx)
```
