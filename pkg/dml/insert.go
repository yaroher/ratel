package dml

import (
	"strconv"
	"strings"

	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/types"
)

// ---------------------------------------------------------------------------
// INSERT builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type InsertQuery[T types.TableAlias, C types.ColumnAlias] struct {
	BaseQuery[T, C]

	columns []C
	values  []any // single‑row insert for simplicity; extendable to [][]any

	// ON CONFLICT handling
	conflictTarget []C // columns in UNIQUE/PK to target; empty → global
	doNothing      bool
	updateAssigns  []C // SET … = … when DO UPDATE used
}

func (q *InsertQuery[T, C]) ScanAbleFields() []string {
	if len(q.UsingFields) == 0 {
		return nil
	}
	fields := make([]string, len(q.UsingFields))
	for i, f := range q.UsingFields {
		fields[i] = f.String()
	}
	return fields
}

func (q *InsertQuery[T, C]) Build() (string, []any) {
	idx := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()

	sb.Grow(128 + len(q.columns)*16)

	args := make([]any, 0, len(q.values)+len(q.updateAssigns))
	q.AddToBuilder(sb, q.Ta.String(), &idx, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

//goland:noinspection t
func (q *InsertQuery[T, C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	// INSERT INTO tbl (c1,c2) VALUES ($1,$2)
	buf.WriteString("INSERT INTO ")
	buf.WriteString(q.fromName())
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

func (q *InsertQuery[T, C]) Columns(columns ...C) *InsertQuery[T, C] {
	q.columns = columns
	return q
}
func (q *InsertQuery[T, C]) Values(values ...any) *InsertQuery[T, C] {
	q.values = values
	return q
}
func (q *InsertQuery[T, C]) From(setters ...set.ValueSetter[C]) *InsertQuery[T, C] {
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
func (q *InsertQuery[T, C]) OnConflict(columns ...C) *InsertQuery[T, C] {
	q.conflictTarget = columns
	return q
}
func (q *InsertQuery[T, C]) DoNothing() *InsertQuery[T, C] {
	q.doNothing = true
	return q
}
func (q *InsertQuery[T, C]) DoUpdate(assign ...C) *InsertQuery[T, C] {
	q.updateAssigns = assign
	return q
}
func (q *InsertQuery[T, C]) Returning(fields ...C) *InsertQuery[T, C] {
	q.UsingFields = fields
	return q
}
func (q *InsertQuery[T, C]) ReturningAll() *InsertQuery[T, C] {
	q.UsingFields = q.AllFields
	return q
}
