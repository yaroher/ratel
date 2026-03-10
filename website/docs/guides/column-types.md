---
sidebar_position: 2
title: Column Types
---

# Column Types

## Type Mapping

### Scalar Types

| Proto Type | Go Type | SQL Type | Column Constructor |
|-----------|---------|----------|-------------------|
| `int32` | `int32` | `INTEGER` | `IntegerColumn` |
| `int64` | `int64` | `BIGINT` | `BigIntColumn` |
| `uint32` | `uint32` | `INTEGER` | `IntegerColumn` |
| `uint64` | `uint64` | `BIGINT` | `BigIntColumn` |
| `float` | `float32` | `REAL` | `RealColumn` |
| `double` | `float64` | `DOUBLE PRECISION` | `DoubleColumn` |
| `bool` | `bool` | `BOOLEAN` | `BooleanColumn` |
| `string` | `string` | `TEXT` | `TextColumn` |
| `bytes` | `[]byte` | `BYTEA` | `ByteaColumn` |

### Well-Known Types

| Proto Type | Go Type | SQL Type | Column Constructor |
|-----------|---------|----------|-------------------|
| `google.protobuf.Timestamp` | `time.Time` | `TIMESTAMPTZ` | `TimestamptzColumn` |
| `google.protobuf.Duration` | `time.Duration` | `INTERVAL` | `IntervalColumn` |
| `google.protobuf.Struct` | `map[string]any` | `JSONB` | `JSONBColumn` |

### Wrapper Types (Nullable)

Wrapper types map to nullable columns:

| Proto Type | Go Type | SQL Type | Column Constructor |
|-----------|---------|----------|-------------------|
| `google.protobuf.Int32Value` | `*int32` | `INTEGER NULL` | `NullIntegerColumn` |
| `google.protobuf.Int64Value` | `*int64` | `BIGINT NULL` | `NullBigIntColumn` |
| `google.protobuf.UInt32Value` | `*uint32` | `INTEGER NULL` | `NullIntegerColumn` |
| `google.protobuf.UInt64Value` | `*uint64` | `BIGINT NULL` | `NullBigIntColumn` |
| `google.protobuf.FloatValue` | `*float32` | `REAL NULL` | `NullRealColumn` |
| `google.protobuf.DoubleValue` | `*float64` | `DOUBLE PRECISION NULL` | `NullDoubleColumn` |
| `google.protobuf.BoolValue` | `*bool` | `BOOLEAN NULL` | `NullBooleanColumn` |
| `google.protobuf.StringValue` | `*string` | `TEXT NULL` | `NullTextColumn` |
| `google.protobuf.BytesValue` | `[]byte` | `BYTEA NULL` | `NullByteaColumn` |

### Serial Types

Primary key fields with integer types automatically use serial variants:

| Proto Type | With `primary_key: true` | SQL Type |
|-----------|-------------------------|----------|
| `int32` | Yes | `SERIAL` |
| `int64` | Yes | `BIGSERIAL` |

## Nullability

A column is nullable when:

1. The field uses a **wrapper type** (`Int64Value`, `StringValue`, etc.)
2. The field uses proto3 `optional` keyword
3. The field is explicitly declared nullable in Go models

```protobuf
message User {
  // NOT NULL — required field
  string email = 1;

  // NULL — wrapper type
  google.protobuf.StringValue bio = 2;

  // NULL — optional keyword
  optional string nickname = 3;
}
```

## Type Aliases

Define custom types that map to scalar types:

```protobuf
// EntityID resolves to int64 → BIGINT
message EntityID {
  option (goplain.message) = { generate: true, type_alias: "int64" };
}

message User {
  EntityID id = 1 [(ratel.column) = {
    constraints: { primary_key: true }
  }];
  // Generates: BIGSERIAL PRIMARY KEY
}
```

## Available Column Constructors (Go)

```go
// Integer types
schema.SmallIntColumn[C](alias, opts...)
schema.IntegerColumn[C](alias, opts...)
schema.BigIntColumn[C](alias, opts...)
schema.SerialColumn[C](alias, opts...)
schema.SmallSerialColumn[C](alias, opts...)
schema.BigSerialColumn[C](alias, opts...)

// Floating point
schema.RealColumn[C](alias, opts...)
schema.DoubleColumn[C](alias, opts...)
schema.NumericColumn[C](alias, precision, scale, opts...)

// Text
schema.TextColumn[C](alias, opts...)
schema.VarcharColumn[C](alias, length, opts...)
schema.CharColumn[C](alias, length, opts...)

// Date/Time
schema.DateColumn[C](alias, opts...)
schema.TimeColumn[C](alias, opts...)
schema.TimestampColumn[C](alias, opts...)
schema.TimestamptzColumn[C](alias, opts...)

// Other
schema.BooleanColumn[C](alias, opts...)
schema.UUIDColumn[C](alias, opts...)
schema.JSONColumn[C](alias, opts...)
schema.JSONBColumn[C](alias, opts...)
schema.ByteaColumn[C](alias, opts...)
```

Every type has a `Null` variant (e.g., `NullBigIntColumn`) for nullable columns.
