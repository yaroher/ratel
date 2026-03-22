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

// ---- TestDiffAddSchemaWithTables ----

func TestDiffAddSchemaWithTables(t *testing.T) {
	current := &migrate.SchemaState{
		Schemas: []migrate.Schema{{Name: "public"}},
	}
	desired := &migrate.SchemaState{
		Schemas: []migrate.Schema{
			{Name: "public"},
			{
				Name: "store",
				Tables: []migrate.Table{
					{
						Name:   "users",
						Schema: "store",
						Columns: []migrate.Column{
							{Name: "id", Type: "bigint", Nullable: false},
							{Name: "email", Type: "text", Nullable: false},
						},
					},
					{
						Name:   "orders",
						Schema: "store",
						Columns: []migrate.Column{
							{Name: "id", Type: "bigint", Nullable: false},
						},
					},
				},
			},
		},
	}

	d := newDiffer()
	changes, err := d.Diff(current, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var addedSchemas []string
	var addedTables []string
	for _, c := range changes {
		switch v := c.(type) {
		case migrate.AddSchema:
			addedSchemas = append(addedSchemas, v.S.Name)
		case migrate.AddTable:
			addedTables = append(addedTables, v.T.Schema+"."+v.T.Name)
		}
	}

	if len(addedSchemas) != 1 || addedSchemas[0] != "store" {
		t.Errorf("expected AddSchema 'store', got %v", addedSchemas)
	}
	if len(addedTables) != 2 {
		t.Fatalf("expected 2 AddTable changes for tables in new schema, got %d: %v", len(addedTables), addedTables)
	}
}

// ---- TestDiffDropSchemaWithContents ----

func TestDiffDropSchemaWithContents(t *testing.T) {
	current := &migrate.SchemaState{
		Schemas: []migrate.Schema{
			{Name: "public"},
			{
				Name: "auth",
				Tables: []migrate.Table{
					{Name: "users", Schema: "auth", Columns: []migrate.Column{{Name: "id", Type: "text"}}},
					{Name: "sessions", Schema: "auth", Columns: []migrate.Column{{Name: "id", Type: "text"}}},
				},
			},
		},
	}
	desired := &migrate.SchemaState{
		Schemas: []migrate.Schema{{Name: "public"}},
	}

	d := newDiffer()
	changes, err := d.Diff(current, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect: DropTable for both tables BEFORE DropSchema
	var order []string
	for _, c := range changes {
		switch v := c.(type) {
		case migrate.DropTable:
			order = append(order, "drop_table:"+v.T.Name)
		case migrate.DropSchema:
			order = append(order, "drop_schema:"+v.S.Name)
		}
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 changes (2 DropTable + 1 DropSchema), got %d: %v", len(order), order)
	}
	// DropSchema must be last
	if order[len(order)-1] != "drop_schema:auth" {
		t.Errorf("DropSchema should be last, got order: %v", order)
	}
}

// ---- TestDiffAddTablesTopologicalOrder ----

func TestDiffAddTablesTopologicalOrder(t *testing.T) {
	// refresh_tokens → sessions → users: tables must be created in dependency order.
	current := stateWithTables()
	desired := stateWithTables(
		// Deliberately listed in wrong order to verify sorting.
		migrate.Table{
			Name:   "refresh_tokens",
			Schema: "public",
			Columns: []migrate.Column{
				{Name: "id", Type: "text", Nullable: false},
				{Name: "session_id", Type: "text", Nullable: false},
			},
			ForeignKeys: []migrate.ForeignKey{
				{Name: "fk_session", Columns: []string{"session_id"}, RefTable: "sessions", RefSchema: "public", RefColumns: []string{"id"}},
			},
		},
		migrate.Table{
			Name:   "users",
			Schema: "public",
			Columns: []migrate.Column{
				{Name: "id", Type: "text", Nullable: false},
			},
		},
		migrate.Table{
			Name:   "sessions",
			Schema: "public",
			Columns: []migrate.Column{
				{Name: "id", Type: "text", Nullable: false},
				{Name: "user_id", Type: "text", Nullable: false},
			},
			ForeignKeys: []migrate.ForeignKey{
				{Name: "fk_user", Columns: []string{"user_id"}, RefTable: "users", RefSchema: "public", RefColumns: []string{"id"}},
			},
		},
	)

	d := newDiffer()
	changes, err := d.Diff(current, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var order []string
	for _, c := range changes {
		if at, ok := c.(migrate.AddTable); ok {
			order = append(order, at.T.Name)
		}
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 AddTable, got %d: %v", len(order), order)
	}

	// users must come before sessions, sessions before refresh_tokens
	pos := make(map[string]int)
	for i, name := range order {
		pos[name] = i
	}
	if pos["users"] >= pos["sessions"] {
		t.Errorf("users (pos %d) must come before sessions (pos %d), got order: %v", pos["users"], pos["sessions"], order)
	}
	if pos["sessions"] >= pos["refresh_tokens"] {
		t.Errorf("sessions (pos %d) must come before refresh_tokens (pos %d), got order: %v", pos["sessions"], pos["refresh_tokens"], order)
	}
}

// ---- TestDiffCrossSchemaFKOrder ----

func TestDiffCrossSchemaFKOrder(t *testing.T) {
	// komeet.message_reports has FK → auth.users.
	// Both schemas are new. All CREATE SCHEMA must come before any CREATE TABLE,
	// and auth.users must come before komeet.message_reports.
	current := &migrate.SchemaState{
		Schemas: []migrate.Schema{{Name: "public"}},
	}
	desired := &migrate.SchemaState{
		Schemas: []migrate.Schema{
			{Name: "public"},
			{
				Name: "komeet",
				Tables: []migrate.Table{
					{
						Name:   "message_reports",
						Schema: "komeet",
						Columns: []migrate.Column{
							{Name: "id", Type: "bigint", Nullable: false},
							{Name: "reporter_id", Type: "bigint", Nullable: false},
						},
						ForeignKeys: []migrate.ForeignKey{
							{Name: "fk_reporter", Columns: []string{"reporter_id"}, RefTable: "users", RefSchema: "auth", RefColumns: []string{"id"}},
						},
					},
				},
			},
			{
				Name: "auth",
				Tables: []migrate.Table{
					{
						Name:   "users",
						Schema: "auth",
						Columns: []migrate.Column{
							{Name: "id", Type: "bigint", Nullable: false},
							{Name: "email", Type: "text", Nullable: false},
						},
					},
				},
			},
		},
	}

	d := newDiffer()
	changes, err := d.Diff(current, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var order []string
	for _, c := range changes {
		switch v := c.(type) {
		case migrate.AddSchema:
			order = append(order, "schema:"+v.S.Name)
		case migrate.AddTable:
			order = append(order, "table:"+v.T.Schema+"."+v.T.Name)
		}
	}

	// Both schemas must appear before any table.
	schemaEnd := 0
	for i, entry := range order {
		if entry[:6] == "schema" {
			schemaEnd = i + 1
		}
	}
	for i, entry := range order {
		if entry[:5] == "table" && i < schemaEnd {
			t.Errorf("table %s (pos %d) appears before last schema (pos %d); order: %v", entry, i, schemaEnd-1, order)
		}
	}

	// auth.users must come before komeet.message_reports.
	pos := make(map[string]int)
	for i, entry := range order {
		pos[entry] = i
	}
	if pos["table:auth.users"] >= pos["table:komeet.message_reports"] {
		t.Errorf("auth.users (pos %d) must come before komeet.message_reports (pos %d); order: %v",
			pos["table:auth.users"], pos["table:komeet.message_reports"], order)
	}
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
