package exec

import (
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/dml/set"
)

type Scanner[F types.ColumnAlias] interface {
	GetTarget(string) func() any
	GetSetter(F) func() set.ValueSetter[F]
	GetValue(F) func() any
}
