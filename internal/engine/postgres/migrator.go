package postgres

// Based on Atlas (https://github.com/ariga/atlas) — Apache 2.0 License
// Copyright 2021-present The Atlas Authors

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yaroher/ratel/pkg/migrate"
)

// Migrator is the full PostgreSQL migration engine.
type Migrator struct {
	pool      *pgxpool.Pool
	inspector *Inspector
	differ    *Differ
	planner   *Planner
}

// NewMigrator creates a new PostgreSQL Migrator.
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{
		pool:      pool,
		inspector: NewInspector(pool),
		differ:    &Differ{},
		planner:   &Planner{},
	}
}

// InspectSchema delegates to Inspector.
func (m *Migrator) InspectSchema(ctx context.Context, name string) (*migrate.Schema, error) {
	return m.inspector.InspectSchema(ctx, name)
}

// InspectRealm delegates to Inspector.
func (m *Migrator) InspectRealm(ctx context.Context) (*migrate.SchemaState, error) {
	return m.inspector.InspectRealm(ctx)
}

// Diff delegates to Differ.
func (m *Migrator) Diff(current, desired *migrate.SchemaState) ([]migrate.Change, error) {
	return m.differ.Diff(current, desired)
}

// Plan delegates to Planner.
func (m *Migrator) Plan(ctx context.Context, name string, changes []migrate.Change) (*migrate.Plan, error) {
	return m.planner.Plan(ctx, name, changes)
}

// Lock acquires a PostgreSQL advisory lock.
func (m *Migrator) Lock(ctx context.Context, name string, timeout time.Duration) (func() error, error) {
	// Use pg_advisory_lock with a hash of the name
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquiring connection for lock: %w", err)
	}

	// FNV-1a hash of name to int64 for advisory lock (good distribution).
	const (
		fnvOffset int64 = -3750763034362895579 // 14695981039346656037 as int64
		fnvPrime  int64 = 1099511628211
	)
	lockID := fnvOffset
	for _, c := range name {
		lockID ^= int64(c)
		lockID *= fnvPrime
	}

	// Set session-level statement timeout for the lock attempt.
	_, err = conn.Exec(ctx, fmt.Sprintf("SET statement_timeout = '%dms'", timeout.Milliseconds()))
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("setting lock timeout: %w", err)
	}

	_, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockID)
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("acquiring advisory lock: %w", err)
	}

	// Reset timeout after acquiring lock.
	_, _ = conn.Exec(ctx, "SET statement_timeout = 0")

	unlock := func() error {
		defer conn.Release()
		_, err := conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", lockID)
		return err
	}

	return unlock, nil
}
