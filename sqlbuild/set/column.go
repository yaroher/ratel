package set

import (
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/sqlbuild/dml/clause"
)

type Column[V any, C types.ColumnAlias] struct {
	fieldAlias C
}

func NewColumn[V any, C types.ColumnAlias](fa C) *Column[V, C] {
	return &Column[V, C]{
		fieldAlias: fa,
	}
}

func (f *Column[V, C]) Set(val V) ValueSetter[C] {
	return NewValueSetter[C](f.fieldAlias, val)
}
func (f *Column[V, C]) SetExpr(expr string) ValueSetter[C] {
	return NewValueSetter[C](f.fieldAlias, expr)
}
func (f *Column[V, C]) SetRaw(sql string, value ...any) ValueSetter[C] {
	return NewValueSetter[C](f.fieldAlias, &clause.RawExprClause[C]{SQL: sql, Args: value})
}
