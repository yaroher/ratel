package schema

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/dml/clause"
	"github.com/yaroher/ratel/set"
)

type collType interface {
	ability()
}

// SMALLINT

type SmallIntColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[int16, C]
	clause.CommonScalarOperand[int16, C]
}

type NullSmallIntColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*int16, C]
	clause.CommonScalarOperand[*int16, C]
	clause.IsNullOperand[C]
}

// INTEGER

type IntegerColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[int32, C]
	clause.CommonScalarOperand[int32, C]
}

type NullIntegerColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*int32, C]
	clause.CommonScalarOperand[*int32, C]
	clause.IsNullOperand[C]
}

// BIGINT

type BigIntColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[int64, C]
	clause.CommonScalarOperand[int64, C]
}

type NullBigIntColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*int64, C]
	clause.CommonScalarOperand[*int64, C]
	clause.IsNullOperand[C]
}

// REAL

type RealColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[float32, C]
	clause.CommonScalarOperand[float32, C]
}

type NullRealColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*float32, C]
	clause.CommonScalarOperand[*float32, C]
	clause.IsNullOperand[C]
}

// DOUBLE PRECISION

type DoublePrecisionColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[float64, C]
	clause.CommonScalarOperand[float64, C]
}

type NullDoublePrecisionColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*float64, C]
	clause.CommonScalarOperand[*float64, C]
	clause.IsNullOperand[C]
}

// TEXT

type TextColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[string, C]
	clause.LikeOperand[C]
}

type NullTextColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*string, C]
	clause.CommonScalarOperand[*string, C]
	clause.IsNullOperand[C]
}

// BOOLEAN

type BooleanColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[bool, C]
	clause.CommonScalarOperand[bool, C]
}

type NullBooleanColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*bool, C]
	clause.CommonScalarOperand[*bool, C]
	clause.IsNullOperand[C]
}

// UUID

type UuidColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[pgtype.UUID, C]
	clause.CommonScalarOperand[pgtype.UUID, C]
}

type NullUuidColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*pgtype.UUID, C]
	clause.CommonScalarOperand[*pgtype.UUID, C]
	clause.IsNullOperand[C]
}

// TIME

type TimeColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[time.Time, C]
	clause.CommonScalarOperand[time.Time, C]
}

type NullTimeColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*time.Time, C]
	clause.CommonScalarOperand[*time.Time, C]
	clause.IsNullOperand[C]
}

// TIMESTAMPTZ

type TimestamptzColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[time.Time, C]
	clause.CommonScalarOperand[time.Time, C]
}

type NullTimestamptzColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*time.Time, C]
	clause.CommonScalarOperand[*time.Time, C]
	clause.IsNullOperand[C]
}

// INTERVAL

type IntervalColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[time.Duration, C]
	clause.CommonScalarOperand[time.Duration, C]
}

type NullIntervalColumn[C types.ColumnAlias] interface {
	collType
	set.SetterColumn[*time.Duration, C]
	clause.CommonScalarOperand[*time.Duration, C]
	clause.IsNullOperand[C]
}
