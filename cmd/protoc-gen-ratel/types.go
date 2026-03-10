package main

import (
	"strings"

	"github.com/yaroher/protoc-gen-go-plain/goplain"
	"github.com/yaroher/ratel/ratelproto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// escapeGoString escapes double quotes and backslashes for embedding in a Go string literal.
func escapeGoString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// referenceActionToSQL converts ReferenceAction enum to SQL string.
func referenceActionToSQL(action ratelproto.ReferenceAction) string {
	switch action {
	case ratelproto.ReferenceAction_CASCADE:
		return "CASCADE"
	case ratelproto.ReferenceAction_SET_NULL:
		return "SET NULL"
	case ratelproto.ReferenceAction_SET_DEFAULT:
		return "SET DEFAULT"
	case ratelproto.ReferenceAction_RESTRICT:
		return "RESTRICT"
	default:
		return ""
	}
}

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
// Sources: wrapper type (e.g. StringValue), proto3 optional keyword.
func isFieldNullable(col *RatelColumn) bool {
	return isWrapperType(col.Field) || col.Field.Desc.HasOptionalKeyword()
}

// computeGoType returns the Go type for a column, wrapping in pointer if nullable.
// PK fields are never nullable.
func computeGoType(col *RatelColumn) string {
	isPK := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey
	base := protoFieldToBaseGoType(col.Field)
	if !isPK && isFieldNullable(col) {
		return "*" + base
	}
	return base
}

// protoFieldToSQLType converts a protobuf field to SQL type
func protoFieldToSQLType(field *protogen.Field) string {
	// Check for well-known types first
	if field.Message != nil {
		fullName := string(field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			return "TIMESTAMPTZ"
		case "google.protobuf.Duration":
			return "INTERVAL"
		case "google.protobuf.Struct":
			return "JSONB"
		case "google.protobuf.Int64Value", "google.protobuf.UInt64Value":
			return "BIGINT"
		case "google.protobuf.Int32Value", "google.protobuf.UInt32Value":
			return "INTEGER"
		case "google.protobuf.StringValue":
			return "TEXT"
		case "google.protobuf.BoolValue":
			return "BOOLEAN"
		case "google.protobuf.FloatValue":
			return "REAL"
		case "google.protobuf.DoubleValue":
			return "DOUBLE PRECISION"
		case "google.protobuf.BytesValue":
			return "BYTEA"
		}

		// Check for type_alias messages
		if isTypeAlias(field.Message) {
			// Get the wrapped scalar type
			if len(field.Message.Fields) > 0 {
				valueField := field.Message.Fields[0]
				return protoKindToSQLType(valueField.Desc.Kind())
			}
		}
	}

	return protoKindToSQLType(field.Desc.Kind())
}

// isTypeAlias checks if a message is a type_alias
func isTypeAlias(msg *protogen.Message) bool {
	if msg == nil {
		return false
	}
	opts := msg.Desc.Options()
	if opts == nil {
		return false
	}
	ext := proto.GetExtension(opts, goplain.E_Message)
	if ext == nil {
		return false
	}
	msgOpts, ok := ext.(*goplain.MessageOptions)
	if !ok || msgOpts == nil {
		return false
	}
	return msgOpts.TypeAlias
}

// protoKindToSQLType converts a protoreflect.Kind to SQL type
func protoKindToSQLType(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "BOOLEAN"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "INTEGER"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "BIGINT"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "INTEGER"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "BIGINT"
	case protoreflect.FloatKind:
		return "REAL"
	case protoreflect.DoubleKind:
		return "DOUBLE PRECISION"
	case protoreflect.StringKind:
		return "TEXT"
	case protoreflect.BytesKind:
		return "BYTEA"
	default:
		return "TEXT"
	}
}

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
		// Wrapper types -> same base type as their scalar
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
		case "google.protobuf.BytesValue":
			return "[]byte"
		}

		if isTypeAlias(field.Message) {
			if len(field.Message.Fields) > 0 {
				return protoKindToGoType(field.Message.Fields[0].Desc.Kind())
			}
		}
	}

	return protoKindToGoType(field.Desc.Kind())
}

