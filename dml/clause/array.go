package clause

import (
	"github.com/yaroher/ratel/common/types"
)

// ArrayOperandColumn provides PostgreSQL array-specific operations
type ArrayOperandColumn[V any, C types.ColumnAlias] struct {
	*OperandColumn[[]V, C]
}

func NewArrayColumn[V any, C types.ColumnAlias](fa C) *ArrayOperandColumn[V, C] {
	return &ArrayOperandColumn[V, C]{
		OperandColumn: NewColumn[[]V, C](fa),
	}
}

// Contains (@>) - array contains all elements
func (f *ArrayOperandColumn[V, C]) Contains(vals ...V) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@>",
		Right:    &SliceExprClause[C]{Values: vals},
	}
}

// ContainedBy (<@) - array is contained by
func (f *ArrayOperandColumn[V, C]) ContainedBy(vals ...V) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "<@",
		Right:    &SliceExprClause[C]{Values: vals},
	}
}

// Overlap (&&) - arrays have common elements
func (f *ArrayOperandColumn[V, C]) Overlap(vals ...V) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "&&",
		Right:    &SliceExprClause[C]{Values: vals},
	}
}

// ContainsRaw (@>) - array contains (raw SQL)
func (f *ArrayOperandColumn[V, C]) ContainsRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "@>",
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}

// ContainedByRaw (<@) - array is contained by (raw SQL)
func (f *ArrayOperandColumn[V, C]) ContainedByRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "<@",
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}

// OverlapRaw (&&) - arrays overlap (raw SQL)
func (f *ArrayOperandColumn[V, C]) OverlapRaw(sql string, args ...any) Clause[C] {
	return &FieldClause[C]{
		Field:    f.fieldAlias,
		Operator: "&&",
		Right:    &RawExprClause[C]{SQL: sql, Args: args},
	}
}

// Length - array length comparison
func (f *ArrayOperandColumn[V, C]) LengthEq(length int) Clause[C] {
	return &RawExprClause[C]{
		SQL:  "array_length(" + f.fieldAlias.String() + ", 1) = ?",
		Args: []any{length},
	}
}

func (f *ArrayOperandColumn[V, C]) LengthGt(length int) Clause[C] {
	return &RawExprClause[C]{
		SQL:  "array_length(" + f.fieldAlias.String() + ", 1) > ?",
		Args: []any{length},
	}
}

func (f *ArrayOperandColumn[V, C]) LengthLt(length int) Clause[C] {
	return &RawExprClause[C]{
		SQL:  "array_length(" + f.fieldAlias.String() + ", 1) < ?",
		Args: []any{length},
	}
}
