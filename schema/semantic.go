package schema

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/ddl"
	"github.com/yaroher/ratel/dml/clause"
	"github.com/yaroher/ratel/dml/set"
)

type Column[V any, C types.ColumnAlias] struct {
	*ddl.ColumnDDL[C]
	*clause.ColumnDML[V, C]
	set.SetterColumn[V, C]
}

func newColumn[V any, C types.ColumnAlias](
	alias C,
	dataType ddl.Datatype,
	options ...ddl.ColumnOption[C],
) *Column[V, C] {
	return &Column[V, C]{
		ColumnDDL:    ddl.NewColumnDDL[C](alias, dataType, options...),
		ColumnDML:    clause.NewColumnDML[V, C](alias),
		SetterColumn: set.NewSetColumn[V, C](alias),
	}
}

// SMALLINT

type SmallIntColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[int16, C]
	clause.CommonScalarOperand[int16, C]
}

func SmallIntColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) SmallIntColumnI[C] {
	return newColumn[int16, C](alias, ddl.SMALLINT, options...)
}

type NullSmallIntColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*int16, C]
	clause.CommonScalarOperand[*int16, C]
	clause.IsNullOperand[C]
}

func NullSmallIntColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullSmallIntColumnI[C] {
	return newColumn[*int16, C](alias, ddl.SMALLINT, append(options, ddl.WithNullable[C]())...)
}

// INTEGER

type IntegerColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[int32, C]
	clause.CommonScalarOperand[int32, C]
}

func IntegerColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) IntegerColumnI[C] {
	return newColumn[int32, C](alias, ddl.INTEGER, options...)
}

type NullIntegerColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*int32, C]
	clause.CommonScalarOperand[*int32, C]
	clause.IsNullOperand[C]
}

func NullIntegerColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullIntegerColumnI[C] {
	return newColumn[*int32, C](alias, ddl.INTEGER, append(options, ddl.WithNullable[C]())...)
}

// BIGINT

type BigIntColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[int64, C]
	clause.CommonScalarOperand[int64, C]
}

func BigIntColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) BigIntColumnI[C] {
	return newColumn[int64, C](alias, ddl.BIGINT, options...)
}

type NullBigIntColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*int64, C]
	clause.CommonScalarOperand[*int64, C]
	clause.IsNullOperand[C]
}

func NullBigIntColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullBigIntColumnI[C] {
	return newColumn[*int64, C](alias, ddl.BIGINT, append(options, ddl.WithNullable[C]())...)
}

// REAL

type RealColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[float32, C]
	clause.CommonScalarOperand[float32, C]
}

func RealColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) RealColumnI[C] {
	return newColumn[float32, C](alias, ddl.REAL, options...)
}

type NullRealColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*float32, C]
	clause.CommonScalarOperand[*float32, C]
	clause.IsNullOperand[C]
}

func NullRealColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullRealColumnI[C] {
	return newColumn[*float32, C](alias, ddl.REAL, append(options, ddl.WithNullable[C]())...)
}

// DOUBLE PRECISION

type DoublePrecisionColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[float64, C]
	clause.CommonScalarOperand[float64, C]
}

func DoublePrecisionColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) DoublePrecisionColumnI[C] {
	return newColumn[float64, C](alias, ddl.DOUBLE, options...)
}

type NullDoublePrecisionColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*float64, C]
	clause.CommonScalarOperand[*float64, C]
	clause.IsNullOperand[C]
}

func NullDoublePrecisionColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullDoublePrecisionColumnI[C] {
	return newColumn[*float64, C](alias, ddl.DOUBLE, append(options, ddl.WithNullable[C]())...)
}

// TEXT

type TextColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[string, C]
	clause.CommonScalarOperand[string, C]
	clause.LikeOperand[C]
}

func TextColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) TextColumnI[C] {
	return newColumn[string, C](alias, ddl.TEXT, options...)
}

type NullTextColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*string, C]
	clause.CommonScalarOperand[*string, C]
	clause.IsNullOperand[C]
}

func NullTextColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullTextColumnI[C] {
	return newColumn[*string, C](alias, ddl.TEXT, append(options, ddl.WithNullable[C]())...)
}

// BOOLEAN

type BooleanColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[bool, C]
	clause.CommonScalarOperand[bool, C]
}

func BooleanColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) BooleanColumnI[C] {
	return newColumn[bool, C](alias, ddl.BOOLEAN, options...)
}

type NullBooleanColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*bool, C]
	clause.CommonScalarOperand[*bool, C]
	clause.IsNullOperand[C]
}

func NullBooleanColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullBooleanColumnI[C] {
	return newColumn[*bool, C](alias, ddl.BOOLEAN, append(options, ddl.WithNullable[C]())...)
}

// UUID

type UuidColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[pgtype.UUID, C]
	clause.CommonScalarOperand[pgtype.UUID, C]
}

func UuidColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) UuidColumnI[C] {
	return newColumn[pgtype.UUID, C](alias, ddl.UUID, options...)
}

type NullUuidColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*pgtype.UUID, C]
	clause.CommonScalarOperand[*pgtype.UUID, C]
	clause.IsNullOperand[C]
}

func NullUuidColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullUuidColumnI[C] {
	return newColumn[*pgtype.UUID, C](alias, ddl.UUID, append(options, ddl.WithNullable[C]())...)
}

// TIME

type TimeColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[time.Time, C]
	clause.CommonScalarOperand[time.Time, C]
}

func TimeColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) TimeColumnI[C] {
	return newColumn[time.Time, C](alias, ddl.TIME, options...)
}

type NullTimeColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*time.Time, C]
	clause.CommonScalarOperand[*time.Time, C]
	clause.IsNullOperand[C]
}

func NullTimeColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullTimeColumnI[C] {
	return newColumn[*time.Time, C](alias, ddl.TIME, append(options, ddl.WithNullable[C]())...)
}

// TIMESTAMPTZ

type TimestamptzColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[time.Time, C]
	clause.CommonScalarOperand[time.Time, C]
}

func TimestamptzColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) TimestamptzColumnI[C] {
	return newColumn[time.Time, C](alias, ddl.TIMESTAMPTZ, options...)
}

type NullTimestamptzColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*time.Time, C]
	clause.CommonScalarOperand[*time.Time, C]
	clause.IsNullOperand[C]
}

func NullTimestamptzColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullTimestamptzColumnI[C] {
	return newColumn[*time.Time, C](alias, ddl.TIMESTAMPTZ, append(options, ddl.WithNullable[C]())...)
}

// INTERVAL

type IntervalColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[time.Duration, C]
	clause.CommonScalarOperand[time.Duration, C]
}

func IntervalColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) IntervalColumnI[C] {
	return newColumn[time.Duration, C](alias, ddl.INTERVAL, options...)
}

type NullIntervalColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[*time.Duration, C]
	clause.CommonScalarOperand[*time.Duration, C]
	clause.IsNullOperand[C]
}

func NullIntervalColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullIntervalColumnI[C] {
	return newColumn[*time.Duration, C](alias, ddl.INTERVAL, append(options, ddl.WithNullable[C]())...)
}

// JSON/JSONB

type JSONColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[[]byte, C]
	clause.JsonOperand[C]
}

func JSONColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) JSONColumnI[C] {
	return newColumn[[]byte, C](alias, ddl.JSON, options...)
}

type NullJSONColumnI[C types.ColumnAlias] interface {
	set.SetterColumn[[]byte, C]
	clause.JsonOperand[C]
	clause.IsNullOperand[C]
}

func NullJSONColumn[C types.ColumnAlias](alias C, options ...ddl.ColumnOption[C]) NullJSONColumnI[C] {
	return newColumn[[]byte, C](alias, ddl.JSON, append(options, ddl.WithNullable[C]())...)
}
