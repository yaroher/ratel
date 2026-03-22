package clause

import (
	"github.com/yaroher/ratel/pkg/types"
)

type ColumnDML[V any, C types.ColumnAlias] struct {
	fieldAlias C
	//constructor func() *V
}

func NewColumnDML[V any, C types.ColumnAlias](fa C) *ColumnDML[V, C] {
	return &ColumnDML[V, C]{
		fieldAlias: fa,
	}
}

func (f *ColumnDML[V, C]) Eq(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &ParamExprClause[C]{Value: val}}
}
func (f *ColumnDML[V, C]) Neq(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "!=", Right: &ParamExprClause[C]{Value: val}}
}
func (f *ColumnDML[V, C]) EqOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *ColumnDML[V, C]) EqRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &RawExprClause[C]{SQL: sql, Args: args}}
}

// EqRef compares this column to a column from another table (for correlated subqueries).
// Example: Orders.UserId.EqRef(Users.Ref(UserColumnId)) produces "orders.user_id = users.id".
func (f *ColumnDML[V, C]) EqRef(ref *ColumnRefExpr) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: ref}
}
func (f *ColumnDML[V, C]) NeqRef(ref *ColumnRefExpr) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "!=", Right: ref}
}

func (f *ColumnDML[V, C]) Gt(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: ">", Right: &ParamExprClause[C]{Value: val}}
}
func (f *ColumnDML[V, C]) Gte(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: ">=", Right: &ParamExprClause[C]{Value: val}}
}
func (f *ColumnDML[V, C]) Lt(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "<", Right: &ParamExprClause[C]{Value: val}}
}
func (f *ColumnDML[V, C]) Lte(val V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "<=", Right: &ParamExprClause[C]{Value: val}}
}

func (f *ColumnDML[V, C]) In(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &SliceExprClause[C]{Values: vals}}
}
func (f *ColumnDML[V, C]) NotIn(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &SliceExprClause[C]{Values: vals}, Negate: true}
}
func (f *ColumnDML[V, C]) InOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *ColumnDML[V, C]) InRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IN", Right: &RawExprClause[C]{SQL: sql, Args: args}}
}

func (f *ColumnDML[V, C]) Any(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "= ANY", Right: &SliceExprClause[C]{Values: vals}}
}
func (f *ColumnDML[V, C]) NotAny(vals ...V) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "!= ALL", Right: &SliceExprClause[C]{Values: vals}}
}
func (f *ColumnDML[V, C]) AnyOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "= ANY", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *ColumnDML[V, C]) AnyRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "= ANY", Right: &RawExprClause[C]{SQL: sql, Args: args}}
}

func (f *ColumnDML[V, C]) Between(lower, upper V) Clause[C] {
	return &AndClause[C]{Clauses: []Clause[C]{
		&FieldClause[C]{Field: f.fieldAlias, Operator: ">=", Right: &ParamExprClause[C]{Value: lower}},
		&FieldClause[C]{Field: f.fieldAlias, Operator: "<=", Right: &ParamExprClause[C]{Value: upper}},
	}}
}
func (f *ColumnDML[V, C]) NotBetween(lower, upper V) Clause[C] {
	return &NotClause[C]{Inner: f.Between(lower, upper)}
}

func (f *ColumnDML[V, C]) Like(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "LIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}}
}
func (f *ColumnDML[V, C]) NotLike(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "LIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}, Negate: true}
}
func (f *ColumnDML[V, C]) ILike(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "ILIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}}
}
func (f *ColumnDML[V, C]) NotILike(pattern string) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "ILIKE", Right: &ParamExprClause[C]{Value: pattern, LikeWrapp: true}, Negate: true}
}

func (f *ColumnDML[V, C]) IsNull() Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IS NULL", Right: &RawExprClause[C]{SQL: ""}}
}
func (f *ColumnDML[V, C]) IsNotNull() Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "IS NOT NULL", Right: &RawExprClause[C]{SQL: ""}}
}

func (f *ColumnDML[V, C]) Or(clauses ...Clause[C]) Clause[C] {
	return &OrClause[C]{Clauses: clauses}
}
func (f *ColumnDML[V, C]) And(clauses ...Clause[C]) Clause[C] {
	return &AndClause[C]{Clauses: clauses}
}
func (f *ColumnDML[V, C]) Not(notThisclause Clause[C]) Clause[C] {
	return &NotClause[C]{Inner: notThisclause}
}

