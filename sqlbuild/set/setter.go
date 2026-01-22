package set

import (
	"strconv"
	"strings"

	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/sqlbuild/dml/clause"
)

type SetterColumn[V any, C types.ColumnAlias] interface {
	SetExpr(string) ValueSetter[C]
	Set(V) ValueSetter[C]
	SetRaw(sql string, value ...any) ValueSetter[C]
}

type ValueSetter[C types.ColumnAlias] interface {
	types.Builder
	Value() any
	Column() C
}

type ValueSetterImpl[C types.ColumnAlias] struct {
	field C
	value any
	expr  string
	raw   *clause.RawExprClause[C]
}

func NewValueSetter[C types.ColumnAlias](field C, value any) *ValueSetterImpl[C] {
	return &ValueSetterImpl[C]{field: field, value: value}
}

func (s ValueSetterImpl[C]) Column() C {
	return s.field
}

func (s ValueSetterImpl[C]) Value() any {
	if s.expr != "" {
		return nil
	}
	return s.value
}

func (s ValueSetterImpl[F]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if s.raw != nil {
		s.raw.AddToBuilder(buf, ta, paramIndex, args)
	} else {
		buf.WriteString(s.field.String())
		buf.WriteString(" = ")
		if s.expr != "" {
			buf.WriteString(s.expr)
			return
		}
		buf.WriteByte('$')
		buf.WriteString(strconv.Itoa(*paramIndex))
		*paramIndex++
		*args = append(*args, s.value)
	}
}
