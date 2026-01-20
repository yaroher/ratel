package clause

import "github.com/yaroher/ratel/pkg/types"

type BinaryOperator[V any, C types.ColumnAlias] interface {
	Gt(V) types.Clause[C]
	Gte(V) types.Clause[C]
	Lt(V) types.Clause[C]
	Lte(V) types.Clause[C]
}
type BetweenOperator[V any, C types.ColumnAlias] interface {
	Between(V, V) types.Clause[C]
	NotBetween(V, V) types.Clause[C]
}
type InOperator[V any, C types.ColumnAlias] interface {
	In(...V) types.Clause[C]
	NotIn(...V) types.Clause[C]
	InOf(query types.OrmQuery) types.Clause[C]
	InRaw(string, ...any) types.Clause[C]
}
type AnyOperator[V any, C types.ColumnAlias] interface {
	Any(...V) types.Clause[C]
	NotAny(...V) types.Clause[C]
	AnyOf(query types.OrmQuery) types.Clause[C]
	AnyRaw(string, ...any) types.Clause[C]
}
type ScalarOperator[V any, C types.ColumnAlias] interface {
	BinaryOperator[V, C]
	AnyOperator[V, C]
	InOperator[V, C]
	BetweenOperator[V, C]
}

type SetterOperator[V any, C types.ColumnAlias] interface {
	SetExpr(string) types.ValueSetter[C]
	Set(V) types.ValueSetter[C]
	SetRaw(sql string, value ...any) types.ValueSetter[C]
}
type EqOperator[V any, C types.ColumnAlias] interface {
	Eq(V) types.Clause[C]
	Neq(V) types.Clause[C]
	EqOf(query types.OrmQuery) types.Clause[C]
	EqRaw(string, ...any) types.Clause[C]
}
type LogicalOperator[V any, C types.ColumnAlias] interface {
	Or(clause ...types.Clause[C]) types.Clause[C]
	And(clause ...types.Clause[C]) types.Clause[C]
	Not(clause types.Clause[C]) types.Clause[C]
}
type AdditionalQueryOperator[V any, C types.ColumnAlias] interface {
	Of(query types.OrmQuery) types.Clause[C]
	NotOf(query types.OrmQuery) types.Clause[C]
	Raw(string, string, ...any) types.Clause[C]
	NotRaw(string, string, ...any) types.Clause[C]
	ExistsOf(query types.OrmQuery) types.Clause[C]
	ExistsRaw(string, ...any) types.Clause[C]
}
type CommonOperator[V any, C types.ColumnAlias] interface {
	SetterOperator[V, C]
	EqOperator[V, C]
	LogicalOperator[V, C]
	AdditionalQueryOperator[V, C]
}

type LikeOperator[C types.ColumnAlias] interface {
	Like(s string) types.Clause[C]
	NotLike(s string) types.Clause[C]
	ILike(s string) types.Clause[C]
	NotILike(s string) types.Clause[C]
}

type IsNullOperator[C types.ColumnAlias] interface {
	IsNull() types.Clause[C]
	IsNotNull() types.Clause[C]
}

type CommonScalarOperator[V any, C types.ColumnAlias] interface {
	CommonOperator[V, C]
	ScalarOperator[V, C]
}
