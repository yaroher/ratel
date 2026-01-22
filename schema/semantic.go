package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/sqlbuild/dml/clause"
	"github.com/yaroher/ratel/sqlbuild/set"
)

type collType interface {
	ability()
}

type smallIntColumn[V int16 | *int16, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type intergerColumn[V int32 | *int32, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type bigIntColumn[V int64 | *int64, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type realColumn[V float32 | *float32, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type doublePresitionColumn[V float64 | *float64, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type textColumn[V string | *string, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.LikeOperand[C]
}

type booleanColumn[V bool | *bool, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type uuidColumn[V uuid.UUID | *uuid.UUID, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type timeColumn[V time.Time | *time.Time, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type timestamptzColumn[V time.Time | *time.Time, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}

type intervalColumn[V time.Duration | *time.Duration, C types.ColumnAlias] interface {
	collType
	set.SetterColumn[V, C]
	clause.CommonScalarOperand[V, C]
}
