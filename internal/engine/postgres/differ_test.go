package postgres

import (
	"testing"

	"github.com/yaroher/ratel/pkg/migrate"
)

func newDiffer() *Differ { return &Differ{} }

// helper: wrap tables into a single-schema SchemaState.
func stateWithTables(tables ...migrate.Table) *migrate.SchemaState {
	return &migrate.SchemaState{
		Schemas: []migrate.Schema{
			{Name: "public", Tables: tables},
		},
	}
}

// ---- TestDiffAddTable ----

func TestDiffAddTable(t *testing.T) {
	current := stateWithTables()
	desired := stateWithTables(migrate.Table{
		Name:   "users",
		Schema: "public",
		Columns: []migrate.Column{
			{Name: "id", Type: "bigint", Nullable: false},
		},
	})

	d := newDiffer()
	changes, err := d.Diff(current, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var added []migrate.AddTable
	for _, c := range changes {
		if at, ok := c.(migrate.AddTable); ok {
			added = append(added, at)
		}
	}
	if len(added) != 1 {
		t.Fatalf("expected 1 AddTable, got %d", len(added))
	}
	if added[0].T.Name != "users" {
		t.Errorf("expected table name 'users', got %q", added[0].T.Name)
	}
}

// ---- TestDiffDropTable ----

func TestDiffDropTable(t *testing.T) {
	current := stateWithTables(migrate.Table{
		Name:   "old_table",
		Schema: "public",
	})
	desired := stateWithTables()

	d := newDiffer()
	changes, err := d.Diff(current, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var dropped []migrate.DropTable
	for _, c := range changes {
		if dt, ok := c.(migrate.DropTable); ok {
			dropped = append(dropped, dt)
		}
	}
	if len(dropped) != 1 {
		t.Fatalf("expected 1 DropTable, got %d", len(dropped))
	}
	if dropped[0].T.Name != "old_table" {
		t.Errorf("expected table name 'old_table', got %q", dropped[0].T.Name)
	}
}

// ---- TestDiffModifyTable ----

func TestDiffModifyTable(t *testing.T) {
	baseTable := migrate.Table{
		Name:   "products",
		Schema: "public",
		Columns: []migrate.Column{
			{Name: "id", Type: "bigint", Nullable: false},
			{Name: "price", Type: "numeric", Nullable: false},
			{Name: "name", Type: "text", Nullable: true},
		},
	}

	// desired: add "description", drop "price", change "name" type to varchar(255)
	desiredTable := migrate.Table{
		Name:   "products",
		Schema: "public",
		Columns: []migrate.Column{
			{Name: "id", Type: "bigint", Nullable: false},
			{Name: "name", Type: "varchar(255)", Nullable: true}, // type changed
			{Name: "description", Type: "text", Nullable: true},  // added
			// "price" removed
		},
	}

	current := stateWithTables(baseTable)
	desired := stateWithTables(desiredTable)

	d := newDiffer()
	changes, err := d.Diff(current, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect exactly one ModifyTable wrapping sub-changes
	var modifies []migrate.ModifyTable
	for _, c := range changes {
		if mt, ok := c.(migrate.ModifyTable); ok {
			modifies = append(modifies, mt)
		}
	}
	if len(modifies) != 1 {
		t.Fatalf("expected 1 ModifyTable, got %d", len(modifies))
	}

	mt := modifies[0]

	var addCols, dropCols, modifyCols int
	for _, sub := range mt.Changes {
		switch sub.(type) {
		case migrate.AddColumn:
			addCols++
		case migrate.DropColumn:
			dropCols++
		case migrate.ModifyColumn:
			modifyCols++
		}
	}

	if addCols != 1 {
		t.Errorf("expected 1 AddColumn, got %d", addCols)
	}
	if dropCols != 1 {
		t.Errorf("expected 1 DropColumn, got %d", dropCols)
	}
	if modifyCols != 1 {
		t.Errorf("expected 1 ModifyColumn, got %d", modifyCols)
	}
}

// ---- TestDiffRLS ----

func TestDiffRLS(t *testing.T) {
	t.Run("EnableRLS", func(t *testing.T) {
		current := stateWithTables(migrate.Table{Name: "secure", Schema: "public", RLSEnabled: false})
		desired := stateWithTables(migrate.Table{Name: "secure", Schema: "public", RLSEnabled: true})

		d := newDiffer()
		changes, err := d.Diff(current, desired)
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, c := range changes {
			if mt, ok := c.(migrate.ModifyTable); ok {
				for _, sub := range mt.Changes {
					if _, ok := sub.(migrate.EnableRLS); ok {
						found = true
					}
				}
			}
		}
		if !found {
			t.Error("expected EnableRLS change inside ModifyTable")
		}
	})

	t.Run("DisableRLS", func(t *testing.T) {
		current := stateWithTables(migrate.Table{Name: "secure", Schema: "public", RLSEnabled: true})
		desired := stateWithTables(migrate.Table{Name: "secure", Schema: "public", RLSEnabled: false})

		d := newDiffer()
		changes, err := d.Diff(current, desired)
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, c := range changes {
			if mt, ok := c.(migrate.ModifyTable); ok {
				for _, sub := range mt.Changes {
					if _, ok := sub.(migrate.DisableRLS); ok {
						found = true
					}
				}
			}
		}
		if !found {
			t.Error("expected DisableRLS change inside ModifyTable")
		}
	})

	t.Run("AddPolicy", func(t *testing.T) {
		policy := migrate.Policy{
			Name:       "user_isolation",
			Permissive: true,
			Command:    "ALL",
			Using:      "user_id = current_user_id()",
		}
		current := stateWithTables(migrate.Table{Name: "docs", Schema: "public", RLSEnabled: true})
		desired := stateWithTables(migrate.Table{Name: "docs", Schema: "public", RLSEnabled: true, Policies: []migrate.Policy{policy}})

		d := newDiffer()
		changes, err := d.Diff(current, desired)
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, c := range changes {
			if mt, ok := c.(migrate.ModifyTable); ok {
				for _, sub := range mt.Changes {
					if ap, ok := sub.(migrate.AddPolicy); ok {
						if ap.P.Name == "user_isolation" {
							found = true
						}
					}
				}
			}
		}
		if !found {
			t.Error("expected AddPolicy 'user_isolation' inside ModifyTable")
		}
	})

	t.Run("DropPolicy", func(t *testing.T) {
		policy := migrate.Policy{
			Name:       "old_policy",
			Permissive: true,
			Command:    "SELECT",
		}
		current := stateWithTables(migrate.Table{Name: "docs", Schema: "public", RLSEnabled: true, Policies: []migrate.Policy{policy}})
		desired := stateWithTables(migrate.Table{Name: "docs", Schema: "public", RLSEnabled: true})

		d := newDiffer()
		changes, err := d.Diff(current, desired)
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, c := range changes {
			if mt, ok := c.(migrate.ModifyTable); ok {
				for _, sub := range mt.Changes {
					if dp, ok := sub.(migrate.DropPolicy); ok {
						if dp.P.Name == "old_policy" {
							found = true
						}
					}
				}
			}
		}
		if !found {
			t.Error("expected DropPolicy 'old_policy' inside ModifyTable")
		}
	})
}

// ---- TestDiffExtensions ----

func TestDiffExtensions(t *testing.T) {
	t.Run("AddExtension", func(t *testing.T) {
		current := &migrate.SchemaState{
			Schemas: []migrate.Schema{{Name: "public"}},
		}
		desired := &migrate.SchemaState{
			Schemas: []migrate.Schema{{
				Name:       "public",
				Extensions: []migrate.Extension{{Name: "pgcrypto", Schema: "public"}},
			}},
		}

		d := newDiffer()
		changes, err := d.Diff(current, desired)
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, c := range changes {
			if ae, ok := c.(migrate.AddExtension); ok {
				if ae.E.Name == "pgcrypto" {
					found = true
				}
			}
		}
		if !found {
			t.Error("expected AddExtension 'pgcrypto'")
		}
	})

	t.Run("DropExtension", func(t *testing.T) {
		current := &migrate.SchemaState{
			Schemas: []migrate.Schema{{
				Name:       "public",
				Extensions: []migrate.Extension{{Name: "uuid-ossp", Schema: "public"}},
			}},
		}
		desired := &migrate.SchemaState{
			Schemas: []migrate.Schema{{Name: "public"}},
		}

		d := newDiffer()
		changes, err := d.Diff(current, desired)
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, c := range changes {
			if de, ok := c.(migrate.DropExtension); ok {
				if de.E.Name == "uuid-ossp" {
					found = true
				}
			}
		}
		if !found {
			t.Error("expected DropExtension 'uuid-ossp'")
		}
	})
}
