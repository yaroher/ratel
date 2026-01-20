package scanning

import (
	"fmt"

	"github.com/yaroher/ratel/pkg/types"
)

type Valuer interface {
	Values() []any
}

// TargetResolver resolves scan targets by column name.
type TargetResolver interface {
	Targets(columns []string) ([]any, error)
}

type Targeter[F types.ColumnAlias] interface {
	Valuer
	GetTarget(string) func() any
	GetSetter(F) func() types.ValueSetter[F]
	GetValue(F) func() any
}

// UnknownColumnError reports missing scan target by column name.
type UnknownColumnError struct {
	Column string
}

func (e UnknownColumnError) Error() string {
	return fmt.Sprintf("scanning: unknown column %q", e.Column)
}

// FieldAccess describes how to scan and serialize a field.
// Name should be a database column name, compatible with query.ScanAbleFields().
type FieldAccess[F types.ColumnAlias] struct {
	Name   string
	Target func() any
	Value  func() any
	Setter func() types.ValueSetter[F]
}

// BaseTargeter is a generator-friendly implementation of Targeter.
// It provides lookup by column name and ordered Values() for CopyFrom.
type BaseTargeter[F types.ColumnAlias] struct {
	targets      map[string]func() any
	valuesByName map[string]func() any
	setters      map[string]func() types.ValueSetter[F]
	values       []func() any
}

// NewBaseTargeter builds a BaseTargeter using ordered field accessors.
func NewBaseTargeter[F types.ColumnAlias](fields ...FieldAccess[F]) *BaseTargeter[F] {
	t := &BaseTargeter[F]{
		targets:      make(map[string]func() any, len(fields)),
		valuesByName: make(map[string]func() any, len(fields)),
		setters:      make(map[string]func() types.ValueSetter[F], len(fields)),
		values:       make([]func() any, 0, len(fields)),
	}
	for _, f := range fields {
		if f.Name == "" {
			continue
		}
		if f.Target != nil {
			t.targets[f.Name] = f.Target
		}
		if f.Value != nil {
			t.valuesByName[f.Name] = f.Value
			t.values = append(t.values, f.Value)
		} else {
			t.values = append(t.values, func() any { return nil })
		}
		if f.Setter != nil {
			t.setters[f.Name] = f.Setter
		}
	}
	return t
}

// Targets resolves scan targets in the given column order.
func (t *BaseTargeter[F]) Targets(columns []string) ([]any, error) {
	targets := make([]any, 0, len(columns))
	for _, col := range columns {
		fn, ok := t.targets[col]
		if !ok || fn == nil {
			return nil, UnknownColumnError{Column: col}
		}
		targets = append(targets, fn())
	}
	return targets, nil
}

// Values returns ordered values for CopyFrom.
func (t *BaseTargeter[F]) Values() []any {
	out := make([]any, 0, len(t.values))
	for _, fn := range t.values {
		if fn == nil {
			out = append(out, nil)
			continue
		}
		out = append(out, fn())
	}
	return out
}

// GetTarget returns a scan target factory for the column name.
func (t *BaseTargeter[F]) GetTarget(column string) func() any {
	return t.targets[column]
}

// GetSetter returns a setter factory for the given column alias.
func (t *BaseTargeter[F]) GetSetter(column F) func() types.ValueSetter[F] {
	return t.setters[column.String()]
}

// GetValue returns a value factory for the given column alias.
func (t *BaseTargeter[F]) GetValue(column F) func() any {
	return t.valuesByName[column.String()]
}
