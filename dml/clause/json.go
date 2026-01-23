package clause

import (
	"strconv"
	"strings"

	"github.com/yaroher/ratel/common/types"
)

// JsonOperandColumn provides PostgreSQL JSON/JSONB operations
type JsonOperandColumn[C types.ColumnAlias] struct {
	fieldAlias C
}

func NewJsonColumn[C types.ColumnAlias](fa C) *JsonOperandColumn[C] {
	return &JsonOperandColumn[C]{
		fieldAlias: fa,
	}
}

// GetField (->) - get JSON object field
func (f *JsonOperandColumn[C]) GetField(key string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "->",
		key:      key,
	}
}

// GetFieldText (->>) - get JSON field as text
func (f *JsonOperandColumn[C]) GetFieldText(key string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "->>",
		key:      key,
	}
}

// GetPath (#>) - get nested JSON object
func (f *JsonOperandColumn[C]) GetPath(path ...string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "#>",
		path:     path,
	}
}

// GetPathText (#>>) - get nested JSON object as text
func (f *JsonOperandColumn[C]) GetPathText(path ...string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "#>>",
		path:     path,
	}
}

// Contains (@>) - JSON contains another JSON
func (f *JsonOperandColumn[C]) Contains(jsonValue string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@>",
		Right:    &ParamExprClause[C]{Value: jsonValue},
	}
}

// ContainedBy (<@) - JSON is contained by another JSON
func (f *JsonOperandColumn[C]) ContainedBy(jsonValue string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "<@",
		Right:    &ParamExprClause[C]{Value: jsonValue},
	}
}

// HasKey (?) - JSON has key
func (f *JsonOperandColumn[C]) HasKey(key string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "?",
		Right:    &ParamExprClause[C]{Value: key},
	}
}

// HasAnyKey (?|) - JSON has any of the keys
func (f *JsonOperandColumn[C]) HasAnyKey(keys ...string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "?|",
		Right:    &SliceExprClause[C]{Values: keys},
	}
}

// HasAllKeys (?&) - JSON has all of the keys
func (f *JsonOperandColumn[C]) HasAllKeys(keys ...string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "?&",
		Right:    &SliceExprClause[C]{Values: keys},
	}
}

// JsonPathQuery (@?) - JSONPath query
func (f *JsonOperandColumn[C]) JsonPathQuery(path string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@?",
		Right:    &ParamExprClause[C]{Value: path},
	}
}

// JsonPathPredicate (@@) - JSONPath predicate
func (f *JsonOperandColumn[C]) JsonPathPredicate(path string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@@",
		Right:    &ParamExprClause[C]{Value: path},
	}
}

// IsNull checks if JSON is null
func (f *JsonOperandColumn[C]) IsNull() Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "IS NULL",
		Right:    &RawExprClause[C]{SQL: ""},
	}
}

// IsNotNull checks if JSON is not null
func (f *JsonOperandColumn[C]) IsNotNull() Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "IS NOT NULL",
		Right:    &RawExprClause[C]{SQL: ""},
	}
}

// Raw allows custom JSON operations
func (f *JsonOperandColumn[C]) Raw(operator string, sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: operator,
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}

// JsonFieldAccess represents a JSON field access that can be used in comparisons
type JsonFieldAccess[C types.ColumnAlias] struct {
	base     C
	operator string
	key      string
	path     []string
}

// Eq compares JSON field value
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

// Neq compares JSON field value (not equal)
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

// Gt compares JSON field value (greater than)
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

// Lt compares JSON field value (less than)
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

// In checks if JSON field value is in list
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

// Like performs LIKE comparison on JSON text field
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

// JsonFieldClause represents a clause comparing JSON field
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
