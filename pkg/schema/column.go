package schema

import (
	"github.com/yaroher/ratel/pkg/clause"
	"github.com/yaroher/ratel/pkg/query"
	"github.com/yaroher/ratel/pkg/types"
)

type aliasImpl[C types.ColumnAlias] struct {
	field string
}

func (c *aliasImpl[C]) String() string {
	return c.field
}

func NewColumnAlias[C types.ColumnAlias](fa string) types.ColumnAlias {
	return &aliasImpl[C]{field: fa}
}

type collWrapper[V any, C types.ColumnAlias] struct {
	Column[V, C]
}

func newCollWrapper[V any, C types.ColumnAlias](c *Column[V, C]) *collWrapper[V, C] {
	return &collWrapper[V, C]{Column: *c}
}

type Column[V any, C types.ColumnAlias] struct {
	fieldAlias C
	//constructor func() *V
}

func NewColumn[V any, C types.ColumnAlias](fa C) *Column[V, C] {
	return &Column[V, C]{
		fieldAlias: fa,
	}
}

func (f *Column[V, C]) Set(val V) types.ValueSetter[C] {
	return query.NewValueSetter[C](f.fieldAlias, val)
}
func (f *Column[V, C]) SetExpr(expr string) types.ValueSetter[C] {
	return query.NewValueSetter[C](f.fieldAlias, expr)
}
func (f *Column[V, C]) SetRaw(sql string, value ...any) types.ValueSetter[C] {
	return query.NewValueSetter[C](f.fieldAlias, &clause.RawExprClause[C]{SQL: sql, Args: value})
}
func (f *Column[V, C]) String() string {
	return f.fieldAlias.String()
}

func (f *Column[V, C]) Eq(val V) types.Clause[C] {
	return &clause.FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &clause.ParamExprClause[C]{Value: val}}
}
func (f *Column[V, С]) Neq(val V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "!=", Right: &clause.ParamExprClause[С]{Value: val}}
}
func (f *Column[V, С]) EqOf(query types.OrmQuery) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "=", Right: &clause.SubQueryExprClause[С]{Query: query}}
}
func (f *Column[V, С]) EqRaw(sql string, args ...any) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "=", Right: &clause.RawExprClause[С]{SQL: sql, Args: args}}
}

func (f *Column[V, С]) Gt(val V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: ">", Right: &clause.ParamExprClause[С]{Value: val}}
}
func (f *Column[V, С]) Gte(val V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: ">=", Right: &clause.ParamExprClause[С]{Value: val}}
}
func (f *Column[V, С]) Lt(val V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "<", Right: &clause.ParamExprClause[С]{Value: val}}
}
func (f *Column[V, С]) Lte(val V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "<=", Right: &clause.ParamExprClause[С]{Value: val}}
}

func (f *Column[V, С]) In(vals ...V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "IN", Right: &clause.SliceExprClause[С]{Values: vals}}
}
func (f *Column[V, С]) NotIn(vals ...V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "IN", Right: &clause.SliceExprClause[С]{Values: vals}, Negate: true}
}
func (f *Column[V, С]) InOf(query types.OrmQuery) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "IN", Right: &clause.SubQueryExprClause[С]{Query: query}}
}
func (f *Column[V, С]) InRaw(sql string, args ...any) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "IN", Right: &clause.RawExprClause[С]{SQL: sql, Args: args}}
}

func (f *Column[V, С]) Any(vals ...V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "= ANY", Right: &clause.SliceExprClause[С]{Values: vals}}
}
func (f *Column[V, С]) NotAny(vals ...V) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "!= ALL", Right: &clause.SliceExprClause[С]{Values: vals}}
}
func (f *Column[V, С]) AnyOf(query types.OrmQuery) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "= ANY", Right: &clause.SubQueryExprClause[С]{Query: query}}
}
func (f *Column[V, С]) AnyRaw(sql string, args ...any) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "= ANY", Right: &clause.RawExprClause[С]{SQL: sql, Args: args}}
}

func (f *Column[V, С]) Between(lower, upper V) types.Clause[С] {
	return &clause.AndClause[С]{Clauses: []types.Clause[С]{
		&clause.FieldClause[С]{Field: f.fieldAlias, Operator: ">=", Right: &clause.ParamExprClause[С]{Value: lower}},
		&clause.FieldClause[С]{Field: f.fieldAlias, Operator: "<=", Right: &clause.ParamExprClause[С]{Value: upper}},
	}}
}
func (f *Column[V, С]) NotBetween(lower, upper V) types.Clause[С] {
	return &clause.NotClause[С]{Inner: f.Between(lower, upper)}
}

func (f *Column[V, С]) Like(pattern string) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "LIKE", Right: &clause.ParamExprClause[С]{Value: pattern, LikeWrapp: true}}
}
func (f *Column[V, С]) NotLike(pattern string) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "LIKE", Right: &clause.ParamExprClause[С]{Value: pattern, LikeWrapp: true}, Negate: true}
}
func (f *Column[V, С]) ILike(pattern string) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "ILIKE", Right: &clause.ParamExprClause[С]{Value: pattern, LikeWrapp: true}}
}
func (f *Column[V, С]) NotILike(pattern string) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "ILIKE", Right: &clause.ParamExprClause[С]{Value: pattern, LikeWrapp: true}, Negate: true}
}

func (f *Column[V, С]) IsNull() types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "IS NULL", Right: &clause.RawExprClause[С]{SQL: ""}}
}
func (f *Column[V, С]) IsNotNull() types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "IS NOT NULL", Right: &clause.RawExprClause[С]{SQL: ""}}
}

func (f *Column[V, С]) Or(clauses ...types.Clause[С]) types.Clause[С] {
	return &clause.OrClause[С]{Clauses: clauses}
}
func (f *Column[V, С]) And(clauses ...types.Clause[С]) types.Clause[С] {
	return &clause.AndClause[С]{Clauses: clauses}
}
func (f *Column[V, С]) Not(notThisclause types.Clause[С]) types.Clause[С] {
	return &clause.NotClause[С]{Inner: notThisclause}
}

func (f *Column[V, С]) Of(query types.OrmQuery) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "=", Right: &clause.SubQueryExprClause[С]{Query: query}}
}
func (f *Column[V, С]) NotOf(query types.OrmQuery) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: "!=", Right: &clause.SubQueryExprClause[С]{Query: query}}
}
func (f *Column[V, С]) Raw(operator string, sql string, args ...any) types.Clause[С] {
	return &clause.FieldClause[С]{Field: f.fieldAlias, Operator: operator, Right: &clause.RawExprClause[С]{SQL: sql, Args: args}}
}
func (f *Column[V, С]) NotRaw(operator string, sql string, args ...any) types.Clause[С] {
	return &clause.NotClause[С]{Inner: f.Raw(operator, sql, args...)}
}
func (f *Column[V, С]) ExistsOf(query types.OrmQuery) types.Clause[С] {
	return &clause.ExistsClause[С]{SubQuery: &clause.SubQueryExprClause[С]{Query: query}, Negate: false}
}
func (f *Column[V, С]) ExistsRaw(sql string, args ...any) types.Clause[С] {
	return &clause.ExistsClause[С]{SubQuery: &clause.RawExprClause[С]{SQL: sql, Args: args}, Negate: false}
}
