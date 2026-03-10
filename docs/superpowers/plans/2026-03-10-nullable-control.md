# Nullable Control Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicit nullable control via `Constraint.nullable` and proto3 `optional` keyword support, with unified nullability logic replacing separate wrapper-type code paths.

**Architecture:** A single `isFieldNullable(col)` function determines nullability from three sources: wrapper type, proto3 optional keyword, or explicit `constraints.nullable`. All type-mapping functions (`getSchemaColumnType`, `getSchemaColumnConstructor`, Go type resolution) use this unified flag to pick between `Null*` and non-null variants via a single switch block instead of separate wrapper/scalar branches.

**Tech Stack:** protobuf, Go, protoc-gen-ratel codegen

---

## File Structure

| File | Responsibility |
|------|---------------|
| `ratelproto/ratelproto.proto` | Add `bool nullable = 5` to `Constraint` |
| `ratelproto/ratelproto.pb.go` | Regenerated from proto |
| `cmd/protoc-gen-ratel/types.go` | Unified nullable logic, refactored type mapping |
| `cmd/protoc-gen-ratel/tables.go` | Pass nullable info to GoType computation |
| `cmd/protoc-gen-ratel/casters.go` | Register nullable Timestamp/Duration casters |
| `pkg/ratelcast/casters.go` | Add nullable Timestamp/Duration conversion functions |
| `examples/proto/store.proto` | Add test fields with `nullable: true` and `optional` |

---

### Task 1: Add `nullable` field to Constraint proto

**Files:**
- Modify: `ratelproto/ratelproto.proto:188-195`

- [ ] **Step 1: Add nullable field to Constraint message**

In `ratelproto/ratelproto.proto`, add `bool nullable = 5` to `Constraint`:

```protobuf
message Constraint {
  bool unique = 1;
  bool primary_key = 2;
  string default_value = 3;

  // Raw SQL constraint (overrides above if set)
  string raw = 4;

  // Make column nullable regardless of proto type
  // Also applies implicitly for wrapper types and proto3 optional fields
  bool nullable = 5;
}
```

- [ ] **Step 2: Regenerate Go code from proto**

Run: `cd ratelproto && protoc --go_out=. --go_opt=paths=source_relative ratelproto.proto`

Verify `ratelproto.pb.go` has the new `Nullable` field in `Constraint` struct.

- [ ] **Step 3: Commit**

```bash
git add ratelproto/ratelproto.proto ratelproto/ratelproto.pb.go
git commit -m "feat(proto): add nullable field to Constraint message"
```

---

### Task 2: Add nullable casters for Timestamp and Duration

**Files:**
- Modify: `pkg/ratelcast/casters.go`

- [ ] **Step 1: Add nullable Timestamp casters**

Add to `pkg/ratelcast/casters.go`:

```go
// TimestampToNullableTime converts *timestamppb.Timestamp to *time.Time
func TimestampToNullableTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// NullableTimeToTimestamp converts *time.Time to *timestamppb.Timestamp
func NullableTimeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
```

- [ ] **Step 2: Add nullable Duration casters**

Add to `pkg/ratelcast/casters.go`:

```go
// DurationPbToNullableDuration converts *durationpb.Duration to *time.Duration
func DurationPbToNullableDuration(d *durationpb.Duration) *time.Duration {
	if d == nil {
		return nil
	}
	dur := d.AsDuration()
	return &dur
}

// NullableDurationToDurationPb converts *time.Duration to *durationpb.Duration
func NullableDurationToDurationPb(d *time.Duration) *durationpb.Duration {
	if d == nil {
		return nil
	}
	return durationpb.New(*d)
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /home/yaroher/devel/github/ratel && go build ./pkg/ratelcast/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add pkg/ratelcast/casters.go
git commit -m "feat(ratelcast): add nullable casters for Timestamp and Duration"
```

---

### Task 3: Register nullable casters in protoc-gen-ratel

**Files:**
- Modify: `cmd/protoc-gen-ratel/casters.go`

- [ ] **Step 1: Add nullable Timestamp casters to getRatelCasters()**

Add two entries to `getRatelCasters()` return slice:

