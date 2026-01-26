package repository

import (
	"context"

	"github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/types"
)

// ProtoRepository provides query operations with automatic proto <-> scanner conversion
type ProtoRepository[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C], P any] struct {
	scanner   *ScannerRepository[T, C, S]
	converter Converter[S, P]
}

// NewProtoRepository creates a new ProtoRepository
func NewProtoRepository[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C], P any](
	scanner *ScannerRepository[T, C, S],
	converter Converter[S, P],
) *ProtoRepository[T, C, S, P] {
	return &ProtoRepository[T, C, S, P]{
		scanner:   scanner,
		converter: converter,
	}
}

// Query executes a select query and returns results as proto slice
func (r *ProtoRepository[T, C, S, P]) Query(
	ctx context.Context,
	query types.Scannable,
	opts ...exec.QueryOption[C, S],
) ([]P, error) {
	scanners, err := r.scanner.Query(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	protos := make([]P, len(scanners))
	for i, s := range scanners {
		protos[i] = r.converter.ToProto(s)
	}

	return protos, nil
}

// QueryRow executes a select query and returns a single result as proto
func (r *ProtoRepository[T, C, S, P]) QueryRow(
	ctx context.Context,
	query types.Scannable,
	opts ...exec.QueryOption[C, S],
) (P, error) {
	result, err := r.scanner.QueryRow(ctx, query, opts...)
	if err != nil {
		var zero P
		return zero, err
	}

	return r.converter.ToProto(result), nil
}

// Execute executes a non-select query and returns affected rows
func (r *ProtoRepository[T, C, S, P]) Execute(ctx context.Context, query types.Buildable) (int64, error) {
	return r.scanner.Execute(ctx, query)
}

// Scanner returns the underlying ScannerRepository for complex operations
func (r *ProtoRepository[T, C, S, P]) Scanner() *ScannerRepository[T, C, S] {
	return r.scanner
}

// Converter returns the converter for manual conversions
func (r *ProtoRepository[T, C, S, P]) Converter() Converter[S, P] {
	return r.converter
}

// WithDB returns a new ProtoRepository with a different DB (useful for transactions)
func (r *ProtoRepository[T, C, S, P]) WithDB(db exec.DB) *ProtoRepository[T, C, S, P] {
	return &ProtoRepository[T, C, S, P]{
		scanner:   r.scanner.WithDB(db),
		converter: r.converter,
	}
}
