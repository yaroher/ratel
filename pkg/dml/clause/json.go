package clause

import (
	"github.com/yaroher/ratel/pkg/types"
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