```go
// *timestamppb.Timestamp -> *time.Time (nullable)
{
    SourceType: generator.GoType{
        Name:       "Timestamp",
        ImportPath: "google.golang.org/protobuf/types/known/timestamppb",
        IsPointer:  true,
    },
    TargetType: generator.GoType{
        Name:       "Time",
        ImportPath: "time",
        IsPointer:  true,
    },
    CasterIdent: generator.GoIdent{
        Name:       "TimestampToNullableTime",
        ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
    },
    IsFunc: true,
},
// *time.Time -> *timestamppb.Timestamp (nullable)
{
    SourceType: generator.GoType{
        Name:       "Time",
        ImportPath: "time",
        IsPointer:  true,
    },
    TargetType: generator.GoType{
        Name:       "Timestamp",
        ImportPath: "google.golang.org/protobuf/types/known/timestamppb",
        IsPointer:  true,
    },
    CasterIdent: generator.GoIdent{
        Name:       "NullableTimeToTimestamp",
        ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
    },
    IsFunc: true,
},
```

- [ ] **Step 2: Add nullable Duration casters to getRatelCasters()**

Add two entries:

```go
// *durationpb.Duration -> *time.Duration (nullable)
{
    SourceType: generator.GoType{
        Name:       "Duration",
        ImportPath: "google.golang.org/protobuf/types/known/durationpb",
        IsPointer:  true,
    },
    TargetType: generator.GoType{
        Name:       "Duration",
        ImportPath: "time",
        IsPointer:  true,
    },
    CasterIdent: generator.GoIdent{
        Name:       "DurationPbToNullableDuration",
        ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
    },
    IsFunc: true,
},
// *time.Duration -> *durationpb.Duration (nullable)
{
    SourceType: generator.GoType{
        Name:       "Duration",
        ImportPath: "time",
        IsPointer:  true,
    },
    TargetType: generator.GoType{
        Name:       "Duration",
        ImportPath: "google.golang.org/protobuf/types/known/durationpb",
        IsPointer:  true,
    },
    CasterIdent: generator.GoIdent{
        Name:       "NullableDurationToDurationPb",
        ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
    },
    IsFunc: true,
},
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./cmd/protoc-gen-ratel/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add cmd/protoc-gen-ratel/casters.go
git commit -m "feat(codegen): register nullable Timestamp/Duration casters"
```

---

### Task 4: Unify nullable logic in types.go

**Files:**
- Modify: `cmd/protoc-gen-ratel/types.go`

This is the core change. We add `isFieldNullable()`, then refactor all three type-mapping functions to use a single switch with nullable branching.

- [ ] **Step 1: Add helper functions**

Add at the top of `types.go` (after imports):

```go
// isWrapperType checks if a message field is a protobuf wrapper type
func isWrapperType(field *protogen.Field) bool {
	if field.Message == nil {
		return false
	}
	switch string(field.Message.Desc.FullName()) {
	case "google.protobuf.Int64Value", "google.protobuf.UInt64Value",
		"google.protobuf.Int32Value", "google.protobuf.UInt32Value",
		"google.protobuf.StringValue", "google.protobuf.BoolValue",
		"google.protobuf.FloatValue", "google.protobuf.DoubleValue",
		"google.protobuf.BytesValue":
		return true
	}
	return false
}

// isFieldNullable determines if a column should be nullable.
// Sources: wrapper type, proto3 optional keyword, explicit constraint.nullable.
func isFieldNullable(col *RatelColumn) bool {
	// Wrapper types are always nullable
	if isWrapperType(col.Field) {
		return true
	}
	// proto3 optional keyword
	if col.Field.Desc.HasOptionalKeyword() {
		return true
	}
	// Explicit constraint
	if col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.Nullable {
		return true
	}
	return false
}
```

- [ ] **Step 2: Refactor `protoFieldToSQLType` — no changes needed**

`protoFieldToSQLType` maps to base SQL types. Wrappers already map to the same base type as their scalar counterparts. No changes needed here.

- [ ] **Step 3: Refactor `protoFieldToGoType` into `protoFieldToBaseGoType`**

Rename `protoFieldToGoType` to `protoFieldToBaseGoType`. It always returns the non-pointer base type:

```go
// protoFieldToBaseGoType converts a protobuf field to its base (non-pointer) Go type string.
// For nullable fields, the caller wraps this in a pointer.
func protoFieldToBaseGoType(field *protogen.Field) string {
	if field.Message != nil {
		fullName := string(field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			return "time.Time"
		case "google.protobuf.Duration":
			return "time.Duration"
		// Wrapper types → same base type as their scalar
		case "google.protobuf.Int64Value":
			return "int64"
		case "google.protobuf.Int32Value":
			return "int32"
		case "google.protobuf.UInt64Value":
			return "uint64"
		case "google.protobuf.UInt32Value":
			return "uint32"
		case "google.protobuf.StringValue":
			return "string"
		case "google.protobuf.BoolValue":
			return "bool"
		case "google.protobuf.FloatValue":
			return "float32"
		case "google.protobuf.DoubleValue":
			return "float64"
		}

		if isTypeAlias(field.Message) {
			if len(field.Message.Fields) > 0 {
				return protoKindToGoType(field.Message.Fields[0].Desc.Kind())
			}
		}
	}

	return protoKindToGoType(field.Desc.Kind())
}
```

