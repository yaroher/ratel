package schema

import (
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/sqlbuild/ddl"
	"github.com/yaroher/ratel/sqlbuild/dml/clause"
	"github.com/yaroher/ratel/sqlbuild/set"
)

type Column[V any, C types.ColumnAlias] struct {
	def ddl.ColumnDef[C]
	ops *clause.OperandColumn[V, C]
	set set.SetterColumn[V, C]
}

func (c *Column[V, C]) ability() {}

func newColumn[V any, C types.ColumnAlias](def ddl.ColumnDef[C]) any {
	return &Column[V, C]{
		def: def,
		set: set.NewColumn[V, C](def.Alias()),
		ops: clause.NewColumn[V, C](def.Alias()),
	}
}

func NewColumn[V any, C types.ColumnAlias, A collType](def ddl.ColumnDef[C]) A {
	return newColumn[V, C](def).(A)
}

func IntergerColumn(def ddl.ColumnDef[types.ColumnAlias]) interface{} {

}
