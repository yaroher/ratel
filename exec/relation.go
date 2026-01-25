package exec

import "context"

type RelationLoader[S any] interface {
	Load(ctx context.Context, db DB, base S) error
}

type RelationProvider[S any] interface {
	Relations() []RelationLoader[S]
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
