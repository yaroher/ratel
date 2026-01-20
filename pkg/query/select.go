package query

import (
	"strconv"
	"strings"

	"github.com/yaroher/ratel/pkg/types"
)

type JoinType string

const (
	InnerJoinType JoinType = "INNER JOIN"
	LeftJoinType  JoinType = "LEFT JOIN"
	RightJoinType JoinType = "RIGHT JOIN"
	FullJoinType  JoinType = "FULL JOIN"
)

type JoinClause[C types.ColumnAlias] struct {
	JoinType  JoinType
	Table     string
	Alias     string
	OnClauses []types.Clause[C]
}

type SelectQuery[C types.ColumnAlias] struct {
	BaseQuery[C]
	distinct      bool
	joins         []JoinClause[C]
	whereClauses  []types.Clause[C]
	groupBy       []C
	havingClauses []types.Clause[C]
	orderByASC    []C
	orderByDESC   []C
	orderByRaw    []string
	limit         int
	offset        int
	forUpdate     bool
}

func (q *SelectQuery[F]) Build() (string, []any) {
	i := 1
	// берём буфер из пула
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()

	// heuristics: 128 базовый + ~32 на каждый where/order/group + ~16 на поле
	approxCap := 128 + (len(q.whereClauses)+len(q.orderByASC)+len(q.orderByDESC)+len(q.orderByRaw)+len(q.groupBy))*32 + len(q.UsingFields)*16
	sb.Grow(approxCap)

	args := make([]any, 0, len(q.whereClauses)*2) // простой грубый estimate
	q.AddToBuilder(sb, q.TableAlias(), &i, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

//goland:noinspection t
func (q *SelectQuery[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	inSubquery := buf.Len() > 0 || *paramIndex > 1 || len(*args) > 0
	// ---------- SELECT ----------
	buf.WriteString("SELECT ")
	if q.distinct {
		buf.WriteString("DISTINCT ")
	}
	if len(q.UsingFields) > 0 {
		for i, f := range q.UsingFields {
			//if f.IsCount() {
			//	buf.WriteString("COUNT(")
			//	buf.WriteString(ta)
			//	buf.WriteByte('.')
			//	buf.WriteString(f.String())
			//	buf.WriteString(")")
			//	continue
			//}
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(f.String())
		}
	} else {
		buf.WriteByte('1')
	}

	// ---------- FROM ----------
	buf.WriteString(" FROM ")
	buf.WriteString(ta)
	buf.WriteString(" AS ")
	buf.WriteString(ta)

	// ---------- JOIN ----------
	for _, join := range q.joins {
		buf.WriteByte(' ')
		buf.WriteString(string(join.JoinType))
		buf.WriteByte(' ')
		buf.WriteString(join.Table)
		if join.Alias != "" {
			buf.WriteString(" AS ")
			buf.WriteString(join.Alias)
		}
		if len(join.OnClauses) > 0 {
			buf.WriteString(" ON ")
			for i, clause := range join.OnClauses {
				if i > 0 {
					buf.WriteString(" AND ")
				}
				clause.AddToBuilder(buf, ta, paramIndex, args)
			}
		}
	}

	// ---------- WHERE ----------
	if len(q.whereClauses) > 0 {
		buf.WriteString(" WHERE ")
		for i, clause := range q.whereClauses {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			clause.AddToBuilder(buf, ta, paramIndex, args)
		}
	}

	// ---------- GROUP BY ----------
	if len(q.groupBy) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, g := range q.groupBy {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(g.String())
		}
	}

	// ---------- HAVING ----------
	if len(q.havingClauses) > 0 {
		buf.WriteString(" HAVING ")
		for i, clause := range q.havingClauses {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			clause.AddToBuilder(buf, ta, paramIndex, args)
		}
	}

	// ---------- ORDER BY ----------
	if len(q.orderByASC) > 0 || len(q.orderByDESC) > 0 || len(q.orderByRaw) > 0 {
		buf.WriteString(" ORDER BY ")
		first := true
		for _, f := range q.orderByASC {
			if !first {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(f.String())
			buf.WriteString(" ASC")
			first = false
		}
		for _, f := range q.orderByDESC {
			if !first {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(f.String())
			buf.WriteString(" DESC")
			first = false
		}
		for _, raw := range q.orderByRaw {
			if !first {
				buf.WriteString(", ")
			}
			buf.WriteString(raw)
			first = false
		}
	}

	// ---------- LIMIT / OFFSET ----------
	if q.limit > 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.Itoa(q.limit))
	}
	if q.offset > 0 {
		buf.WriteString(" OFFSET ")
		buf.WriteString(strconv.Itoa(q.offset))
	}

	// ---------- FOR UPDATE ----------
	if q.forUpdate {
		buf.WriteString(" FOR UPDATE")
	}
	if !inSubquery {
		buf.WriteByte(';')
	}
}
func (q *SelectQuery[F]) Fields(
	fields ...F,
) *SelectQuery[F] {
	q.UsingFields = fields
	return q
}
func (q *SelectQuery[F]) Distinct() *SelectQuery[F] {
	q.distinct = true
	return q
}
func (q *SelectQuery[F]) Alias(
	alias string,
) *SelectQuery[F] {
	q.Ta = alias
	return q
}
func (q *SelectQuery[C]) Where(clause ...types.Clause[C]) *SelectQuery[C] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *SelectQuery[C]) GroupBy(fields ...C) *SelectQuery[C] {
	q.groupBy = append(q.groupBy, fields...)
	return q
}
func (q *SelectQuery[C]) Having(clause ...types.Clause[C]) *SelectQuery[C] {
	q.havingClauses = append(q.havingClauses, clause...)
	return q
}
func (q *SelectQuery[C]) OrderByASC(fields ...C) *SelectQuery[C] {
	q.orderByASC = append(q.orderByASC, fields...)
	return q
}
func (q *SelectQuery[C]) OrderByDESC(fields ...C) *SelectQuery[C] {
	q.orderByDESC = append(q.orderByDESC, fields...)
	return q
}
func (q *SelectQuery[C]) OrderByRaw(rawSQL ...string) *SelectQuery[C] {
	q.orderByRaw = append(q.orderByRaw, rawSQL...)
	return q
}
func (q *SelectQuery[C]) Limit(limit int) *SelectQuery[C] {
	q.limit = limit
	return q
}
func (q *SelectQuery[C]) Offset(offset int) *SelectQuery[C] {
	q.offset = offset
	return q
}
func (q *SelectQuery[C]) ForUpdate() *SelectQuery[C] {
	q.forUpdate = true
	return q
}
func (q *SelectQuery[C]) SetLimit(limit int) {
	q.limit = limit
}
func (q *SelectQuery[C]) SetOffset(offset int) {
	q.offset = offset
}
func (q *SelectQuery[C]) SetOrderBy(asc bool, fields ...C) {
	if asc {
		q.orderByASC = append(q.orderByASC, fields...)
	} else {
		q.orderByDESC = append(q.orderByDESC, fields...)
	}
}

// Join adds a generic JOIN clause
func (q *SelectQuery[C]) Join(joinType JoinType, table string, alias string, onClauses ...types.Clause[C]) *SelectQuery[C] {
	q.joins = append(q.joins, JoinClause[C]{
		JoinType:  joinType,
		Table:     table,
		Alias:     alias,
		OnClauses: onClauses,
	})
	return q
}

// InnerJoin adds an INNER JOIN clause
func (q *SelectQuery[C]) InnerJoin(table string, alias string, onClauses ...types.Clause[C]) *SelectQuery[C] {
	return q.Join(InnerJoinType, table, alias, onClauses...)
}

// LeftJoin adds a LEFT JOIN clause
func (q *SelectQuery[C]) LeftJoin(table string, alias string, onClauses ...types.Clause[C]) *SelectQuery[C] {
	return q.Join(LeftJoinType, table, alias, onClauses...)
}

// RightJoin adds a RIGHT JOIN clause
func (q *SelectQuery[C]) RightJoin(table string, alias string, onClauses ...types.Clause[C]) *SelectQuery[C] {
	return q.Join(RightJoinType, table, alias, onClauses...)
}

// FullJoin adds a FULL JOIN clause
func (q *SelectQuery[C]) FullJoin(table string, alias string, onClauses ...types.Clause[C]) *SelectQuery[C] {
	return q.Join(FullJoinType, table, alias, onClauses...)
}
