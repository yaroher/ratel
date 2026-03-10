# Nullable Control for Columns

## Problem

Ratel determines nullability from proto field type:
- Scalars (`string`, `int64`) → NOT NULL
- Wrapper types (`StringValue`, `Int64Value`) → nullable
- `Timestamp`, `Duration` → NOT NULL, no nullable wrapper exists

This makes it impossible to:
1. Make `Timestamp` fields nullable (e.g. `email_confirmed_at TIMESTAMPTZ NULL`)
2. Make scalar fields nullable (e.g. `user_id TEXT NULL` for FK with ON DELETE SET NULL)

## Solution

### Two new sources of nullable

1. **Explicit flag** `nullable: true` in `Constraint`:
```protobuf
google.protobuf.Timestamp email_confirmed_at = 4 [(ratel.column) = {
  constraints: { nullable: true }
}];
```

2. **proto3 `optional`** keyword — automatically makes field nullable:
```protobuf
optional string user_id = 5;
// → TEXT NULL
```

### Unified nullable logic

Single function determines nullability from all sources:

```
isNullable =
  wrapper type (StringValue, Int64Value, ...)
  OR proto3 optional keyword
  OR constraint.nullable == true
```

Wrapper types no longer have separate code paths. They map to their base SQL type with `isNullable=true`, going through the same logic as scalars with `nullable: true`.

Example: `google.protobuf.StringValue` → base type TEXT + isNullable=true → NullTextColumnI.
Same path as `string` + `nullable: true`.

### Go types

Nullable fields use pointer types (consistent with current wrapper behavior):
- `string` → `*string`
- `int64` → `*int64`
- `time.Time` → `*time.Time`

### Changes

| File | Change |
|------|--------|
| `ratelproto/ratelproto.proto` | Add `bool nullable = 5` to `Constraint` |
| `cmd/protoc-gen-ratel/types.go` | Add `isNullable()` helper; unify `getSchemaColumnType()` and `getSchemaColumnConstructor()` to use single switch + nullable flag instead of separate wrapper/scalar branches |
| `cmd/protoc-gen-ratel/types.go` | `protoFieldToGoType()` returns pointer type when nullable |
| Casters | Add nullable Timestamp cast: `*timestamppb.Timestamp → *time.Time` |

### Not in scope

- `not_null: true` to force wrapper types NOT NULL (can add later)
- Changes to `pkg/schema` or `pkg/ddl` — all Null* variants already exist