// protoKindToGoType converts a protoreflect.Kind to Go type string
func protoKindToGoType(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	default:
		return "interface{}"
	}
}

func getSchemaColumnType(col *RatelColumn, msgName string) string {
	isPK := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey
	nullable := isFieldNullable(col)
	alias := msgName + "ColumnAlias"

	// Well-known message types with special SQL mappings
	if col.Field.Message != nil && !isTypeAlias(col.Field.Message) && !isWrapperType(col.Field) {
		fullName := string(col.Field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			if isPK {
				return "schema.TimestamptzColumnI[" + alias + "]"
			}
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
			if nullable {
				return "schema.NullJSONColumnI[" + alias + "]"
			}
			return "schema.JSONColumnI[" + alias + "]"
		}
	}

	// Resolve kind: for type aliases and wrapper types, use the inner scalar kind
	kind := col.Field.Desc.Kind()
	if col.Field.Message != nil && (isTypeAlias(col.Field.Message) || isWrapperType(col.Field)) {
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

	// Scalars, wrapper types, and type aliases
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

func getSchemaColumnConstructor(col *RatelColumn, constName string, msgName string) string {
	isPK := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey
	isUnique := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.Unique
	nullable := isFieldNullable(col)
	defaultVal := ""
	checkExpr := ""
	refTable := ""
	refColumn := ""
	var onDelete, onUpdate ratelproto.ReferenceAction
	if col.Options != nil && col.Options.Constraints != nil {
		defaultVal = col.Options.Constraints.DefaultValue
		checkExpr = col.Options.Constraints.Check
		refTable = col.Options.Constraints.ReferencesTable
		refColumn = col.Options.Constraints.ReferencesColumn
		onDelete = col.Options.Constraints.OnDelete
		onUpdate = col.Options.Constraints.OnUpdate
		// Build qualified table name from schema + table
		if refSchema := col.Options.Constraints.ReferencesSchema; refSchema != "" {
			refTable = `"` + refSchema + `"."` + refTable + `"`
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
		opts = append(opts, "ddl.WithDefault["+msgName+"ColumnAlias](\""+escapeGoString(defaultVal)+"\")")
	}
	if checkExpr != "" {
		opts = append(opts, "ddl.WithCheck["+msgName+"ColumnAlias](\""+escapeGoString(checkExpr)+"\")")
	}
	if refTable != "" {
		opts = append(opts, "ddl.WithReferences["+msgName+"ColumnAlias](\""+escapeGoString(refTable)+"\", \""+escapeGoString(refColumn)+"\")")
	}
	if onDelete != ratelproto.ReferenceAction_NO_ACTION {
		opts = append(opts, "ddl.WithOnDelete["+msgName+"ColumnAlias](\""+referenceActionToSQL(onDelete)+"\")")
	}
	if onUpdate != ratelproto.ReferenceAction_NO_ACTION {
		opts = append(opts, "ddl.WithOnUpdate["+msgName+"ColumnAlias](\""+referenceActionToSQL(onUpdate)+"\")")
	}

	optStr := ""
	for _, opt := range opts {
		optStr += ", " + opt
	}

	// Well-known message types with special SQL mappings
	if col.Field.Message != nil && !isTypeAlias(col.Field.Message) && !isWrapperType(col.Field) {
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
		case "google.protobuf.Struct":
			if nullable {
				return "schema.NullJSONColumn(" + constName + optStr + ")"
			}
			return "schema.JSONColumn(" + constName + optStr + ")"
		}
	}

	// Resolve kind: for type aliases and wrapper types, use the inner scalar kind
	kind := col.Field.Desc.Kind()
	if col.Field.Message != nil && (isTypeAlias(col.Field.Message) || isWrapperType(col.Field)) {
		if len(col.Field.Message.Fields) > 0 {
			kind = col.Field.Message.Fields[0].Desc.Kind()
		}
	}

	// PK overrides — never nullable, must come first
	if isPK && kind == protoreflect.Int64Kind {
		return "schema.BigSerialColumn(" + constName + optStr + ")"
	}
	if isPK && kind == protoreflect.StringKind {
		return "schema.TextColumn(" + constName + optStr + ")"
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
