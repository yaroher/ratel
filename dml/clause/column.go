package clause

import (
	"github.com/yaroher/ratel/common/types"
)

type OperandColumn[V any, C types.ColumnAlias] struct {
	fieldAlias C
	//constructor func() *V
}

func NewColumn[V any, C types.ColumnAlias](fa C) *OperandColumn[V, C] {
	return &OperandColumn[V, C]{
		fieldAlias: fa,
	}
}

func (f *OperandColumn[V, C]) Eq(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &ParamExprClause[C]{Value: val}}
}
func (f *OperandColumn[V, C]) Neq(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "!=", Right: &ParamExprClause[C]{Value: val}}
}
func (f *OperandColumn[V, C]) EqOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *OperandColumn[V, C]) EqRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &RawExprClause[C]{SQL: sql, Args: args}}
}

func (f *OperandColumn[V, C]) Gt(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: ">", Right: &ParamExprClause[C]{Value: val}}
}
func (f *OperandColumn[V, C]) Gte(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: ">=", Right: &ParamExprClause[C]{Value: val}}
}
func (f *OperandColumn[V, C]) Lt(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "<", Right: &ParamExprClause[C]{Value: val}}
}
func (f *OperandColumn[V, C]) Lte(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "<=", Right: &ParamExprClause[C]{Value: val}}
}

func (f *OperandColumn[V, C]) In(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &SliceExprClause[C]{Values: vals}}
}
func (f *OperandColumn[V, C]) NotIn(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &SliceExprClause[C]{Values: vals}, Negate: true}
}
func (f *OperandColumn[V, C]) InOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *OperandColumn[V, C]) InRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &RawExprClause[C]{SQL: sql, Args: args}}
}

func (f *OperandColumn[V, C]) Any(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "= ANY", Right: &SliceExprClause[C]{Values: vals}}
}
func (f *OperandColumn[V, C]) NotAny(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "!= ALL", Right: &SliceExprClause[C]{Values: vals}}
}
func (f *OperandColumn[V, C]) AnyOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "= ANY", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *OperandColumn[V, C]) AnyRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "= ANY", Right: &RawExprClause[C]{SQL: sql, Args: args}}
}

func (f *OperandColumn[V, C]) Between(lower, upper V) Clause[C] {
	return &AndClause[C]{Clauses: []Clause[C]{
		&FieldClause[C]{Field: f.fieldAlias, Operator: ">=", Right: &ParamExprClause[C]{Value: lower}},
		&FieldClause[C]{Field: f.fieldAlias, Operator: "<=", Right: &ParamExprClause[C]{Value: upper}},
	}}
}
func (f *OperandColumn[V, C]) NotBetween(lower, upper V) Clause[C] {
	return &NotClause[C]{Inner: f.Between(lower, upper)}
}

func (f *OperandColumn[V, C]) Like(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "LIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}}
}
func (f *OperandColumn[V, C]) NotLike(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "LIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}, Negate: true}
}
func (f *OperandColumn[V, C]) ILike(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "ILIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}}
}
func (f *OperandColumn[V, C]) NotILike(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "ILIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}, Negate: true}
}

func (f *OperandColumn[V, C]) IsNull() Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IS NULL", Right: &RawExprClause[C]{SQL: ""}}
}
func (f *OperandColumn[V, C]) IsNotNull() Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IS NOT NULL", Right: &RawExprClause[C]{SQL: ""}}
}

func (f *OperandColumn[V, C]) Or(clauses ...Clause[C]) Clause[C] {
	return &OrClause[C]{Clauses: clauses}
}
func (f *OperandColumn[V, C]) And(clauses ...Clause[C]) Clause[C] {
	return &AndClause[C]{Clauses: clauses}
}
func (f *OperandColumn[V, C]) Not(notThisclause Clause[C]) Clause[C] {
	return &NotClause[C]{Inner: notThisclause}
}

func (f *OperandColumn[V, C]) Of(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *OperandColumn[V, C]) NotOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "!=", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *OperandColumn[V, C]) Raw(operator string, sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: operator, Right: &RawExprClause[C]{SQL: sql, Args: args}}
}
func (f *OperandColumn[V, C]) NotRaw(operator string, sql string, args ...any) Clause[C] {
	return &NotClause[C]{Inner: f.Raw(operator, sql, args...)}
}
func (f *OperandColumn[V, C]) ExistsOf(query types.Query) Clause[C] {
	return &ExistsClause[C]{SubQuery: &SubQueryExprClause[C]{Query: query}, Negate: false}
}
func (f *OperandColumn[V, C]) ExistsRaw(sql string, args ...any) Clause[C] {
	return &ExistsClause[C]{SubQuery: &RawExprClause[C]{SQL: sql, Args: args}, Negate: false}
}
