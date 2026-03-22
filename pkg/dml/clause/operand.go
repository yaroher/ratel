package clause

import (
	"github.com/yaroher/ratel/pkg/types"
)

type BinaryOperand[V any, C types.ColumnAlias] interface {
	Gt(V) Clause[C]
	Gte(V) Clause[C]
	Lt(V) Clause[C]
	Lte(V) Clause[C]
}
type BetweenOperand[V any, C types.ColumnAlias] interface {
	Between(V, V) Clause[C]
	NotBetween(V, V) Clause[C]
}
type InOperand[V any, C types.ColumnAlias] interface {
	In(...V) Clause[C]
	NotIn(...V) Clause[C]
	InOf(query types.Query) Clause[C]
	InRaw(string, ...any) Clause[C]
}
type AnyOperand[V any, C types.ColumnAlias] interface {
	Any(...V) Clause[C]
	NotAny(...V) Clause[C]
	AnyOf(query types.Query) Clause[C]
	AnyRaw(string, ...any) Clause[C]
}

type EqOperand[V any, C types.ColumnAlias] interface {
	Eq(V) Clause[C]
	Neq(V) Clause[C]
	EqOf(query types.Query) Clause[C]
	EqRaw(string, ...any) Clause[C]
	EqRef(ref *ColumnRefExpr) Clause[C]
	NeqRef(ref *ColumnRefExpr) Clause[C]
}
type LogicalOperand[C types.ColumnAlias] interface {
	Or(clause ...Clause[C]) Clause[C]
	And(clause ...Clause[C]) Clause[C]
	Not(clause Clause[C]) Clause[C]
}
type AdditionalQueryOperand[C types.ColumnAlias] interface {
	Of(query types.Query) Clause[C]
	NotOf(query types.Query) Clause[C]
	Raw(string, string, ...any) Clause[C]
	NotRaw(string, string, ...any) Clause[C]
	ExistsOf(query types.Query) Clause[C]
	ExistsRaw(string, ...any) Clause[C]
}

type ScalarOperand[V any, C types.ColumnAlias] interface {
	BinaryOperand[V, C]
	AnyOperand[V, C]
	InOperand[V, C]
	BetweenOperand[V, C]
}

type CommonOperand[V any, C types.ColumnAlias] interface {
	EqOperand[V, C]
	LogicalOperand[C]
	AdditionalQueryOperand[C]
}

type CommonScalarOperand[V any, C types.ColumnAlias] interface {
	CommonOperand[V, C]
	ScalarOperand[V, C]
}

type LikeOperand[C types.ColumnAlias] interface {
	Like(s string) Clause[C]
	NotLike(s string) Clause[C]
	ILike(s string) Clause[C]
	NotILike(s string) Clause[C]
}

type IsNullOperand[C types.ColumnAlias] interface {
	IsNull() Clause[C]
	IsNotNull() Clause[C]
}

// ArrayOperand defines operations specific to PostgreSQL array types
type ArrayOperand[V any, C types.ColumnAlias] interface {
	ARRAYContains(...V) Clause[C]
	ARRAYContainedBy(...V) Clause[C]
	ARRAYOverlap(...V) Clause[C]
	ARRAYContainsRaw(string, ...any) Clause[C]
	ARRAYContainedByRaw(string, ...any) Clause[C]
	ARRAYOverlapRaw(string, ...any) Clause[C]
	ARRAYLengthEq(int) Clause[C]
	ARRAYLengthGt(int) Clause[C]
	ARRAYLengthLt(int) Clause[C]
}

// JsonOperand defines operations specific to PostgreSQL JSON/JSONB types
type JsonOperand[C types.ColumnAlias] interface {
	JSONGetField(key string) *JsonFieldAccess[C]
	JSONGetFieldText(key string) *JsonFieldAccess[C]
	JSONGetPath(path ...string) *JsonFieldAccess[C]
	JSONGetPathText(path ...string) *JsonFieldAccess[C]
	JSONContains(jsonValue string) Clause[C]
	JSONContainedBy(jsonValue string) Clause[C]
	JSONHasKey(key string) Clause[C]
	JSONHasAnyKey(keys ...string) Clause[C]
	JSONHasAllKeys(keys ...string) Clause[C]
	JSONJsonPathQuery(path string) Clause[C]
	JSONJsonPathPredicate(path string) Clause[C]
	JSONIsNull() Clause[C]
	JSONIsNotNull() Clause[C]
	JSONRaw(operator string, sql string, args ...any) Clause[C]
}
