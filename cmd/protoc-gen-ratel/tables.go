package main

import (
	"github.com/iancoleman/strcase"
	"github.com/yaroher/protoc-gen-go-plain/goplain"
	"github.com/yaroher/ratel/ratelproto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
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

		// Process fields
		for _, field := range msg.Fields {
			// Skip oneof fields (handled separately)
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
				GoType:    protoFieldToGoType(field),
				GoName:    field.GoName,
				IsSkipped: colOpts != nil && colOpts.Skip,
			}

			if !col.IsSkipped {
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
func collectEmbeddedColumns(msg *protogen.Message, prefix string) []*RatelColumn {
	var cols []*RatelColumn

	for _, field := range msg.Fields {
		// Skip oneof fields
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			continue
		}

		// Check for nested embed
		if isEmbeddedField(field) && field.Message != nil {
			// Recursively collect columns from nested embedded message
			nestedCols := collectEmbeddedColumns(field.Message, prefix)
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
			GoType:     protoFieldToGoType(field),
			GoName:     field.GoName, // Plain struct uses field.GoName directly for embed without prefix
			IsSkipped:  colOpts != nil && colOpts.Skip,
			IsEmbedded: true,
		}

		if !col.IsSkipped {
			cols = append(cols, col)
		}
	}

	return cols
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
