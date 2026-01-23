package dml

import (
	"strings"

	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/dml/clause"
)

// ---------------------------------------------------------------------------
// DELETE builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type DeleteQuery[T types.TableAlias, C types.ColumnAlias] struct {
	BaseQuery[T, C]
	whereClauses []clause.Clause[C]
}

func (q *DeleteQuery[T, C]) ScanAbleFields() []string {
	//TODO implement me
	panic("implement me")
}

func (q *DeleteQuery[T, C]) Build() (string, []any) {
	idx := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()
	sb.Grow(96)

	args := make([]any, 0, len(q.whereClauses)*2)
	q.AddToBuilder(sb, q.Ta.String(), &idx, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}
func (q *DeleteQuery[T, C]) AddToBuilder(sb *strings.Builder, ta string, paramIndex *int, args *[]any) {
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

func (q *DeleteQuery[T, C]) Where(clause ...clause.Clause[C]) *DeleteQuery[T, C] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *DeleteQuery[T, C]) Returning(fields ...C) *DeleteQuery[T, C] {
	q.UsingFields = fields
	return q
}
func (q *DeleteQuery[T, C]) ReturningAll() *DeleteQuery[T, C] {
	q.UsingFields = q.AllFields
	return q
}
