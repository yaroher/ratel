package exec

import (
	"context"

	"github.com/yaroher/ratel/pkg/types"
)

type RelationLoader[S any] interface {
	Load(ctx context.Context, db DB, base S) error
}

type RelationProvider[S any] interface {
	Relations() []RelationLoader[S]
}

// QueryOption - типизированная опция для Query/QueryRow
type QueryOption[C types.ColumnAlias, S Scanner[C]] interface {
	ApplyQuery(*QueryConfig[C, S])
}

// QueryConfig holds configuration for query execution
type QueryConfig[C types.ColumnAlias, S Scanner[C]] struct {
	Loaders []RelationLoader[S]
}

// LoadOption реализует QueryOption для загрузки указанных relations
type LoadOption[C types.ColumnAlias, S Scanner[C]] struct {
	loaders []RelationLoader[S]
}

// ApplyQuery implements QueryOption interface
func (o *LoadOption[C, S]) ApplyQuery(cfg *QueryConfig[C, S]) {
	cfg.Loaders = append(cfg.Loaders, o.loaders...)
}

// WithRelationLoaders создаёт опцию из loaders (используется в генерации)
func WithRelationLoaders[C types.ColumnAlias, S Scanner[C]](
	loaders ...RelationLoader[S],
) QueryOption[C, S] {
	return &LoadOption[C, S]{loaders: loaders}
}

type skipRelationsKey struct{}
type loadRelationsKey struct{}

func WithSkipRelations(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipRelationsKey{}, true)
}

func SkipRelations(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value, ok := ctx.Value(skipRelationsKey{}).(bool)
	return ok && value
}

func WithRelations(ctx context.Context) context.Context {
	return context.WithValue(ctx, loadRelationsKey{}, true)
}

func LoadRelations(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value, ok := ctx.Value(loadRelationsKey{}).(bool)
	return ok && value
}
