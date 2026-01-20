package query

import (
	"strings"

	"github.com/yaroher/ratel/pkg/types"
)

// ---------------------------------------------------------------------------
// DELETE builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type DeleteQuery[C types.ColumnAlias] struct {
	BaseQuery[C]
	whereClauses []types.Clause[C]
}

func (q *DeleteQuery[C]) mustOrmQuery() {}

func (q *DeleteQuery[C]) Build() (string, []any) {
	idx := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()
	sb.Grow(96)

	args := make([]any, 0, len(q.whereClauses)*2)
	q.AddToBuilder(sb, q.TableAlias(), &idx, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

func (q *DeleteQuery[C]) AddToBuilder(sb *strings.Builder, ta string, paramIndex *int, args *[]any) {
	sb.WriteString("DELETE FROM ")
	sb.WriteString(ta)

	if len(q.whereClauses) > 0 {
		sb.WriteString(" WHERE ")
		for i, cl := range q.whereClauses {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			cl.AddToBuilder(sb, ta, paramIndex, args)
		}
	}
	q.buildReturning(sb)
	sb.WriteByte(';')
}

func (q *DeleteQuery[C]) Where(clause ...types.Clause[C]) *DeleteQuery[C] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *DeleteQuery[C]) Returning(fields ...C) *DeleteQuery[C] {
	q.UsingFields = fields
	return q
}
func (q *DeleteQuery[C]) ReturningAll() *DeleteQuery[C] {
	q.UsingFields = q.AllFields
	return q
}