func (f *ColumnDML[V, C]) Of(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "=", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *ColumnDML[V, C]) NotOf(query types.Query) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: "!=", Right: &SubQueryExprClause[C]{Query: query}}
}
func (f *ColumnDML[V, C]) Raw(operator string, sql string, args ...any) Clause[C] {
	return &FieldClause[C]{Field: f.fieldAlias, Operator: operator, Right: &RawExprClause[C]{SQL: sql, Args: args}}
}
func (f *ColumnDML[V, C]) NotRaw(operator string, sql string, args ...any) Clause[C] {
	return &NotClause[C]{Inner: f.Raw(operator, sql, args...)}
}
func (f *ColumnDML[V, C]) ExistsOf(query types.Query) Clause[C] {
	return &ExistsClause[C]{SubQuery: &SubQueryExprClause[C]{Query: query}, Negate: false}
}
func (f *ColumnDML[V, C]) ExistsRaw(sql string, args ...any) Clause[C] {
	return &ExistsClause[C]{SubQuery: &RawExprClause[C]{SQL: sql, Args: args}, Negate: false}
}

func (f *ColumnDML[V, C]) ARRAYContains(vals ...V) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@>",
		Right:    &SliceExprClause[C]{Values: vals},
	}
}
func (f *ColumnDML[V, C]) ARRAYContainedBy(vals ...V) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "<@",
		Right:    &SliceExprClause[C]{Values: vals},
	}
}
func (f *ColumnDML[V, C]) ARRAYOverlap(vals ...V) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "&&",
		Right:    &SliceExprClause[C]{Values: vals},
	}
}
func (f *ColumnDML[V, C]) ARRAYContainsRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@>",
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}
func (f *ColumnDML[V, C]) ARRAYContainedByRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "<@",
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}
func (f *ColumnDML[V, C]) ARRAYOverlapRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "&&",
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}
func (f *ColumnDML[V, C]) ARRAYLengthEq(length int) Clause[C] {
	return &RawExprClause[C]{
		SQL:  "array_length(" + f.fieldAlias.String() + ", 1) = ?",
		Args: []any{length},
	}
}
func (f *ColumnDML[V, C]) ARRAYLengthGt(length int) Clause[C] {
	return &RawExprClause[C]{
		SQL:  "array_length(" + f.fieldAlias.String() + ", 1) > ?",
		Args: []any{length},
	}
}
func (f *ColumnDML[V, C]) ARRAYLengthLt(length int) Clause[C] {
	return &RawExprClause[C]{
		SQL:  "array_length(" + f.fieldAlias.String() + ", 1) < ?",
		Args: []any{length},
	}
}

func (f *ColumnDML[V, C]) JSONGetField(key string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "->",
		key:      key,
	}
}
func (f *ColumnDML[V, C]) JSONGetFieldText(key string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "->>",
		key:      key,
	}
}
func (f *ColumnDML[V, C]) JSONGetPath(path ...string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "#>",
		path:     path,
	}
}
func (f *ColumnDML[V, C]) JSONGetPathText(path ...string) *JsonFieldAccess[C] {
	return &JsonFieldAccess[C]{
		base:     f.fieldAlias,
		operator: "#>>",
		path:     path,
	}
}
func (f *ColumnDML[V, C]) JSONContains(jsonValue string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@>",
		Right:    &ParamExprClause[C]{Value: jsonValue},
	}
}
func (f *ColumnDML[V, C]) JSONContainedBy(jsonValue string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "<@",
		Right:    &ParamExprClause[C]{Value: jsonValue},
	}
}
func (f *ColumnDML[V, C]) JSONHasKey(key string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "?",
		Right:    &ParamExprClause[C]{Value: key},
	}
}
func (f *ColumnDML[V, C]) JSONHasAnyKey(keys ...string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "?|",
		Right:    &SliceExprClause[C]{Values: keys},
	}
}
func (f *ColumnDML[V, C]) JSONHasAllKeys(keys ...string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "?&",
		Right:    &SliceExprClause[C]{Values: keys},
	}
}
func (f *ColumnDML[V, C]) JSONJsonPathQuery(path string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@?",
		Right:    &ParamExprClause[C]{Value: path},
	}
}
func (f *ColumnDML[V, C]) JSONJsonPathPredicate(path string) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@@",
		Right:    &ParamExprClause[C]{Value: path},
	}
}
func (f *ColumnDML[V, C]) JSONIsNull() Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "IS NULL",
		Right:    &RawExprClause[C]{SQL: ""},
	}
}
func (f *ColumnDML[V, C]) JSONIsNotNull() Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "IS NOT NULL",
		Right:    &RawExprClause[C]{SQL: ""},
	}
}
func (f *ColumnDML[V, C]) JSONRaw(operator string, sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: operator,
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}