- [ ] **Step 4: Refactor `getSchemaColumnType` — unified switch**

Replace the entire function body. Single switch using `baseSQLType` + `nullable` flag:

```go
func getSchemaColumnType(col *RatelColumn, msgName string) string {
	isPK := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey
	nullable := isFieldNullable(col)
	alias := msgName + "ColumnAlias"

	kind := col.Field.Desc.Kind()
	if col.Field.Message != nil && isTypeAlias(col.Field.Message) {
		if len(col.Field.Message.Fields) > 0 {
			kind = col.Field.Message.Fields[0].Desc.Kind()
		}
	}

	// PK overrides — never nullable
	if isPK && kind == protoreflect.Int64Kind {
		return "schema.BigSerialColumnI[" + alias + "]"
	}
	if isPK && kind == protoreflect.StringKind {
		return "schema.TextColumnI[" + alias + "]"
	}

	// Well-known message types
	if col.Field.Message != nil && !isTypeAlias(col.Field.Message) {
		fullName := string(col.Field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			if nullable {
				return "schema.NullTimestamptzColumnI[" + alias + "]"
			}
			return "schema.TimestamptzColumnI[" + alias + "]"
		case "google.protobuf.Duration":
			if nullable {
				return "schema.NullIntervalColumnI[" + alias + "]"
			}
			return "schema.IntervalColumnI[" + alias + "]"
		case "google.protobuf.Struct":
			return "schema.JSONColumnI[" + alias + "]"
		}
	}

	// Scalars and wrapper types (wrappers resolved to base kind above)
	prefix := "schema."
	if nullable {
		prefix = "schema.Null"
	}
	switch kind {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return prefix + "IntegerColumnI[" + alias + "]"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return prefix + "BigIntColumnI[" + alias + "]"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return prefix + "IntegerColumnI[" + alias + "]"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return prefix + "BigIntColumnI[" + alias + "]"
	case protoreflect.StringKind:
		return prefix + "TextColumnI[" + alias + "]"
	case protoreflect.BoolKind:
		return prefix + "BooleanColumnI[" + alias + "]"
	case protoreflect.FloatKind:
		return prefix + "RealColumnI[" + alias + "]"
	case protoreflect.DoubleKind:
		return prefix + "DoublePrecisionColumnI[" + alias + "]"
	case protoreflect.BytesKind:
		return prefix + "ByteaColumnI[" + alias + "]"
	default:
		return prefix + "TextColumnI[" + alias + "]"
	}
}
```

- [ ] **Step 5: Refactor `getSchemaColumnConstructor` — unified switch**

Replace the entire function body. Same pattern:

```go
func getSchemaColumnConstructor(col *RatelColumn, constName string, msgName string) string {
	isPK := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey
	isUnique := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.Unique
	nullable := isFieldNullable(col)
	defaultVal := ""
	if col.Options != nil && col.Options.Constraints != nil {
		defaultVal = col.Options.Constraints.DefaultValue
	}

	kind := col.Field.Desc.Kind()
	if col.Field.Message != nil && isTypeAlias(col.Field.Message) {
		if len(col.Field.Message.Fields) > 0 {
			kind = col.Field.Message.Fields[0].Desc.Kind()
		}
	}

	// Build options
	var opts []string
	if isPK {
		opts = append(opts, "ddl.WithPrimaryKey["+msgName+"ColumnAlias]()")
	}
	if isUnique {
		opts = append(opts, "ddl.WithUnique["+msgName+"ColumnAlias]()")
	}
	if defaultVal != "" {
		opts = append(opts, "ddl.WithDefault["+msgName+"ColumnAlias](\""+defaultVal+"\")")
	}

	optStr := ""
	for _, opt := range opts {
		optStr += ", " + opt
	}

	// Well-known message types
	if col.Field.Message != nil && !isTypeAlias(col.Field.Message) {
		fullName := string(col.Field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			if nullable {
				return "schema.NullTimestamptzColumn(" + constName + optStr + ")"
			}
			return "schema.TimestamptzColumn(" + constName + optStr + ")"
		case "google.protobuf.Duration":
			if nullable {
				return "schema.NullIntervalColumn(" + constName + optStr + ")"
			}
			return "schema.IntervalColumn(" + constName + optStr + ")"
		}
	}

	// PK override
	if isPK && kind == protoreflect.Int64Kind {
		return "schema.BigSerialColumn(" + constName + optStr + ")"
	}

	prefix := "schema."
	if nullable {
		prefix = "schema.Null"
	}
	switch kind {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return prefix + "IntegerColumn(" + constName + optStr + ")"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return prefix + "BigIntColumn(" + constName + optStr + ")"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return prefix + "IntegerColumn(" + constName + optStr + ")"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return prefix + "BigIntColumn(" + constName + optStr + ")"
	case protoreflect.StringKind:
		return prefix + "TextColumn(" + constName + optStr + ")"
	case protoreflect.BoolKind:
		return prefix + "BooleanColumn(" + constName + optStr + ")"
	case protoreflect.FloatKind:
		return prefix + "RealColumn(" + constName + optStr + ")"
	case protoreflect.DoubleKind:
		return prefix + "DoublePrecisionColumn(" + constName + optStr + ")"
	case protoreflect.BytesKind:
		return prefix + "ByteaColumn(" + constName + optStr + ")"
	default:
		return prefix + "TextColumn(" + constName + optStr + ")"
	}
}
```

