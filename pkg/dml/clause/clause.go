package clause

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/yaroher/ratel/pkg/types"
)

type Clause[C types.ColumnAlias] interface {
	AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any)
}

type ParamExprClause[C types.ColumnAlias] struct {
	Value     any
	LikeWrapp bool
}

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
	Query types.Query
}

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

// ColumnRefExpr writes a qualified column reference (e.g. "users.id") without parameters.
// Used for correlated subqueries where the inner query references an outer table's column.
type ColumnRefExpr struct {
	TableAlias string
	Column     string
}

func (e *ColumnRefExpr) AddToBuilder(buf *strings.Builder, _ string, _ *int, _ *[]any) {
	buf.WriteString(e.TableAlias)
	buf.WriteByte('.')
	buf.WriteString(e.Column)
}

type FieldClause[C types.ColumnAlias] struct {
	Field    C
	Operator string
	Right    types.Builder
	Negate   bool
}

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
	Clauses []Clause[C]
}

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
	Clauses []Clause[C]
}

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
	Inner Clause[C]
}

func (c *NotClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("NOT (")
	c.Inner.AddToBuilder(buf, ta, paramIndex, args)
	buf.WriteByte(')')
}

type ExistsClause[C types.ColumnAlias] struct {
	SubQuery Clause[C]
	Negate   bool
}

func (c *ExistsClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if c.Negate {
		buf.WriteString("NOT ")
	}
	buf.WriteString("EXISTS ")
	c.SubQuery.AddToBuilder(buf, ta, paramIndex, args)
}

type JsonFieldAccess[C types.ColumnAlias] struct {
	base     C
	operator string
	key      string
	path     []string
}

func (j *JsonFieldAccess[C]) Eq(val any) Clause[C] {
	return &JsonFieldClause[C]{
		base:         j.base,
		accessOp:     j.operator,
		key:          j.key,
		path:         j.path,
		compareOp:    "=",
		compareValue: val,
	}
}
func (j *JsonFieldAccess[C]) Neq(val any) Clause[C] {
	return &JsonFieldClause[C]{
		base:         j.base,
		accessOp:     j.operator,
		key:          j.key,
		path:         j.path,
		compareOp:    "!=",
		compareValue: val,
	}
}
func (j *JsonFieldAccess[C]) Gt(val any) Clause[C] {
	return &JsonFieldClause[C]{
		base:         j.base,
		accessOp:     j.operator,
		key:          j.key,
		path:         j.path,
		compareOp:    ">",
		compareValue: val,
	}
}
func (j *JsonFieldAccess[C]) Lt(val any) Clause[C] {
	return &JsonFieldClause[C]{
		base:         j.base,
		accessOp:     j.operator,
		key:          j.key,
		path:         j.path,
		compareOp:    "<",
		compareValue: val,
	}
}
func (j *JsonFieldAccess[C]) In(vals ...any) Clause[C] {
	return &JsonFieldClause[C]{
		base:         j.base,
		accessOp:     j.operator,
		key:          j.key,
		path:         j.path,
		compareOp:    "IN",
		compareValue: vals,
	}
}
func (j *JsonFieldAccess[C]) Like(pattern string) Clause[C] {
	return &JsonFieldClause[C]{
		base:         j.base,
		accessOp:     j.operator,
		key:          j.key,
		path:         j.path,
		compareOp:    "LIKE",
		compareValue: pattern,
		likeWrapp:    true,
	}
}

type JsonFieldClause[C types.ColumnAlias] struct {
	base         C
	accessOp     string
	key          string
	path         []string
	compareOp    string
	compareValue any
	likeWrapp    bool
}

func (c *JsonFieldClause[C]) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	// Build field access: ta.field->>'key' or ta.field#>>'{path,to,field}'
	buf.WriteString(ta)
	buf.WriteByte('.')
	buf.WriteString(c.base.String())
	buf.WriteString(c.accessOp)

	if len(c.path) > 0 {
		// For path operators (#>, #>>)
		buf.WriteString("'{")
		for i, p := range c.path {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(p)
		}
		buf.WriteString("}'")
	} else {
		// For key operators (->, ->>)
		buf.WriteByte('\'')
		buf.WriteString(c.key)
		buf.WriteByte('\'')
	}

	buf.WriteByte(' ')
	buf.WriteString(c.compareOp)
	buf.WriteByte(' ')

	// Add comparison value
	if c.compareOp == "IN" {
		buf.WriteByte('(')
		buf.WriteByte('$')
		buf.WriteString(strconv.Itoa(*paramIndex))
		*paramIndex++
		*args = append(*args, c.compareValue)
		buf.WriteByte(')')
	} else {
		if c.likeWrapp {
			buf.WriteString("'%' || ")
		}
		buf.WriteByte('$')
		buf.WriteString(strconv.Itoa(*paramIndex))
		*paramIndex++
		if c.likeWrapp {
			buf.WriteString(" || '%'")
		}
		*args = append(*args, c.compareValue)
	}
}
