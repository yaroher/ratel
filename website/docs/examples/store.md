---
sidebar_position: 1
title: Store Example
---

# Store Example

A complete e-commerce schema demonstrating all Ratel features.

## Schema Overview

```
store.currency     — reference table (code PK)
store.users        — user accounts
store.profiles     — user profiles (HasOne)
store.categories   — self-referencing tree
store.products     — product catalog
store.orders       — user orders
store.order_items  — order line items (composite PK)
product_categories — auto-generated pivot table (ManyToMany)
```

## Proto Definition

```protobuf title="store.proto"
syntax = "proto3";
package store;

import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";
import "ratelproto/ratelproto.proto";

// Shared base fields
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

// Simple reference table with string PK
message Currency {
  option (ratel.table) = {
    generate: true
    table_name: "currency"
  };
  string code = 1 [(ratel.column) = {
    constraints: { primary_key: true, check: "code <> ''" }
  }];
  string name = 2;
}

// User with relations to orders and profile
message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
    schema: "store"
    indexes: [{ columns: ["email"], unique: true }]
  };

  BaseEntity base = 1;
  string email = 2 [(ratel.column) = { constraints: { unique: true } }];
  string full_name = 3;
  bool is_active = 4 [(ratel.column) = { constraints: { default_value: "true" } }];

  // Relations
  repeated Order orders = 10 [(ratel.relation) = {
    one_to_many: { ref_name: "user_id", on_delete_cascade: true }
  }];
  Profile profile = 11 [(ratel.relation) = {
    has_one: { ref_name: "user_id" }
  }];
}

// Self-referencing category tree
message Category {
  option (ratel.table) = {
    generate: true
    table_name: "categories"
    schema: "store"
  };

  BaseEntity base = 1;
  string name = 2;
  google.protobuf.Int64Value parent_id = 3 [(ratel.column) = {
    constraints: {
      references_schema: "store"
      references_table: "categories"
      on_delete: SET_NULL
    }
  }];

  repeated Product products = 10 [(ratel.relation) = {
    many_to_many: {
      ref_on_delete_cascade: true
      back_ref_on_delete_cascade: true
    }
  }];
}

// Order with status constraint
message Order {
  option (ratel.table) = {
    generate: true
    table_name: "orders"
    schema: "store"
    constraints: ["CHECK (status IN ('NEW', 'PAID', 'CANCELLED'))"]
  };

  BaseEntity base = 1;
  int64 user_id = 2 [(ratel.column) = {
    constraints: {
      references_schema: "store"
      references_table: "users"
      on_delete: CASCADE
    }
  }];
  string status = 3 [(ratel.column) = { constraints: { default_value: "'NEW'" } }];
  string currency = 4 [(ratel.column) = {
    constraints: {
      references_table: "currency"
      references_column: "code"
      on_delete: RESTRICT
    }
  }];

  repeated OrderItem items = 10 [(ratel.relation) = {
    one_to_many: { ref_name: "order_id", on_delete_cascade: true }
  }];
  User user = 11 [(ratel.relation) = {
    belongs_to: { foreign_key: "user_id" }
  }];
}

// Order line item with composite PK
message OrderItem {
  option (ratel.table) = {
    generate: true
    table_name: "order_items"
    schema: "store"
    primary_key: { columns: ["order_id", "line_no"] }
    unique: [{ columns: ["order_id", "product_id"] }]
  };

  int64 order_id = 1 [(ratel.column) = {
    constraints: {
      references_schema: "store"
      references_table: "orders"
      on_delete: CASCADE
    }
  }];
  int32 line_no = 2;
  int64 product_id = 3 [(ratel.column) = {
    constraints: {
      references_schema: "store"
      references_table: "products"
      on_delete: RESTRICT
    }
  }];
  int32 quantity = 4 [(ratel.column) = { constraints: { default_value: "1" } }];
}
```

## Usage

```go
// Create schema
stmts, _ := ddl.SchemaSortedStatements(
    Currency, Users, Profiles, Categories,
    Products, Orders, OrderItems,
)
for _, stmt := range stmts {
    pool.Exec(ctx, stmt)
}

// Query users with orders
users, _ := Users.Query(ctx, pool,
    Users.SelectAll().
        Where(Users.IsActive.Eq(true)).
        OrderByDESC(Users.CreatedAt),
)
for _, u := range users {
    fmt.Printf("%s has %d orders\n", u.Email, len(u.Orders))
}

// Create an order in a transaction
tx, _ := pool.Begin(ctx)
order, _ := Orders.QueryRow(ctx, tx,
    Orders.Insert().
        Columns(Orders.UserID, Orders.Status, Orders.Currency).
        Values(users[0].ID, "NEW", "USD").
        Returning(Orders.ID),
)
OrderItems.QueryRow(ctx, tx,
    OrderItems.Insert().
        Columns(OrderItems.OrderID, OrderItems.LineNo, OrderItems.ProductID, OrderItems.Quantity).
        Values(order.ID, 1, productID, 2).
        Returning(OrderItems.OrderID),
)
tx.Commit(ctx)
```
