package dml

import (
	"strconv"
	"strings"

	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/sqlbuild/set"
)

// ---------------------------------------------------------------------------
// INSERT builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type InsertQuery[C types.ColumnAlias] struct {
	BaseQuery[C]

	columns []C
	values  []any // single‑row insert for simplicity; extendable to [][]any

	// ON CONFLICT handling
	conflictTarget []C // columns in UNIQUE/PK to target; empty → global
	doNothing      bool
	updateAssigns  []C // SET … = … when DO UPDATE used
}

func (q *InsertQuery[C]) ScanAbleFields() []string {
	//TODO implement me
	panic("implement me")
}

func (q *InsertQuery[C]) Build() (string, []any) {
	idx := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()

	sb.Grow(128 + len(q.columns)*16)

	args := make([]any, 0, len(q.values)+len(q.updateAssigns))
	q.AddToBuilder(sb, q.TableAlias(), &idx, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

//goland:noinspection t
func (q *InsertQuery[F]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	// INSERT INTO tbl (c1,c2) VALUES ($1,$2)
	buf.WriteString("INSERT INTO ")
	buf.WriteString(ta)
	if len(q.columns) > 0 {
		buf.WriteString(" (")
		for i, c := range q.columns {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(c.String())
		}
		buf.WriteByte(')')
	}

	// VALUES
	buf.WriteString(" VALUES (")
	for i, v := range q.values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteByte('$')
		buf.WriteString(strconv.Itoa(*paramIndex))
		*paramIndex++
		vals := *args
		vals = append(vals, v)
		*args = vals
	}
	buf.WriteByte(')')

	// ON CONFLICT
	if q.doNothing || len(q.updateAssigns) > 0 {
		buf.WriteString(" ON CONFLICT")
		if len(q.conflictTarget) > 0 {
			buf.WriteString(" (")
			for i, c := range q.conflictTarget {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(c.String())
			}
			buf.WriteByte(')')
		}
		if q.doNothing {
			buf.WriteString(" DO NOTHING")
		} else {
			buf.WriteString(" DO UPDATE SET ")
			for i, asg := range q.updateAssigns {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(asg.String() + "=EXCLUDED." + asg.String())
			}
		}
	}

	// RETURNING
	q.buildReturning(buf)
	buf.WriteByte(';')
}

func (q *InsertQuery[C]) Columns(columns ...C) *InsertQuery[C] {
	q.columns = columns
	return q
}
func (q *InsertQuery[C]) Values(values ...any) *InsertQuery[C] {
	q.values = values
	return q
}
func (q *InsertQuery[C]) From(setters ...set.ValueSetter[C]) *InsertQuery[C] {
	cols := make([]C, 0, len(setters))
	vals := make([]any, 0, len(setters))
	for _, setter := range setters {
		cols = append(cols, setter.Column())
		vals = append(vals, setter.Value())
	}
	q.Columns(cols...)
	q.Values(vals...)
	return q
}
func (q *InsertQuery[F]) OnConflict(columns ...F) *InsertQuery[F] {
	q.conflictTarget = columns
	return q
}
func (q *InsertQuery[F]) DoNothing() *InsertQuery[F] {
	q.doNothing = true
	return q
}
func (q *InsertQuery[F]) DoUpdate(assign ...F) *InsertQuery[F] {
	q.updateAssigns = assign
	return q
}
func (q *InsertQuery[F]) Returning(fields ...F) *InsertQuery[F] {
	q.UsingFields = fields
	return q
}
func (q *InsertQuery[F]) ReturningAll() *InsertQuery[F] {
	q.UsingFields = q.AllFields
	return q
}