- [ ] **Step 6: Verify it compiles**

Run: `go build ./cmd/protoc-gen-ratel/...`
Expected: success

- [ ] **Step 7: Commit**

```bash
git add cmd/protoc-gen-ratel/types.go
git commit -m "feat(codegen): unified nullable logic in type mapping"
```

---

### Task 5: Wire nullable into GoType computation in tables.go

**Files:**
- Modify: `cmd/protoc-gen-ratel/tables.go:83-91` and `145-154`

The `GoType` on `RatelColumn` must reflect nullable. Since `isFieldNullable` needs a `RatelColumn`, we compute GoType after constructing the column.

- [ ] **Step 1: Update `collectRatelTables` column construction**

Change lines 83-91 in `tables.go`:

```go
col := &RatelColumn{
    Field:     field,
    Options:   colOpts,
    SQLName:   strcase.ToSnake(string(field.Desc.Name())),
    SQLType:   protoFieldToSQLType(field),
    GoName:    field.GoName,
    IsSkipped: colOpts != nil && colOpts.Skip,
}
// GoType depends on nullable which needs the full RatelColumn
baseType := protoFieldToBaseGoType(field)
if isFieldNullable(col) {
    col.GoType = "*" + baseType
} else {
    col.GoType = baseType
}
```

- [ ] **Step 2: Update `collectEmbeddedColumns` column construction**

Same change for lines 145-154:

```go
col := &RatelColumn{
    Field:      field,
    Options:    colOpts,
    SQLName:    strcase.ToSnake(string(field.Desc.Name())),
    SQLType:    protoFieldToSQLType(field),
    GoName:     field.GoName,
    IsSkipped:  colOpts != nil && colOpts.Skip,
    IsEmbedded: true,
}
baseType := protoFieldToBaseGoType(field)
if isFieldNullable(col) {
    col.GoType = "*" + baseType
} else {
    col.GoType = baseType
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./cmd/protoc-gen-ratel/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add cmd/protoc-gen-ratel/tables.go
git commit -m "feat(codegen): compute GoType with nullable awareness"
```

---

### Task 6: End-to-end test with example proto

**Files:**
- Modify: `examples/proto/store.proto` (add nullable test fields)

- [ ] **Step 1: Add nullable fields to store.proto**

Add fields to an existing message (e.g. `User`) to test all three nullable sources:

```protobuf
// Nullable via constraint
google.protobuf.Timestamp email_confirmed_at = 8 [(ratel.column) = {
    constraints: { nullable: true }
}];

// Nullable via proto3 optional
optional string nickname = 9;

// Nullable via proto3 optional on Timestamp
optional google.protobuf.Timestamp deleted_at = 10;
```

- [ ] **Step 2: Regenerate and verify**

Run the full codegen:
```bash
cd /home/yaroher/devel/github/ratel && make generate
```
(or the equivalent protoc command for examples)

- [ ] **Step 3: Verify generated code**

Check the generated `store_ratel.pb.go`:
- `EmailConfirmedAt` should be `schema.NullTimestamptzColumnI[UserColumnAlias]`
- `Nickname` should be `schema.NullTextColumnI[UserColumnAlias]`
- `DeletedAt` should be `schema.NullTimestamptzColumnI[UserColumnAlias]`
- Constructors should use `NullTimestamptzColumn`, `NullTextColumn`

- [ ] **Step 4: Verify it compiles**

Run: `go build ./examples/...` or `go build ./...`
Expected: success

- [ ] **Step 5: Commit**

```bash
git add examples/
git commit -m "feat(examples): add nullable field examples"
```
