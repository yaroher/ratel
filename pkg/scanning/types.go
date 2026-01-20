package scanning

import "github.com/yaroher/ratel/pkg/types"

type Valuer interface {
	Values() []any
}

type Targeter[F types.ColumnAlias] interface {
	Valuer
	GetTarget(string) func() any
	GetSetter(F) func() types.ValueSetter[F]
	GetValue(F) func() any
}
