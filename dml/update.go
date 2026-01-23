package dml

import (
	"strings"

	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/dml/clause"
	"github.com/yaroher/ratel/dml/set"
)

// ---------------------------------------------------------------------------
// UPDATE builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type UpdateQuery[T types.TableAlias, C types.ColumnAlias] struct {
	BaseQuery[T, C]
	setAssigns   []set.ValueSetter[C]
	whereClauses []clause.Clause[C]
}

func (q *UpdateQuery[T, C]) ScanAbleFields() []string {
	if len(q.UsingFields) == 0 {
		return nil
	}
	fields := make([]string, len(q.UsingFields))
	for i, f := range q.UsingFields {
		fields[i] = f.String()
	}
	return fields
}

func (q *UpdateQuery[T, C]) Build() (string, []any) {
	i := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()
	sb.Grow(128 + len(q.setAssigns)*24)

	args := make([]any, 0, len(q.setAssigns)+len(q.whereClauses)*2)
	q.AddToBuilder(sb, q.Ta.String(), &i, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

func (q *UpdateQuery[T, C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("UPDATE ")
	buf.WriteString(ta)
	buf.WriteString(" SET ")
	for i, asg := range q.setAssigns {
		if i > 0 {
			buf.WriteString(", ")
		}
		asg.AddToBuilder(buf, ta, paramIndex, args)
	}

	if len(q.whereClauses) > 0 {
		buf.WriteString(" WHERE ")
		for i, cl := range q.whereClauses {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			cl.AddToBuilder(buf, ta, paramIndex, args)
		}
	}

	q.buildReturning(buf)

	buf.WriteByte(';')
}

func (q *UpdateQuery[T, C]) Where(clause ...clause.Clause[C]) *UpdateQuery[T, C] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *UpdateQuery[T, C]) Set(assign ...set.ValueSetter[C]) *UpdateQuery[T, C] {
	q.setAssigns = append(q.setAssigns, assign...)
	return q
}
func (q *UpdateQuery[T, C]) ReturningAll() *UpdateQuery[T, C] {
	q.UsingFields = q.AllFields
	return q
}
func (q *UpdateQuery[T, C]) Returning(fields ...C) *UpdateQuery[T, C] {
	q.UsingFields = fields
	return q
}
