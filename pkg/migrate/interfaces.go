package migrate

import (
	"context"
	"time"
)

// Inspector inspects the current state of a database.
type Inspector interface {
	InspectSchema(ctx context.Context, name string) (*Schema, error)
	InspectRealm(ctx context.Context) (*SchemaState, error)
}

// Differ compares two schema states and returns changes.
type Differ interface {
	Diff(current, desired *SchemaState) ([]Change, error)
}

// Planner generates SQL migration statements from changes.
type Planner interface {
	Plan(ctx context.Context, name string, changes []Change) (*Plan, error)
}

// Migrator is the full migration engine interface.
type Migrator interface {
	Inspector
	Differ
	Planner
	Lock(ctx context.Context, name string, timeout time.Duration) (unlock func() error, err error)
}
