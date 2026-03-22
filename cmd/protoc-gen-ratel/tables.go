package main

import (
	"github.com/iancoleman/strcase"
	"github.com/yaroher/protoc-gen-go-plain/goplain"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"github.com/yaroher/ratel/ratelproto"
)

// RatelTable represents a table definition for ratel
type RatelTable struct {
	Message   *protogen.Message
	Options   *ratelproto.Table
	Columns   []*RatelColumn
	Relations []*RatelRelation
}

// RatelColumn represents a column in a ratel table
type RatelColumn struct {
	Field      *protogen.Field
	Options    *ratelproto.Column
	SQLName    string
	SQLType    string
	GoType     string
	GoName     string // Override for embedded fields
	IsSkipped  bool
	IsEmbedded bool // True if this column comes from an embedded message
	IsVirtual  bool // True if this is a virtual column (no proto field, DDL-only)
	VirtualDef *ratelproto.VirtualColumn
}

// RatelRelation represents a relation in a ratel table
type RatelRelation struct {
	Field   *protogen.Field
	Options *ratelproto.Relation
}

// collectRatelTables collects tables directly from protogen messages
func collectRatelTables(f *protogen.File) []*RatelTable {
	var tables []*RatelTable

	for _, msg := range f.Messages {
		// Get ratel.table option
		tableOpts := getRatelTableOptions(msg)
		if tableOpts == nil || !tableOpts.Generate {
			continue
		}

		table := &RatelTable{
			Message:   msg,
			Options:   tableOpts,
			Columns:   make([]*RatelColumn, 0),
			Relations: make([]*RatelRelation, 0),
		}

		// Process embedded oneofs
		for _, oneof := range msg.Oneofs {
			if oneof.Desc.IsSynthetic() {
				continue
			}
			if !isEmbeddedOneof(oneof) {
				continue
			}
			cols := collectOneofColumns(oneof)
			table.Columns = append(table.Columns, cols...)
		}

		// Process fields
		for _, field := range msg.Fields {
			// Skip oneof fields (embedded oneofs handled above, non-embedded skipped)
			if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
				continue
			}

			// Check for relation option first
			relOpts := getRatelRelationOptions(field)
			if relOpts != nil {
				table.Relations = append(table.Relations, &RatelRelation{
					Field:   field,
					Options: relOpts,
				})
				continue // Skip adding as column
			}

			// Check for embed option
			if isEmbeddedField(field) && field.Message != nil {
				// Add nested fields from embedded message
				cols := collectEmbeddedColumns(field.Message, "")
				table.Columns = append(table.Columns, cols...)
				continue
			}

			// Get column options
			colOpts := getRatelColumnOptions(field)

			col := &RatelColumn{
				Field:     field,
				Options:   colOpts,
				SQLName:   strcase.ToSnake(string(field.Desc.Name())),
				SQLType:   protoFieldToSQLType(field),
				GoName:    field.GoName,
				IsSkipped: colOpts != nil && colOpts.Skip,
			}
			col.GoType = computeGoType(col)

			if !col.IsSkipped {
				table.Columns = append(table.Columns, col)
			}
		}

		// Add virtual columns from table options
		if tableOpts.VirtualColumns != nil {
			for _, vc := range tableOpts.VirtualColumns {
				col := &RatelColumn{
					SQLName:    vc.SqlName,
					SQLType:    vc.SqlType,
					GoName:     strcase.ToCamel(vc.SqlName),
					IsVirtual:  true,
					VirtualDef: vc,
				}
				table.Columns = append(table.Columns, col)
			}
		}

		tables = append(tables, table)
	}

	return tables
}

// isEmbeddedField checks if a field has (goplain.field).embed = true
func isEmbeddedField(field *protogen.Field) bool {
	if field == nil {
		return false
	}
	opts := field.Desc.Options()
	if opts == nil {
		return false
	}
	ext := proto.GetExtension(opts, goplain.E_Field)
	if ext == nil {
		return false
	}
	fieldOpts, ok := ext.(*goplain.FieldOptions)
	if !ok || fieldOpts == nil {
		return false
	}
	return fieldOpts.Embed
}

// collectEmbeddedColumns collects columns from an embedded message (recursively handles nested embeds)
func collectEmbeddedColumns(msg *protogen.Message, _ string) []*RatelColumn {
	var cols []*RatelColumn

	for _, field := range msg.Fields {
		// Skip oneof fields
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			continue
		}

		// Check for nested embed
		if isEmbeddedField(field) && field.Message != nil {
			// Recursively collect columns from nested embedded message
			nestedCols := collectEmbeddedColumns(field.Message, "")
			cols = append(cols, nestedCols...)
			continue
		}

		// Get ratel column options from the embedded message's field
		colOpts := getRatelColumnOptions(field)

		col := &RatelColumn{
			Field:      field,
			Options:    colOpts,
			SQLName:    strcase.ToSnake(string(field.Desc.Name())),
			SQLType:    protoFieldToSQLType(field),
			GoName:     field.GoName, // Plain struct uses field.GoName directly for embed without prefix
			IsSkipped:  colOpts != nil && colOpts.Skip,
			IsEmbedded: true,
		}
		col.GoType = computeGoType(col)

		if !col.IsSkipped {
			cols = append(cols, col)
		}
	}

	return cols
}

