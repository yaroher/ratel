package schema

import (
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/ddl"
	"github.com/yaroher/ratel/dml/clause"
	"github.com/yaroher/ratel/set"
)

type Column[V any, C types.ColumnAlias] struct {
	def *ddl.ColumnDef[C]
	*clause.OperandColumn[V, C]
	*set.Column[V, C]
}

func (c *Column[V, C]) ability() {}

// Def returns the DDL column definition
func (c *Column[V, C]) Def() *ddl.ColumnDef[C] {
	return c.def
}

// Alias returns the column alias
func (c *Column[V, C]) Alias() C {
	return c.def.Alias()
}

func NewColumn[V any, C types.ColumnAlias](def *ddl.ColumnDef[C]) *Column[V, C] {
	return &Column[V, C]{
		def:           def,
		Column:        set.NewColumn[V, C](def.Alias()),
		OperandColumn: clause.NewColumn[V, C](def.Alias()),
	}
}

// ArrayColumn represents an array column with array-specific operations
type ArrayColumn[V any, C types.ColumnAlias] struct {
	def *ddl.ColumnDef[C]
	*clause.ArrayOperandColumn[V, C]
	*set.Column[[]V, C]
}

func (c *ArrayColumn[V, C]) ability() {}

// Def returns the DDL column definition
func (c *ArrayColumn[V, C]) Def() *ddl.ColumnDef[C] {
	return c.def
}

// Alias returns the column alias
func (c *ArrayColumn[V, C]) Alias() C {
	return c.def.Alias()
}

// NewArrayColumn creates a new array column with array-specific operations
func NewArrayColumn[V any, C types.ColumnAlias](def *ddl.ColumnDef[C]) *ArrayColumn[V, C] {
	return &ArrayColumn[V, C]{
		def:                def,
		Column:             set.NewColumn[[]V, C](def.Alias()),
		ArrayOperandColumn: clause.NewArrayColumn[V, C](def.Alias()),
	}
}

// JsonColumn represents a JSON/JSONB column with JSON-specific operations
type JsonColumn[C types.ColumnAlias] struct {
	def *ddl.ColumnDef[C]
	*clause.JsonOperandColumn[C]
	*set.Column[string, C] // JSON stored as string
}

func (c *JsonColumn[C]) ability() {}

// Def returns the DDL column definition
func (c *JsonColumn[C]) Def() *ddl.ColumnDef[C] {
	return c.def
}

// Alias returns the column alias
func (c *JsonColumn[C]) Alias() C {
	return c.def.Alias()
}

// NewJsonColumn creates a new JSON column with JSON-specific operations
func NewJsonColumn[C types.ColumnAlias](def *ddl.ColumnDef[C]) *JsonColumn[C] {
	return &JsonColumn[C]{
		def:               def,
		Column:            set.NewColumn[string, C](def.Alias()),
		JsonOperandColumn: clause.NewJsonColumn[C](def.Alias()),
	}
}

//
//// Column factory functions for common types
//
//// IntegerCol creates an INTEGER column from an alias
//func IntegerCol[C types.ColumnAlias](alias C) *Column[int32, C] {
//	return NewColumn[int32, C](ddl.Column(alias, ddl.Integer()))
//}
//
//// BigIntCol creates a BIGINT column from an alias
//func BigIntCol[C types.ColumnAlias](alias C) *Column[int64, C] {
//	return NewColumn[int64, C](ddl.Column(alias, ddl.BigInt()))
//}
//
//// TextCol creates a TEXT column from an alias
//func TextCol[C types.ColumnAlias](alias C) *Column[string, C] {
//	return NewColumn[string, C](ddl.Column(alias, ddl.Text()))
//}
//
//// BoolCol creates a BOOLEAN column from an alias
//func BoolCol[C types.ColumnAlias](alias C) *Column[bool, C] {
//	return NewColumn[bool, C](ddl.Column(alias, ddl.Boolean()))
//}
//
//// IntergerColumn is deprecated - use IntegerCol instead
//func IntergerColumn[C types.ColumnAlias](def *ddl.ColumnDef[C]) *Column[int32, C] {
//	return NewColumn[int32, C](def)
//}
