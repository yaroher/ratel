package schema

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	ddl2 "github.com/yaroher/ratel/pkg/ddl"
	clause2 "github.com/yaroher/ratel/pkg/dml/clause"
	set2 "github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/types"
)

type Column[V any, C types.ColumnAlias] struct {
	*ddl2.ColumnDDL[C]
	*clause2.ColumnDML[V, C]
	set2.SetterColumn[V, C]
}

func newColumn[V any, C types.ColumnAlias](
	alias C,
	dataType ddl2.Datatype,
	options ...ddl2.ColumnOption[C],
) *Column[V, C] {
	return &Column[V, C]{
		ColumnDDL:    ddl2.NewColumnDDL[C](alias, dataType, options...),
		ColumnDML:    clause2.NewColumnDML[V, C](alias),
		SetterColumn: set2.NewSetColumn[V, C](alias),
	}
}

func (c *Column[V, C]) DDL() *ddl2.ColumnDDL[C] {
	return c.ColumnDDL
}

type ddlAbles[C types.ColumnAlias] interface {
	DDL() *ddl2.ColumnDDL[C]
}

// SMALLINT

type SmallIntColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[int16, C]
	clause2.CommonScalarOperand[int16, C]
}

func SmallIntColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) SmallIntColumnI[C] {
	return newColumn[int16, C](alias, ddl2.SMALLINT, options...)
}

type NullSmallIntColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*int16, C]
	clause2.CommonScalarOperand[*int16, C]
	clause2.IsNullOperand[C]
}

func NullSmallIntColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullSmallIntColumnI[C] {
	return newColumn[*int16, C](alias, ddl2.SMALLINT, append(options, ddl2.WithNullable[C]())...)
}

// INTEGER

type IntegerColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[int32, C]
	clause2.CommonScalarOperand[int32, C]
}

func IntegerColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) IntegerColumnI[C] {
	return newColumn[int32, C](alias, ddl2.INTEGER, options...)
}

type NullIntegerColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*int32, C]
	clause2.CommonScalarOperand[*int32, C]
	clause2.IsNullOperand[C]
}

func NullIntegerColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullIntegerColumnI[C] {
	return newColumn[*int32, C](alias, ddl2.INTEGER, append(options, ddl2.WithNullable[C]())...)
}

// BIGINT

type BigIntColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[int64, C]
	clause2.CommonScalarOperand[int64, C]
}

func BigIntColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) BigIntColumnI[C] {
	return newColumn[int64, C](alias, ddl2.BIGINT, options...)
}

type NullBigIntColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*int64, C]
	clause2.CommonScalarOperand[*int64, C]
	clause2.IsNullOperand[C]
}

func NullBigIntColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullBigIntColumnI[C] {
	return newColumn[*int64, C](alias, ddl2.BIGINT, append(options, ddl2.WithNullable[C]())...)
}

type BigSerialColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[int64, C]
	clause2.CommonScalarOperand[int64, C]
}

func BigSerialColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) BigSerialColumnI[C] {
	return newColumn[int64, C](alias, ddl2.BIGSERIAL, options...)
}

type NullBigSerialColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*int64, C]
	clause2.CommonScalarOperand[*int64, C]
	clause2.IsNullOperand[C]
}

func NullBigSerialColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullBigSerialColumnI[C] {
	return newColumn[*int64, C](alias, ddl2.BIGSERIAL, append(options, ddl2.WithNullable[C]())...)
}

// NUMERIC

type NumericColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[float64, C]
	clause2.CommonScalarOperand[float64, C]
}

func NumericColumn[C types.ColumnAlias](alias C, precision, scale int, options ...ddl2.ColumnOption[C]) NumericColumnI[C] {
	return newColumn[float64, C](alias, ddl2.Numeric(precision, scale), options...)
}

type NullNumericColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*float64, C]
	clause2.CommonScalarOperand[*float64, C]
	clause2.IsNullOperand[C]
}

func NullNumericColumn[C types.ColumnAlias](alias C, precision, scale int, options ...ddl2.ColumnOption[C]) NullNumericColumnI[C] {
	return newColumn[*float64, C](alias, ddl2.Numeric(precision, scale), append(options, ddl2.WithNullable[C]())...)
}

// REAL

type RealColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[float32, C]
	clause2.CommonScalarOperand[float32, C]
}

func RealColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) RealColumnI[C] {
	return newColumn[float32, C](alias, ddl2.REAL, options...)
}

type NullRealColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*float32, C]
	clause2.CommonScalarOperand[*float32, C]
	clause2.IsNullOperand[C]
}

func NullRealColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullRealColumnI[C] {
	return newColumn[*float32, C](alias, ddl2.REAL, append(options, ddl2.WithNullable[C]())...)
}

// DOUBLE PRECISION

type DoublePrecisionColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[float64, C]
	clause2.CommonScalarOperand[float64, C]
}

func DoublePrecisionColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) DoublePrecisionColumnI[C] {
	return newColumn[float64, C](alias, ddl2.DOUBLE, options...)
}

type NullDoublePrecisionColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*float64, C]
	clause2.CommonScalarOperand[*float64, C]
	clause2.IsNullOperand[C]
}

func NullDoublePrecisionColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullDoublePrecisionColumnI[C] {
	return newColumn[*float64, C](alias, ddl2.DOUBLE, append(options, ddl2.WithNullable[C]())...)
}

// TEXT

type TextColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[string, C]
	clause2.CommonScalarOperand[string, C]
	clause2.LikeOperand[C]
}

func TextColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) TextColumnI[C] {
	return newColumn[string, C](alias, ddl2.TEXT, options...)
}

type NullTextColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*string, C]
	clause2.CommonScalarOperand[*string, C]
	clause2.IsNullOperand[C]
}

func NullTextColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullTextColumnI[C] {
	return newColumn[*string, C](alias, ddl2.TEXT, append(options, ddl2.WithNullable[C]())...)
}

// CHAR

type CharColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[string, C]
	clause2.CommonScalarOperand[string, C]
	clause2.LikeOperand[C]
}

func CharColumn[C types.ColumnAlias](alias C, length int, options ...ddl2.ColumnOption[C]) CharColumnI[C] {
	return newColumn[string, C](alias, ddl2.Char(length), options...)
}

type NullCharColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*string, C]
	clause2.CommonScalarOperand[*string, C]
	clause2.IsNullOperand[C]
}

func NullCharColumn[C types.ColumnAlias](alias C, length int, options ...ddl2.ColumnOption[C]) NullCharColumnI[C] {
	return newColumn[*string, C](alias, ddl2.Char(length), append(options, ddl2.WithNullable[C]())...)
}

// BOOLEAN

type BooleanColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[bool, C]
	clause2.CommonScalarOperand[bool, C]
}

func BooleanColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) BooleanColumnI[C] {
	return newColumn[bool, C](alias, ddl2.BOOLEAN, options...)
}

type NullBooleanColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*bool, C]
	clause2.CommonScalarOperand[*bool, C]
	clause2.IsNullOperand[C]
}

func NullBooleanColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullBooleanColumnI[C] {
	return newColumn[*bool, C](alias, ddl2.BOOLEAN, append(options, ddl2.WithNullable[C]())...)
}

// UUID

type UuidColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[pgtype.UUID, C]
	clause2.CommonScalarOperand[pgtype.UUID, C]
}

func UuidColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) UuidColumnI[C] {
	return newColumn[pgtype.UUID, C](alias, ddl2.UUID, options...)
}

type NullUuidColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*pgtype.UUID, C]
	clause2.CommonScalarOperand[*pgtype.UUID, C]
	clause2.IsNullOperand[C]
}

func NullUuidColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullUuidColumnI[C] {
	return newColumn[*pgtype.UUID, C](alias, ddl2.UUID, append(options, ddl2.WithNullable[C]())...)
}

// TIME

type TimeColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[time.Time, C]
	clause2.CommonScalarOperand[time.Time, C]
}

func TimeColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) TimeColumnI[C] {
	return newColumn[time.Time, C](alias, ddl2.TIME, options...)
}

type NullTimeColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*time.Time, C]
	clause2.CommonScalarOperand[*time.Time, C]
	clause2.IsNullOperand[C]
}

func NullTimeColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullTimeColumnI[C] {
	return newColumn[*time.Time, C](alias, ddl2.TIME, append(options, ddl2.WithNullable[C]())...)
}

// TIMESTAMPTZ

type TimestamptzColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[time.Time, C]
	clause2.CommonScalarOperand[time.Time, C]
}

func TimestamptzColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) TimestamptzColumnI[C] {
	return newColumn[time.Time, C](alias, ddl2.TIMESTAMPTZ, options...)
}

type NullTimestamptzColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*time.Time, C]
	clause2.CommonScalarOperand[*time.Time, C]
	clause2.IsNullOperand[C]
}

func NullTimestamptzColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullTimestamptzColumnI[C] {
	return newColumn[*time.Time, C](alias, ddl2.TIMESTAMPTZ, append(options, ddl2.WithNullable[C]())...)
}

