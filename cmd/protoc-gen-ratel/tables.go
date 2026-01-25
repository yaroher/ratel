package main

import (
	"github.com/iancoleman/strcase"
	"github.com/yaroher/ratel/ratelproto"
	"google.golang.org/protobuf/compiler/protogen"
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
	Field     *protogen.Field
	Options   *ratelproto.Column
	SQLName   string
	SQLType   string
	GoType    string
	IsSkipped bool
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

			// Get column options
			colOpts := getRatelColumnOptions(field)

			col := &RatelColumn{
				Field:     field,
				Options:   colOpts,
				SQLName:   strcase.ToSnake(string(field.Desc.Name())),
				SQLType:   protoFieldToSQLType(field),
				GoType:    protoFieldToGoType(field),
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

// getTableName returns the SQL table name for the table
func getTableName(table *RatelTable) string {
	if table.Options != nil && table.Options.TableName != nil {
		return *table.Options.TableName
	}
	return strcase.ToSnake(table.Message.GoIdent.GoName)
}
