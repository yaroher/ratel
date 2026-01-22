package clause

import (
	"github.com/yaroher/ratel/common/types"
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
	InOf(query types.Buildable) Clause[C]
	InRaw(string, ...any) Clause[C]
}
type AnyOperand[V any, C types.ColumnAlias] interface {
	Any(...V) Clause[C]
	NotAny(...V) Clause[C]
	AnyOf(query types.Buildable) Clause[C]
	AnyRaw(string, ...any) Clause[C]
}
type ScalarOperand[V any, C types.ColumnAlias] interface {
	BinaryOperand[V, C]
	AnyOperand[V, C]
	InOperand[V, C]
	BetweenOperand[V, C]
}

type EqOperand[V any, C types.ColumnAlias] interface {
	Eq(V) Clause[C]
	Neq(V) Clause[C]
	EqOf(query types.Buildable) Clause[C]
	EqRaw(string, ...any) Clause[C]
}
type LogicalOperand[V any, C types.ColumnAlias] interface {
	Or(clause ...Clause[C]) Clause[C]
	And(clause ...Clause[C]) Clause[C]
	Not(clause Clause[C]) Clause[C]
}
type AdditionalQueryOperand[V any, C types.ColumnAlias] interface {
	Of(query types.Buildable) Clause[C]
	NotOf(query types.Buildable) Clause[C]
	Raw(string, string, ...any) Clause[C]
	NotRaw(string, string, ...any) Clause[C]
	ExistsOf(query types.Buildable) Clause[C]
	ExistsRaw(string, ...any) Clause[C]
}
type CommonOperand[V any, C types.ColumnAlias] interface {
	EqOperand[V, C]
	LogicalOperand[V, C]
	AdditionalQueryOperand[V, C]
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

type CommonScalarOperand[V any, C types.ColumnAlias] interface {
	CommonOperand[V, C]
	ScalarOperand[V, C]
}
