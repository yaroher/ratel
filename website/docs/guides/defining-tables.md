---
sidebar_position: 1
title: Defining Tables
---

# Defining Tables

## Proto Approach

Add `ratel.table` option to a protobuf message:

```protobuf
message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
    schema: "public"
  };

  int64 id = 1 [(ratel.column) = {
    constraints: { primary_key: true }
  }];
  string email = 2;
}
```

### Table Options

| Option | Type | Description |
|--------|------|-------------|
| `generate` | bool | Enable code generation for this table |
| `table_name` | string | SQL table name (default: snake_case of message name) |
| `schema` | string | PostgreSQL schema (default: public) |
| `indexes` | repeated Index | Table indexes |
| `unique` | repeated UniqueColumns | Composite unique constraints |
| `primary_key` | PrimaryKeyColumns | Composite primary key |
| `constraints` | repeated string | Table-level CHECK constraints |
| `post_statements` | repeated string | Raw SQL executed after CREATE TABLE |

### Embedded Messages (BaseEntity Pattern)

Share common fields across tables using embedded messages:

```protobuf
message BaseEntity {
  int64 id = 1 [(ratel.column) = {
    constraints: { primary_key: true }
  }];
  google.protobuf.Timestamp created_at = 2 [(ratel.column) = {
    constraints: { default_value: "now()" }
  }];
  google.protobuf.Timestamp updated_at = 3 [(ratel.column) = {
    constraints: { default_value: "now()" }
  }];
}

message User {
  option (ratel.table) = { generate: true, table_name: "users" };
  BaseEntity base = 1 [(goplain.field).embed = true];
  string email = 2;
}
```

The `embed = true` option flattens BaseEntity fields into the User table.

### Composite Primary Keys

```protobuf
message OrderItem {
  option (ratel.table) = {
    generate: true
    table_name: "order_items"
    primary_key: { columns: ["order_id", "line_no"] }
  };

  int64 order_id = 1;
  int32 line_no = 2;
  int64 product_id = 3;
}
```

### Composite Unique Constraints

```protobuf
option (ratel.table) = {
  generate: true
  table_name: "order_items"
  unique: [
    { columns: ["order_id", "product_id"] }
  ]
};
```

### Indexes

```protobuf
option (ratel.table) = {
  generate: true
  table_name: "users"
  indexes: [
    { columns: ["email"], unique: true },
    { columns: ["is_active", "created_at"], where: "is_deleted = false" },
    { columns: ["data"], using: "gin" },
    { expressions: ["lower(email)"], unique: true }
  ]
};
```

### Post Statements

Execute raw SQL after table creation — useful for RLS, grants, triggers:

```protobuf
option (ratel.table) = {
  generate: true
  table_name: "users"
  post_statements: [
    "ALTER TABLE {table} ENABLE ROW LEVEL SECURITY",
    "GRANT SELECT ON {table} TO readonly_role"
  ]
};
```

`{table}` is replaced with the fully qualified table name.

## Go Approach

Define tables programmatically:

```go
var Users = func() UsersTable {
    idCol := schema.BigSerialColumn(UsersColID,
        ddl.WithPrimaryKey[UsersColumnAlias](),
    )
    emailCol := schema.TextColumn(UsersColEmail,
        ddl.WithNotNull[UsersColumnAlias](),
        ddl.WithUnique[UsersColumnAlias](),
    )

    return UsersTable{
        Table: schema.NewTable[UsersAlias, UsersColumnAlias, *UsersScanner](
            UsersAliasName,
            func() *UsersScanner { return &UsersScanner{} },
            []*ddl.ColumnDDL[UsersColumnAlias]{
                idCol.DDL(), emailCol.DDL(),
            },
            ddl.WithSchema[UsersAlias, UsersColumnAlias]("public"),
            ddl.WithIndexes(
                ddl.NewIndex[UsersAlias, UsersColumnAlias](
                    "idx_users_email", UsersAliasName,
                ).OnColumns(UsersColEmail).Unique(),
            ),
        ),
        ID: idCol, Email: emailCol,
    }
}()
```
