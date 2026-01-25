package main

import (
	"github.com/yaroher/protoc-gen-go-plain/goplain"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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

// protoFieldToGoType converts a protobuf field to Go type string
func protoFieldToGoType(field *protogen.Field) string {
	// Check for well-known types
	if field.Message != nil {
		fullName := string(field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			return "time.Time"
		case "google.protobuf.Duration":
			return "time.Duration"
		case "google.protobuf.Int64Value":
			return "*int64"
		case "google.protobuf.Int32Value":
			return "*int32"
		case "google.protobuf.UInt64Value":
			return "*uint64"
		case "google.protobuf.UInt32Value":
			return "*uint32"
		case "google.protobuf.StringValue":
			return "*string"
		case "google.protobuf.BoolValue":
			return "*bool"
		case "google.protobuf.FloatValue":
			return "*float32"
		case "google.protobuf.DoubleValue":
			return "*float64"
		}

		// Check for type_alias messages
		if isTypeAlias(field.Message) {
			// Get the wrapped scalar type
			if len(field.Message.Fields) > 0 {
				valueField := field.Message.Fields[0]
				return protoKindToGoType(valueField.Desc.Kind())
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

// getSchemaColumnType returns the schema.XxxColumnI type for a column
func getSchemaColumnType(col *RatelColumn, msgName string) string {
	isPK := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey
	kind := col.Field.Desc.Kind()

	// Check for type_alias messages first
	if col.Field.Message != nil && isTypeAlias(col.Field.Message) {
		if len(col.Field.Message.Fields) > 0 {
			valueField := col.Field.Message.Fields[0]
			kind = valueField.Desc.Kind()
		}
	}

	// For int64 primary keys, use BigSerialColumn
	if isPK && kind == protoreflect.Int64Kind {
		return "schema.BigSerialColumnI[" + msgName + "ColumnAlias]"
	}

	// For string primary keys
	if isPK && kind == protoreflect.StringKind {
		return "schema.TextColumnI[" + msgName + "ColumnAlias]"
	}

	// Check for well-known types
	if col.Field.Message != nil && !isTypeAlias(col.Field.Message) {
		fullName := string(col.Field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			return "schema.TimestamptzColumnI[" + msgName + "ColumnAlias]"
		case "google.protobuf.Duration":
			return "schema.IntervalColumnI[" + msgName + "ColumnAlias]"
		case "google.protobuf.Int64Value":
			return "schema.NullBigIntColumnI[" + msgName + "ColumnAlias]"
		case "google.protobuf.Int32Value":
			return "schema.NullIntegerColumnI[" + msgName + "ColumnAlias]"
		case "google.protobuf.StringValue":
			return "schema.NullTextColumnI[" + msgName + "ColumnAlias]"
		case "google.protobuf.BoolValue":
			return "schema.NullBooleanColumnI[" + msgName + "ColumnAlias]"
		case "google.protobuf.Struct":
			return "schema.JSONColumnI[" + msgName + "ColumnAlias]"
		}
	}

	// For other types
	switch kind {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "schema.IntegerColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "schema.BigIntColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "schema.IntegerColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "schema.BigIntColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.StringKind:
		return "schema.TextColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.BoolKind:
		return "schema.BooleanColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.FloatKind:
		return "schema.RealColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.DoubleKind:
		return "schema.DoublePrecisionColumnI[" + msgName + "ColumnAlias]"
	case protoreflect.BytesKind:
		return "schema.ByteaColumnI[" + msgName + "ColumnAlias]"
	default:
		return "schema.TextColumnI[" + msgName + "ColumnAlias]"
	}
}

// getSchemaColumnConstructor returns the schema.XxxColumn constructor call for a column
func getSchemaColumnConstructor(col *RatelColumn, constName string) string {
	isPK := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey
	isUnique := col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.Unique
	defaultVal := ""
	if col.Options != nil && col.Options.Constraints != nil {
		defaultVal = col.Options.Constraints.DefaultValue
	}

	kind := col.Field.Desc.Kind()

	// Check for type_alias messages
	if col.Field.Message != nil && isTypeAlias(col.Field.Message) {
		if len(col.Field.Message.Fields) > 0 {
			valueField := col.Field.Message.Fields[0]
			kind = valueField.Desc.Kind()
		}
	}

	// Build options
	var opts []string
	if isPK {
		opts = append(opts, "ddl.WithPrimaryKey["+col.Field.Parent.GoIdent.GoName+"ColumnAlias]()")
	}
	if isUnique {
		opts = append(opts, "ddl.WithUnique["+col.Field.Parent.GoIdent.GoName+"ColumnAlias]()")
	}
	if defaultVal != "" {
		opts = append(opts, "ddl.WithDefault["+col.Field.Parent.GoIdent.GoName+"ColumnAlias](\""+defaultVal+"\")")
	}

	optStr := ""
	if len(opts) > 0 {
		for _, opt := range opts {
			optStr += ", " + opt
		}
	}

	// Check for well-known types
	if col.Field.Message != nil {
		fullName := string(col.Field.Message.Desc.FullName())
		switch fullName {
		case "google.protobuf.Timestamp":
			return "schema.TimestamptzColumn(" + constName + optStr + ")"
		case "google.protobuf.Duration":
			return "schema.IntervalColumn(" + constName + optStr + ")"
		case "google.protobuf.Int64Value":
			return "schema.NullBigIntColumn(" + constName + optStr + ")"
		case "google.protobuf.Int32Value":
			return "schema.NullIntegerColumn(" + constName + optStr + ")"
		case "google.protobuf.StringValue":
			return "schema.NullTextColumn(" + constName + optStr + ")"
		case "google.protobuf.BoolValue":
			return "schema.NullBooleanColumn(" + constName + optStr + ")"
		}
	}

	// For int64 primary keys, use BigSerialColumn
	if isPK && kind == protoreflect.Int64Kind {
		return "schema.BigSerialColumn(" + constName + optStr + ")"
	}

	switch kind {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "schema.IntegerColumn(" + constName + optStr + ")"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "schema.BigIntColumn(" + constName + optStr + ")"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "schema.IntegerColumn(" + constName + optStr + ")"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "schema.BigIntColumn(" + constName + optStr + ")"
	case protoreflect.StringKind:
		return "schema.TextColumn(" + constName + optStr + ")"
	case protoreflect.BoolKind:
		return "schema.BooleanColumn(" + constName + optStr + ")"
	case protoreflect.FloatKind:
		return "schema.RealColumn(" + constName + optStr + ")"
	case protoreflect.DoubleKind:
		return "schema.DoublePrecisionColumn(" + constName + optStr + ")"
	case protoreflect.BytesKind:
		return "schema.ByteaColumn(" + constName + optStr + ")"
	default:
		return "schema.TextColumn(" + constName + optStr + ")"
	}
}