// isEmbeddedOneof checks if a oneof has (goplain.oneof).embed = true
func isEmbeddedOneof(oneof *protogen.Oneof) bool {
	opts := oneof.Desc.Options()
	if opts == nil {
		return false
	}
	ext := proto.GetExtension(opts, goplain.E_Oneof)
	if ext == nil {
		return false
	}
	oneofOpts, ok := ext.(*goplain.OneofOptions)
	if !ok || oneofOpts == nil {
		return false
	}
	return oneofOpts.Embed || oneofOpts.EmbedWithPrefix
}

// collectOneofColumns creates ratel columns for an embedded oneof.
// Each variant field becomes a nullable column, plus a _case TEXT column.
func collectOneofColumns(oneof *protogen.Oneof) []*RatelColumn {
	var cols []*RatelColumn

	oneofName := string(oneof.Desc.Name())

	for _, field := range oneof.Fields {
		variantName := string(field.Desc.Name())

		// Check if field has serialize = true (message stored as bytes)
		fieldOpts := getGoplainFieldOptions(field)
		isSerialized := fieldOpts != nil && fieldOpts.Serialize

		if isSerialized {
			// Serialized message variant → JSONB NULL column (treated as virtual for codegen)
			goName := strcase.ToCamel(variantName) + field.GoName
			sqlName := strcase.ToSnake(variantName) + "_" + strcase.ToSnake(field.GoName)

			cols = append(cols, &RatelColumn{
				SQLName:   sqlName,
				SQLType:   "JSONB",
				GoName:    goName,
				GoType:    "[]byte",
				IsVirtual: true,
				VirtualDef: &ratelproto.VirtualColumn{
					SqlName:    sqlName,
					SqlType:    "JSONB",
					IsNullable: true,
				},
			})
		} else if field.Message != nil {
			// Non-serialized message variant — flatten inner fields as nullable
			for _, innerField := range field.Message.Fields {
				goName := strcase.ToCamel(variantName) + innerField.GoName
				sqlName := strcase.ToSnake(variantName) + "_" + strcase.ToSnake(string(innerField.Desc.Name()))

				col := &RatelColumn{
					Field:      innerField,
					SQLName:    sqlName,
					SQLType:    protoFieldToSQLType(innerField),
					GoName:     goName,
					IsEmbedded: true,
				}
				col.GoType = computeGoType(col)
				cols = append(cols, col)
			}
		} else {
			// Scalar variant → nullable column
			goName := strcase.ToCamel(variantName) + field.GoName
			sqlName := strcase.ToSnake(variantName) + "_" + strcase.ToSnake(string(field.Desc.Name()))

			col := &RatelColumn{
				Field:      field,
				SQLName:    sqlName,
				SQLType:    protoFieldToSQLType(field),
				GoName:     goName,
				IsEmbedded: true,
			}
			col.GoType = computeGoType(col)
			cols = append(cols, col)
		}
	}

	// Add {oneof_name}_case TEXT column (treated as virtual — no proto field)
	caseName := oneofName + "_case"
	cols = append(cols, &RatelColumn{
		SQLName:   caseName,
		SQLType:   "TEXT",
		GoName:    strcase.ToCamel(caseName),
		GoType:    "string",
		IsVirtual: true,
		VirtualDef: &ratelproto.VirtualColumn{
			SqlName: caseName,
			SqlType: "TEXT",
		},
	})

	return cols
}

// getGoplainFieldOptions returns goplain field options for a field
func getGoplainFieldOptions(field *protogen.Field) *goplain.FieldOptions {
	opts := field.Desc.Options()
	if opts == nil {
		return nil
	}
	ext := proto.GetExtension(opts, goplain.E_Field)
	if ext == nil {
		return nil
	}
	fieldOpts, ok := ext.(*goplain.FieldOptions)
	if !ok {
		return nil
	}
	return fieldOpts
}

// getTableName returns the SQL table name for the table (without schema prefix)
func getTableName(table *RatelTable) string {
	if table.Options != nil && table.Options.TableName != nil {
		return *table.Options.TableName
	}
	return strcase.ToSnake(table.Message.GoIdent.GoName)
}

// getTableSchema returns the PostgreSQL schema name (empty = public)
func getTableSchema(table *RatelTable) string {
	if table.Options != nil && table.Options.Schema != nil {
		return *table.Options.Schema
	}
	return ""
}
