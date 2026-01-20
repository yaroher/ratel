package query

import (
	"strings"

	"github.com/yaroher/ratel/pkg/types"
)

// ---------------------------------------------------------------------------
// UPDATE builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type UpdateQuery[C types.ColumnAlias] struct {
	BaseQuery[C]
	setAssigns   []types.ValueSetter[C]
	whereClauses []types.Clause[C]
}

func (q *UpdateQuery[F]) mustOrmQuery() {}

func (q *UpdateQuery[F]) Build() (string, []any) {
	i := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()
	sb.Grow(128 + len(q.setAssigns)*24)

	args := make([]any, 0, len(q.setAssigns)+len(q.whereClauses)*2)
	q.AddToBuilder(sb, q.TableAlias(), &i, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

func (q *UpdateQuery[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
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

func (q *UpdateQuery[C]) Where(clause ...types.Clause[C]) *UpdateQuery[C] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *UpdateQuery[C]) Set(assign ...types.ValueSetter[C]) *UpdateQuery[C] {
	q.setAssigns = append(q.setAssigns, assign...)
	return q
}
func (q *UpdateQuery[C]) ReturningAll() *UpdateQuery[C] {
	q.UsingFields = q.AllFields
	return q
}
func (q *UpdateQuery[C]) Returning(fields ...C) *UpdateQuery[C] {
	q.UsingFields = fields
	return q
}