// INTERVAL

type IntervalColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[time.Duration, C]
	clause2.CommonScalarOperand[time.Duration, C]
}

func IntervalColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) IntervalColumnI[C] {
	return newColumn[time.Duration, C](alias, ddl2.INTERVAL, options...)
}

type NullIntervalColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*time.Duration, C]
	clause2.CommonScalarOperand[*time.Duration, C]
	clause2.IsNullOperand[C]
}

func NullIntervalColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullIntervalColumnI[C] {
	return newColumn[*time.Duration, C](alias, ddl2.INTERVAL, append(options, ddl2.WithNullable[C]())...)
}

// BYTEA

type ByteaColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]byte, C]
	clause2.CommonScalarOperand[[]byte, C]
}

func ByteaColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) ByteaColumnI[C] {
	return newColumn[[]byte, C](alias, ddl2.BYTEA, options...)
}

type NullByteaColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[*[]byte, C]
	clause2.CommonScalarOperand[*[]byte, C]
	clause2.IsNullOperand[C]
}

func NullByteaColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullByteaColumnI[C] {
	return newColumn[*[]byte, C](alias, ddl2.BYTEA, append(options, ddl2.WithNullable[C]())...)
}

// JSON/JSONB

type JSONColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]byte, C]
	clause2.JsonOperand[C]
}

func JSONColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) JSONColumnI[C] {
	return newColumn[[]byte, C](alias, ddl2.JSON, options...)
}

type NullJSONColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]byte, C]
	clause2.JsonOperand[C]
	clause2.IsNullOperand[C]
}

func NullJSONColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullJSONColumnI[C] {
	return newColumn[[]byte, C](alias, ddl2.JSON, append(options, ddl2.WithNullable[C]())...)
}

// JSONB

type JSONBColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]byte, C]
	clause2.JsonOperand[C]
}

func JSONBColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) JSONBColumnI[C] {
	return newColumn[[]byte, C](alias, ddl2.JSONB, options...)
}

type NullJSONBColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]byte, C]
	clause2.JsonOperand[C]
	clause2.IsNullOperand[C]
}

func NullJSONBColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) NullJSONBColumnI[C] {
	return newColumn[[]byte, C](alias, ddl2.JSONB, append(options, ddl2.WithNullable[C]())...)
}

// ============================================================================
// Array Types
// ============================================================================

// TEXT[]

type TextArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]string, C]
}

func TextArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) TextArrayColumnI[C] {
	return newColumn[[]string, C](alias, ddl2.TEXT_ARRAY, options...)
}

// INTEGER[]

type IntegerArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]int32, C]
}

func IntegerArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) IntegerArrayColumnI[C] {
	return newColumn[[]int32, C](alias, ddl2.INTEGER_ARRAY, options...)
}

// BIGINT[]

type BigIntArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]int64, C]
}

func BigIntArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) BigIntArrayColumnI[C] {
	return newColumn[[]int64, C](alias, ddl2.BIGINT_ARRAY, options...)
}

// BOOLEAN[]

type BooleanArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]bool, C]
}

func BooleanArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) BooleanArrayColumnI[C] {
	return newColumn[[]bool, C](alias, ddl2.BOOLEAN_ARRAY, options...)
}

// REAL[]

type RealArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]float32, C]
}

func RealArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) RealArrayColumnI[C] {
	return newColumn[[]float32, C](alias, ddl2.REAL_ARRAY, options...)
}

// DOUBLE PRECISION[]

type DoublePrecisionArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[]float64, C]
}

func DoublePrecisionArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) DoublePrecisionArrayColumnI[C] {
	return newColumn[[]float64, C](alias, ddl2.DOUBLE_ARRAY, options...)
}

// BYTEA[]

type ByteaArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[[][]byte, C]
}

func ByteaArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) ByteaArrayColumnI[C] {
	return newColumn[[][]byte, C](alias, ddl2.BYTEA_ARRAY, options...)
}

// UUID[]

type UuidArrayColumnI[C types.ColumnAlias] interface {
	ddlAbles[C]
	set2.SetterColumn[pgtype.FlatArray[pgtype.UUID], C]
}

func UuidArrayColumn[C types.ColumnAlias](alias C, options ...ddl2.ColumnOption[C]) UuidArrayColumnI[C] {
	return newColumn[pgtype.FlatArray[pgtype.UUID], C](alias, ddl2.UUID_ARRAY, options...)
}
