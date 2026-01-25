package set

import (
	"github.com/yaroher/ratel/pkg/dml/clause"
	"github.com/yaroher/ratel/pkg/types"
)

type Column[V any, C types.ColumnAlias] struct {
	fieldAlias C
}

func NewSetColumn[V any, C types.ColumnAlias](fa C) *Column[V, C] {
	return &Column[V, C]{
		fieldAlias: fa,
	}
}

func (f *Column[V, C]) Set(val V) ValueSetter[C] {
	return NewSetter[C](f.fieldAlias, val)
}
func (f *Column[V, C]) SetExpr(expr string) ValueSetter[C] {
	return NewSetter[C](f.fieldAlias, expr)
}
func (f *Column[V, C]) SetRaw(sql string, value ...any) ValueSetter[C] {
	return NewSetter[C](f.fieldAlias, &clause.RawExprClause[C]{SQL: sql, Args: value})
}
