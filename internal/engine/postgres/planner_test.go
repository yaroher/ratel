package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/yaroher/ratel/pkg/migrate"
)

func TestPlanAddTable(t *testing.T) {
	p := NewPlanner()
	table := &migrate.Table{
		Name:   "users",
		Schema: "public",
		Columns: []migrate.Column{
			{Name: "id", Type: "bigint", Nullable: false},
			{Name: "email", Type: "text", Nullable: false},
			{Name: "created_at", Type: "timestamptz", Nullable: true, Default: "now()"},
		},
		PrimaryKey: &migrate.Index{
			Columns: []migrate.IndexColumn{{Name: "id"}},
		},
	}

	plan, err := p.Plan(context.Background(), "add_users", []migrate.Change{
		migrate.AddTable{T: table},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, `CREATE TABLE "public"."users"`)
	assertContains(t, sql, `"id" bigint NOT NULL`)
	assertContains(t, sql, `"email" text NOT NULL`)
	assertContains(t, sql, `"created_at" timestamptz DEFAULT now()`)
	assertContains(t, sql, `PRIMARY KEY ("id")`)
}

func TestPlanDropTable(t *testing.T) {
	p := NewPlanner()
	table := &migrate.Table{Name: "users", Schema: "public"}

	plan, err := p.Plan(context.Background(), "drop_users", []migrate.Change{
		migrate.DropTable{T: table},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, `DROP TABLE "public"."users"`)
}

func TestPlanAddColumn(t *testing.T) {
	p := NewPlanner()
	table := &migrate.Table{Name: "users", Schema: "app"}
	col := &migrate.Column{Name: "bio", Type: "text", Nullable: true}

	plan, err := p.Plan(context.Background(), "add_bio", []migrate.Change{
		migrate.ModifyTable{
			From:    table,
			To:      table,
			Changes: []migrate.Change{migrate.AddColumn{C: col}},
		},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, `ALTER TABLE "app"."users" ADD COLUMN`)
	assertContains(t, sql, `"bio" text`)
}

func TestPlanEnableRLS(t *testing.T) {
	p := NewPlanner()

	plan, err := p.Plan(context.Background(), "enable_rls", []migrate.Change{
		migrate.EnableRLS{Table: `"public"."orders"`},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, "ENABLE ROW LEVEL SECURITY")
}

func TestPlanAddPolicy(t *testing.T) {
	p := NewPlanner()
	pol := &migrate.Policy{
		Name:       "user_isolation",
		Permissive: true,
		Command:    "SELECT",
		Roles:      []string{"app_user"},
		Using:      "user_id = current_user_id()",
	}

	plan, err := p.Plan(context.Background(), "add_policy", []migrate.Change{
		migrate.AddPolicy{Table: `"public"."orders"`, P: pol},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, `CREATE POLICY "user_isolation"`)
	assertContains(t, sql, "AS PERMISSIVE")
	assertContains(t, sql, "FOR SELECT")
	assertContains(t, sql, `TO "app_user"`)
	assertContains(t, sql, "USING (user_id = current_user_id())")
}

func TestPlanAddExtension(t *testing.T) {
	p := NewPlanner()
	ext := &migrate.Extension{
		Name:    "pgcrypto",
		Schema:  "public",
		Version: "1.3",
	}

	plan, err := p.Plan(context.Background(), "add_ext", []migrate.Change{
		migrate.AddExtension{E: ext},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	assertContains(t, sql, `SCHEMA "public"`)
	assertContains(t, sql, `VERSION '1.3'`)
}

func TestPlanAddTrigger(t *testing.T) {
	p := NewPlanner()
	trig := &migrate.Trigger{
		Name:       "audit_insert",
		Events:     []string{"INSERT", "UPDATE"},
		Timing:     "AFTER",
		ForEachRow: true,
		Function:   "audit_fn",
	}

	plan, err := p.Plan(context.Background(), "add_trigger", []migrate.Change{
		migrate.AddTrigger{Table: `"public"."orders"`, T: trig},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, `CREATE TRIGGER "audit_insert"`)
	assertContains(t, sql, "AFTER")
	assertContains(t, sql, "INSERT OR UPDATE")
	assertContains(t, sql, "FOR EACH ROW")
	assertContains(t, sql, "EXECUTE FUNCTION audit_fn()")
}

func TestPlanAddFunction(t *testing.T) {
	p := NewPlanner()
	fn := &migrate.Function{
		Name:       "get_user",
		Schema:     "public",
		Args:       []migrate.FunctionArg{{Name: "p_id", Type: "bigint"}},
		ReturnType: "text",
		Language:   "sql",
		Body:       "SELECT email FROM users WHERE id = p_id;",
		Volatility: "STABLE",
		Security:   "INVOKER",
	}

	plan, err := p.Plan(context.Background(), "add_func", []migrate.Change{
		migrate.AddFunction{F: fn},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatal("expected at least one planned change")
	}

	sql := plan.Changes[0].SQL
	assertContains(t, sql, `CREATE OR REPLACE FUNCTION "public"."get_user"`)
	assertContains(t, sql, `"p_id" bigint`)
	assertContains(t, sql, "RETURNS text")
	assertContains(t, sql, "LANGUAGE sql")
	assertContains(t, sql, "STABLE")
	assertContains(t, sql, "SECURITY INVOKER")
	assertContains(t, sql, "$body$")
}

// assertContains fails if substr is not found in s (case-sensitive).
func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected SQL to contain %q\ngot: %s", substr, s)
	}
}
