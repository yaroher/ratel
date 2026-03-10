---
sidebar_position: 4
title: protoc-gen-ratel
---

# protoc-gen-ratel

Protoc plugin that generates Ratel table definitions from `.proto` files.

## Installation

```bash
go install github.com/yaroher/ratel/cmd/protoc-gen-ratel@latest
```

## Usage

```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --plain_out=. --plain_opt=paths=source_relative \
  --ratel_out=. --ratel_opt=paths=source_relative \
  schema.proto
```

## Prerequisites

You also need:

- `protoc-gen-go` — standard Go protobuf generator
- `protoc-gen-plain` — plain Go struct generator (no protobuf runtime)

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/yaroher/protoc-gen-go-plain/cmd/protoc-gen-plain@latest
```

## Generated Files

For `schema.proto`, the plugin generates `schema_ratel.pb.go` containing:

- Table alias types and constants
- Column alias types and constants
- Scanner struct with `GetTarget`, `GetSetter`, `GetValue`
- Table struct with typed column fields
- Global table variable with full configuration
- Relation definitions and loaders

## Proto Annotations

The plugin reads `ratel.table`, `ratel.column`, and `ratel.relation` options from your `.proto` files. See [Proto Options](/docs/api/proto-options) for the full reference.

## Example

```protobuf title="schema.proto"
syntax = "proto3";
package myapp;
option go_package = "myapp/models";

import "ratelproto/ratelproto.proto";

message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
  };

  int64 id = 1 [(ratel.column) = {
    constraints: { primary_key: true }
  }];
  string email = 2;
  string name = 3;
}
```

Run protoc:

```bash
protoc --go_out=. --plain_out=. --ratel_out=. schema.proto
```

Use the generated code:

```go
users, err := models.Users.Query(ctx, db,
    models.Users.SelectAll().
        Where(models.Users.Email.Eq("john@example.com")),
)
```
