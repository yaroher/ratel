package main

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/yaroher/protoc-gen-go-plain/goplain"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/yaroher/ratel/ratelproto"
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
	baseType := protoFieldToBaseSQLType(field)
	if field.Desc.IsList() {
		return baseType + "[]"
	}
	return baseType
}

// protoFieldToBaseSQLType returns the base SQL type without array suffix.
func protoFieldToBaseSQLType(field *protogen.Field) string {
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
	isArray := col.Field.Desc.IsList()
	alias := msgName + "ColumnAlias"

	// Repeated (array) fields
	if isArray {
		kind := col.Field.Desc.Kind()
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
			protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return "schema.IntegerArrayColumnI[" + alias + "]"
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
			protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return "schema.BigIntArrayColumnI[" + alias + "]"
		case protoreflect.BoolKind:
			return "schema.BooleanArrayColumnI[" + alias + "]"
		case protoreflect.FloatKind:
			return "schema.RealArrayColumnI[" + alias + "]"
		case protoreflect.DoubleKind:
			return "schema.DoublePrecisionArrayColumnI[" + alias + "]"
		case protoreflect.BytesKind:
			return "schema.ByteaArrayColumnI[" + alias + "]"
		default:
			return "schema.TextArrayColumnI[" + alias + "]"
		}
	}

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
	if !nullable && !isPK {
		opts = append(opts, "ddl.WithNotNull["+msgName+"ColumnAlias]()")
	}

	optStr := ""
	for _, opt := range opts {
		optStr += ", " + opt
	}

	// Repeated (array) fields
	if col.Field.Desc.IsList() {
		kind := col.Field.Desc.Kind()
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
			protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return "schema.IntegerArrayColumn(" + constName + optStr + ")"
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
			protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return "schema.BigIntArrayColumn(" + constName + optStr + ")"
		case protoreflect.BoolKind:
			return "schema.BooleanArrayColumn(" + constName + optStr + ")"
		case protoreflect.FloatKind:
			return "schema.RealArrayColumn(" + constName + optStr + ")"
		case protoreflect.DoubleKind:
			return "schema.DoublePrecisionArrayColumn(" + constName + optStr + ")"
		case protoreflect.BytesKind:
			return "schema.ByteaArrayColumn(" + constName + optStr + ")"
		default:
			return "schema.TextArrayColumn(" + constName + optStr + ")"
		}
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

// getVirtualSchemaColumnType returns the schema column type for a virtual column.
func getVirtualSchemaColumnType(col *RatelColumn, msgName string) string {
	alias := msgName + "ColumnAlias"
	prefix := "schema."
	if col.VirtualDef != nil && col.VirtualDef.IsNullable {
		prefix = "schema.Null"
	}

	sqlType := strings.ToUpper(col.SQLType)
	switch {
	case sqlType == "BIGINT" || sqlType == "INT8" || sqlType == "BIGSERIAL" || sqlType == "SERIAL8":
		return prefix + "BigIntColumnI[" + alias + "]"
	case sqlType == "INTEGER" || sqlType == "INT" || sqlType == "INT4" || sqlType == "SERIAL" || sqlType == "SERIAL4":
		return prefix + "IntegerColumnI[" + alias + "]"
	case sqlType == "SMALLINT" || sqlType == "INT2":
		return prefix + "SmallIntColumnI[" + alias + "]"
	case sqlType == "BOOLEAN" || sqlType == "BOOL":
		return prefix + "BooleanColumnI[" + alias + "]"
	case sqlType == "REAL" || sqlType == "FLOAT4":
		return prefix + "RealColumnI[" + alias + "]"
	case sqlType == "DOUBLE PRECISION" || sqlType == "FLOAT8":
		return prefix + "DoublePrecisionColumnI[" + alias + "]"
	case sqlType == "TIMESTAMPTZ" || sqlType == "TIMESTAMP WITH TIME ZONE":
		return prefix + "TimestamptzColumnI[" + alias + "]"
	case sqlType == "TIMESTAMP" || sqlType == "TIMESTAMP WITHOUT TIME ZONE":
		return prefix + "TimestampColumnI[" + alias + "]"
	case sqlType == "DATE":
		return prefix + "DateColumnI[" + alias + "]"
	case sqlType == "INTERVAL":
		return prefix + "IntervalColumnI[" + alias + "]"
	case sqlType == "UUID":
		return prefix + "UUIDColumnI[" + alias + "]"
	case sqlType == "JSONB":
		return prefix + "JSONBColumnI[" + alias + "]"
	case sqlType == "JSON":
		return prefix + "JSONColumnI[" + alias + "]"
	case sqlType == "BYTEA":
		return prefix + "ByteaColumnI[" + alias + "]"
	default:
		return prefix + "TextColumnI[" + alias + "]"
	}
}

// columnConstName returns the Go constant name for a column (works for both regular and virtual).
func columnConstName(col *RatelColumn, msgName string) string {
	if col.IsVirtual {
		return msgName + "Column" + strcase.ToCamel(col.SQLName)
	}
	return msgName + "Column" + strcase.ToCamel(string(col.Field.Desc.Name()))
}

