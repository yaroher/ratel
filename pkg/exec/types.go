package exec

import (
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/types"
)

type Scanner[F types.ColumnAlias] interface {
	GetTarget(string) func() any
	GetSetter(F) func() set.ValueSetter[F]
	GetValue(F) func() any
}
