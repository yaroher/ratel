package schema

import (
	"time"

	"github.com/yaroher/ratel/pkg/clause"
	"github.com/yaroher/ratel/pkg/types"
)

// Postgres Types

func IntegerColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[int32, C] {
	return NewColumn[int32, C](alias)
}

func BigIntColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[int64, C] {
	return NewColumn[int64, C](alias)
}

func RealColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[float32, C] {
	return NewColumn[float32, C](alias)
}

func DoublePrecisionColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[float64, C] {
	return NewColumn[float64, C](alias)
}

func BoolColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[bool, C] {
	return NewColumn[bool, C](alias)
}

func TextColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[string, C] {
	return NewColumn[string, C](alias)
}

func StringColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[string, C] {
	return NewColumn[string, C](alias)
}

func ByteaColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[[]byte, C] {
	return NewColumn[[]byte, C](alias)
}

func TimestamptzColumn[C types.ColumnAlias](alias C) clause.CommonScalarOperator[time.Time, C] {
	return NewColumn[time.Time, C](alias)
}

func Nullable[V any, C types.ColumnAlias](c clause.CommonScalarOperator[V, C]) interface {
	clause.CommonScalarOperator[V, C]
	clause.IsNullOperator[C]
} {
	col, ok := c.(*Column[V, C])
	if !ok {
		panic("Nullable requires a Column type")
	}
	return &collWrapper[V, C]{*col}
}

//func Array[V any, C types.ColumnAlias](c clause.CommonScalarOperator[V, C]) clause.CommonScalarOperator[[]V, C] {
//
//}

type alias string

func (a alias) String() string {
	return string(a)
}
func Test() {
	a := alias("test")
	_ = Nullable[int32, alias](IntegerColumn(a))
}
