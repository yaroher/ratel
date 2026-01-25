package ddl

import (
	"strings"

	"github.com/yaroher/ratel/common/types"
)

type TableDDL[T types.TableAlias, C types.ColumnAlias] struct {
	alias   T
	columns []*ColumnDDL[C]
	indexes []*Index[T, C]
	unique  [][]C
}

type TableOptions[T types.TableAlias, C types.ColumnAlias] func(*TableDDL[T, C])

func WithIndexes[T types.TableAlias, C types.ColumnAlias](indexes ...*Index[T, C]) TableOptions[T, C] {
	return func(ddl *TableDDL[T, C]) {
		ddl.indexes = append(ddl.indexes, indexes...)
	}
}

func WithUniqueColumns[T types.TableAlias, C types.ColumnAlias](columns ...[]C) TableOptions[T, C] {
	return func(ddl *TableDDL[T, C]) {
		ddl.unique = append(ddl.unique, columns...)
	}
}

func NewTableDDL[T types.TableAlias, C types.ColumnAlias](
	alias T,
	columns []*ColumnDDL[C],
	options ...TableOptions[T, C],
) *TableDDL[T, C] {
	result := &TableDDL[T, C]{
		alias:   alias,
		columns: columns,
	}
	for _, opt := range options {
		opt(result)
	}
	return result
}

func (c *TableDDL[T, C]) Indexes(indexes ...*Index[T, C]) *TableDDL[T, C] {
	c.indexes = append(c.indexes, indexes...)
	return c
}

func (c *TableDDL[T, C]) Unique(columns ...[]C) *TableDDL[T, C] {
	c.unique = append(c.unique, columns...)
	return c
}

func (c *TableDDL[T, C]) Alias() T {
	return c.alias
}

func (c *TableDDL[T, C]) SchemaSql() []string {
	var statements []string
	var sql strings.Builder

	// CREATE TABLE IF NOT EXISTS
	sql.WriteString("CREATE TABLE IF NOT EXISTS ")
	sql.WriteString(c.alias.String())
	sql.WriteString(" (\n")

	// Add columns
	for i, col := range c.columns {
		if i > 0 {
			sql.WriteString(",\n")
		}
		sql.WriteString("  ")
		sql.WriteString(col.SchemaSql())
	}

	// Add table-level unique constraints
	for _, uniqueCols := range c.unique {
		sql.WriteString(",\n  UNIQUE (")
		for i, col := range uniqueCols {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(col.String())
		}
		sql.WriteString(")")
	}

	sql.WriteString("\n)")

	// Add CREATE TABLE statement
	statements = append(statements, sql.String())

	// Add indexes as separate statements
	for _, idx := range c.indexes {
		statements = append(statements, idx.SchemaSql())
	}

	return statements
}
