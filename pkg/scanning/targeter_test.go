package scanning

import (
	"testing"

	"github.com/yaroher/ratel/pkg/types"
)

type testAlias string

func (a testAlias) String() string { return string(a) }

func TestBaseTargeterTargetsAndValues(t *testing.T) {
	type row struct {
		ID   int32
		Name string
	}

	r := &row{}

	bt := NewBaseTargeter[testAlias](
		FieldAccess[testAlias]{
			Name:   "id",
			Target: func() any { return &r.ID },
			Value:  func() any { return r.ID },
			Setter: func() types.ValueSetter[testAlias] { return nil },
		},
		FieldAccess[testAlias]{
			Name:   "name",
			Target: func() any { return &r.Name },
			Value:  func() any { return r.Name },
		},
	)

	targets, err := bt.Targets([]string{"id", "name"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	r.ID = 10
	r.Name = "Bob"
	values := bt.Values()
	if len(values) != 2 || values[0] != int32(10) || values[1] != "Bob" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestBaseTargeterUnknownColumn(t *testing.T) {
	bt := NewBaseTargeter[testAlias](
		FieldAccess[testAlias]{Name: "id", Target: func() any { return new(int32) }},
	)

	_, err := bt.Targets([]string{"missing"})
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(UnknownColumnError); !ok {
		t.Fatalf("unexpected error type: %T", err)
	}
}
