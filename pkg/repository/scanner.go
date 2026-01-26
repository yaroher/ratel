package repository

import (
	"context"

	"github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/schema"
	"github.com/yaroher/ratel/pkg/types"
)

// ScannerRepository provides CRUD operations for scanner types
type ScannerRepository[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C]] struct {
	table *schema.Table[T, C, S]
	db    exec.DB
}

// NewScannerRepository creates a new ScannerRepository
func NewScannerRepository[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C]](
	table *schema.Table[T, C, S],
	db exec.DB,
) *ScannerRepository[T, C, S] {
	return &ScannerRepository[T, C, S]{
		table: table,
		db:    db,
	}
}

// Query executes a select query and returns multiple results
func (r *ScannerRepository[T, C, S]) Query(
	ctx context.Context,
	query types.Scannable,
	opts ...exec.QueryOption[C, S],
) ([]S, error) {
	return r.table.Query(ctx, r.db, query, opts...)
}

// QueryRow executes a select query and returns a single result
func (r *ScannerRepository[T, C, S]) QueryRow(
	ctx context.Context,
	query types.Scannable,
	opts ...exec.QueryOption[C, S],
) (S, error) {
	return r.table.QueryRow(ctx, r.db, query, opts...)
}

// Execute executes a non-select query (INSERT, UPDATE, DELETE) and returns affected rows
func (r *ScannerRepository[T, C, S]) Execute(ctx context.Context, query types.Buildable) (int64, error) {
	return r.table.Execute(ctx, r.db, query)
}

// Table returns the underlying schema.Table for building custom queries
func (r *ScannerRepository[T, C, S]) Table() *schema.Table[T, C, S] {
	return r.table
}

// DB returns the underlying database connection
func (r *ScannerRepository[T, C, S]) DB() exec.DB {
	return r.db
}

// WithDB returns a new ScannerRepository with a different DB (useful for transactions)
func (r *ScannerRepository[T, C, S]) WithDB(db exec.DB) *ScannerRepository[T, C, S] {
	return &ScannerRepository[T, C, S]{
		table: r.table,
		db:    db,
	}
}