// getVirtualColumnConstructor returns the schema column constructor call for a virtual column.
func getVirtualColumnConstructor(col *RatelColumn, constName string, msgName string) string {
	vc := col.VirtualDef
	alias := msgName + "ColumnAlias"

	// Build options from constraints
	var opts []string
	if vc.Constraints != nil {
		if vc.Constraints.PrimaryKey {
			opts = append(opts, "ddl.WithPrimaryKey["+alias+"]()")
		}
		if vc.Constraints.Unique {
			opts = append(opts, "ddl.WithUnique["+alias+"]()")
		}
		if vc.Constraints.DefaultValue != "" {
			opts = append(opts, "ddl.WithDefault["+alias+"](\""+escapeGoString(vc.Constraints.DefaultValue)+"\")")
		}
		if vc.Constraints.Check != "" {
			opts = append(opts, "ddl.WithCheck["+alias+"](\""+escapeGoString(vc.Constraints.Check)+"\")")
		}
		if vc.Constraints.ReferencesTable != "" {
			refTable := vc.Constraints.ReferencesTable
			if vc.Constraints.ReferencesSchema != "" {
				refTable = `"` + vc.Constraints.ReferencesSchema + `"."` + refTable + `"`
			}
			opts = append(opts, "ddl.WithReferences["+alias+"](\""+escapeGoString(refTable)+"\", \""+escapeGoString(vc.Constraints.ReferencesColumn)+"\")")
		}
		if vc.Constraints.OnDelete != ratelproto.ReferenceAction_NO_ACTION {
			opts = append(opts, "ddl.WithOnDelete["+alias+"](\""+referenceActionToSQL(vc.Constraints.OnDelete)+"\")")
		}
		if vc.Constraints.OnUpdate != ratelproto.ReferenceAction_NO_ACTION {
			opts = append(opts, "ddl.WithOnUpdate["+alias+"](\""+referenceActionToSQL(vc.Constraints.OnUpdate)+"\")")
		}
	}
	if !vc.IsNullable {
		opts = append(opts, "ddl.WithNotNull["+alias+"]()")
	}

	optStr := ""
	for _, opt := range opts {
		optStr += ", " + opt
	}

	// Map SQL type string to schema constructor
	prefix := "schema."
	if vc.IsNullable {
		prefix = "schema.Null"
	}

	sqlType := strings.ToUpper(vc.SqlType)
	switch {
	case sqlType == "BIGINT" || sqlType == "INT8":
		return prefix + "BigIntColumn(" + constName + optStr + ")"
	case sqlType == "INTEGER" || sqlType == "INT" || sqlType == "INT4":
		return prefix + "IntegerColumn(" + constName + optStr + ")"
	case sqlType == "SMALLINT" || sqlType == "INT2":
		return prefix + "SmallIntColumn(" + constName + optStr + ")"
	case sqlType == "BIGSERIAL" || sqlType == "SERIAL8":
		return "schema.BigSerialColumn(" + constName + optStr + ")"
	case sqlType == "SERIAL" || sqlType == "SERIAL4":
		return "schema.SerialColumn(" + constName + optStr + ")"
	case sqlType == "TEXT":
		return prefix + "TextColumn(" + constName + optStr + ")"
	case strings.HasPrefix(sqlType, "VARCHAR"):
		return prefix + "TextColumn(" + constName + optStr + ")"
	case sqlType == "BOOLEAN" || sqlType == "BOOL":
		return prefix + "BooleanColumn(" + constName + optStr + ")"
	case sqlType == "REAL" || sqlType == "FLOAT4":
		return prefix + "RealColumn(" + constName + optStr + ")"
	case sqlType == "DOUBLE PRECISION" || sqlType == "FLOAT8":
		return prefix + "DoublePrecisionColumn(" + constName + optStr + ")"
	case sqlType == "TIMESTAMPTZ" || sqlType == "TIMESTAMP WITH TIME ZONE":
		return prefix + "TimestamptzColumn(" + constName + optStr + ")"
	case sqlType == "TIMESTAMP" || sqlType == "TIMESTAMP WITHOUT TIME ZONE":
		return prefix + "TimestampColumn(" + constName + optStr + ")"
	case sqlType == "DATE":
		return prefix + "DateColumn(" + constName + optStr + ")"
	case sqlType == "INTERVAL":
		return prefix + "IntervalColumn(" + constName + optStr + ")"
	case sqlType == "UUID":
		return prefix + "UUIDColumn(" + constName + optStr + ")"
	case sqlType == "JSONB":
		return prefix + "JSONBColumn(" + constName + optStr + ")"
	case sqlType == "JSON":
		return prefix + "JSONColumn(" + constName + optStr + ")"
	case sqlType == "BYTEA":
		return prefix + "ByteaColumn(" + constName + optStr + ")"
	case sqlType == "NUMERIC" || strings.HasPrefix(sqlType, "NUMERIC("):
		return prefix + "BigIntColumn(" + constName + optStr + ")" // fallback
	default:
		return prefix + "TextColumn(" + constName + optStr + ")"
	}
}
