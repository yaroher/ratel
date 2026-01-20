package clause

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/yaroher/ratel/pkg/types"
)

// renumberPlaceholders — renumber $placeholders when embedding sub‑queries
func renumberPlaceholders(sql string, paramIndex *int) string {
	var b strings.Builder
	for i := 0; i < len(sql); {
		if sql[i] == '$' {
			j := i + 1
			for j < len(sql) && sql[j] >= '0' && sql[j] <= '9' {
				j++
			}
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(*paramIndex))
			*paramIndex++
			i = j
		} else {
			b.WriteByte(sql[i])
			i++
		}
	}
	return b.String()
}

type ParamExprClause[C types.ColumnAlias] struct {
	Value     any
	LikeWrapp bool
}

func (e *ParamExprClause[C]) mustClauseAlias(_ C) {}
func (e *ParamExprClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if e.LikeWrapp {
		buf.WriteString("'%' || ")
	}
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(*paramIndex))
	*paramIndex++
	if e.LikeWrapp {
		buf.WriteString("::text || '%'")
	}
	*args = append(*args, e.Value)
}

type SliceExprClause[C types.ColumnAlias] struct {
	Values any // slice
}

func (e *SliceExprClause[C]) mustClauseAlias(_ C) {}
func (e *SliceExprClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	rv := reflect.ValueOf(e.Values)
	if rv.Kind() != reflect.Slice {
		panic("SliceExprClause expects slice")
	}
	buf.WriteByte('(')
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(*paramIndex))
	*paramIndex++
	*args = append(*args, e.Values)
	buf.WriteByte(')')
}

type RawExprClause[C types.ColumnAlias] struct {
	SQL  string
	Args []any
}

func (e *RawExprClause[C]) mustClauseAlias(_ C) {}
func (e *RawExprClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	parts := strings.Split(e.SQL, "?")
	for i, part := range parts {
		if i > 0 {
			buf.WriteByte('$')
			buf.WriteString(strconv.Itoa(*paramIndex))
			*paramIndex++
		}
		buf.WriteString(part)
	}
	*args = append(*args, e.Args...)
}

type SubQueryExprClause[C types.ColumnAlias] struct {
	Query types.OrmQuery
}

func (e *SubQueryExprClause[C]) mustClauseAlias(_ C) {}
func (e *SubQueryExprClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	//sql, subArgs := e.Query.Build()
	//sql = renumberPlaceholders(sql, paramIndex)
	//buf.WriteByte('(')
	//buf.WriteString(sql)
	//buf.WriteByte(')')
	//*args = append(*args, subArgs...)
	buf.WriteByte('(')
	e.Query.AddToBuilder(buf, e.Query.TableAlias(), paramIndex, args)
	buf.WriteByte(')')

}

type FieldClause[C types.ColumnAlias] struct {
	Field    C
	Operator string
	Right    types.BufferBuilder
	Negate   bool
}

func (c *FieldClause[C]) mustClauseAlias(_ C) {}
func (c *FieldClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if c.Negate {
		buf.WriteString("NOT (")
	}
	buf.WriteString(ta)
	buf.WriteByte('.')
	buf.WriteString(c.Field.String())
	buf.WriteByte(' ')
	buf.WriteString(c.Operator)
	buf.WriteByte(' ')
	c.Right.AddToBuilder(buf, ta, paramIndex, args)
	if c.Negate {
		buf.WriteByte(')')
	}
}

type AndClause[C types.ColumnAlias] struct {
	Clauses []types.Clause[C]
}

func (c *AndClause[C]) mustClauseAlias(_ C) {}
func (c *AndClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteByte('(')
	for i, cl := range c.Clauses {
		if i > 0 {
			buf.WriteString(" AND ")
		}
		cl.AddToBuilder(buf, ta, paramIndex, args)
	}
	buf.WriteByte(')')
}

type OrClause[C types.ColumnAlias] struct {
	Clauses []types.Clause[C]
}

func (c *OrClause[C]) mustClauseAlias(_ C) {}
func (c *OrClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteByte('(')
	for i, cl := range c.Clauses {
		if i > 0 {
			buf.WriteString(" OR ")
		}
		cl.AddToBuilder(buf, ta, paramIndex, args)
	}
	buf.WriteByte(')')
}

type NotClause[C types.ColumnAlias] struct {
	Inner types.Clause[C]
}

func (c *NotClause[C]) mustClauseAlias(_ C) {}
func (c *NotClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("NOT (")
	c.Inner.AddToBuilder(buf, ta, paramIndex, args)
	buf.WriteByte(')')
}

type ExistsClause[C types.ColumnAlias] struct {
	SubQuery types.Clause[C]
	Negate   bool
}

func (c *ExistsClause[C]) mustClauseAlias(_ C) {}
func (c *ExistsClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if c.Negate {
		buf.WriteString("NOT ")
	}
	buf.WriteString("EXISTS ")
	c.SubQuery.AddToBuilder(buf, ta, paramIndex, args)
}
