package postgres

import (
	"fmt"
	"strings"

	"github.com/yaroher/ratel/pkg/migrate"
)

// quoteRoles quotes role names, preserving SQL keywords like PUBLIC unquoted.
func quoteRoles(roles []string) []string {
	out := make([]string, len(roles))
	for i, r := range roles {
		upper := strings.ToUpper(r)
		if upper == "PUBLIC" || upper == "CURRENT_USER" || upper == "SESSION_USER" {
			out[i] = upper
		} else {
			out[i] = q(r)
		}
	}
	return out
}

func (p *Planner) planEnableRLS(c migrate.EnableRLS) ([]migrate.PlannedChange, error) {
	sql := fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", c.Table)
	rev := fmt.Sprintf("ALTER TABLE %s DISABLE ROW LEVEL SECURITY", c.Table)
	return []migrate.PlannedChange{pc(sql, "enable RLS on "+c.Table, rev)}, nil
}

func (p *Planner) planDisableRLS(c migrate.DisableRLS) ([]migrate.PlannedChange, error) {
	sql := fmt.Sprintf("ALTER TABLE %s DISABLE ROW LEVEL SECURITY", c.Table)
	rev := fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", c.Table)
	return []migrate.PlannedChange{pc(sql, "disable RLS on "+c.Table, rev)}, nil
}

func (p *Planner) planForceRLS(c migrate.ForceRLS) ([]migrate.PlannedChange, error) {
	sql := fmt.Sprintf("ALTER TABLE %s FORCE ROW LEVEL SECURITY", c.Table)
	rev := fmt.Sprintf("ALTER TABLE %s NO FORCE ROW LEVEL SECURITY", c.Table)
	return []migrate.PlannedChange{pc(sql, "force RLS on "+c.Table, rev)}, nil
}

func (p *Planner) planUnforceRLS(c migrate.UnforceRLS) ([]migrate.PlannedChange, error) {
	sql := fmt.Sprintf("ALTER TABLE %s NO FORCE ROW LEVEL SECURITY", c.Table)
	rev := fmt.Sprintf("ALTER TABLE %s FORCE ROW LEVEL SECURITY", c.Table)
	return []migrate.PlannedChange{pc(sql, "no force RLS on "+c.Table, rev)}, nil
}

func (p *Planner) planAddPolicy(c migrate.AddPolicy) ([]migrate.PlannedChange, error) {
	pol := c.P
	var sb strings.Builder

	fmt.Fprintf(&sb, "CREATE POLICY %s ON %s", q(pol.Name), c.Table)

	if pol.Permissive {
		sb.WriteString(" AS PERMISSIVE")
	} else {
		sb.WriteString(" AS RESTRICTIVE")
	}

	if pol.Command != "" {
		fmt.Fprintf(&sb, " FOR %s", strings.ToUpper(pol.Command))
	}

	if len(pol.Roles) > 0 {
		quoted := quoteRoles(pol.Roles)
		fmt.Fprintf(&sb, " TO %s", strings.Join(quoted, ", "))
	}

	if pol.Using != "" {
		fmt.Fprintf(&sb, " USING (%s)", pol.Using)
	}

	if pol.WithCheck != "" {
		fmt.Fprintf(&sb, " WITH CHECK (%s)", pol.WithCheck)
	}

	rev := fmt.Sprintf("DROP POLICY %s ON %s", q(pol.Name), c.Table)
	return []migrate.PlannedChange{pc(sb.String(), "add policy "+pol.Name, rev)}, nil
}

func (p *Planner) planDropPolicy(c migrate.DropPolicy) ([]migrate.PlannedChange, error) {
	pol := c.P
	sql := fmt.Sprintf("DROP POLICY %s ON %s", q(pol.Name), c.Table)
	return []migrate.PlannedChange{pc(sql, "drop policy "+pol.Name, "")}, nil
}

func (p *Planner) planModifyPolicy(c migrate.ModifyPolicy) ([]migrate.PlannedChange, error) {
	pol := c.To
	var sb strings.Builder

	fmt.Fprintf(&sb, "ALTER POLICY %s ON %s", q(pol.Name), c.Table)

	if len(pol.Roles) > 0 {
		quoted := quoteRoles(pol.Roles)
		fmt.Fprintf(&sb, " TO %s", strings.Join(quoted, ", "))
	}

	if pol.Using != "" {
		fmt.Fprintf(&sb, " USING (%s)", pol.Using)
	}

	if pol.WithCheck != "" {
		fmt.Fprintf(&sb, " WITH CHECK (%s)", pol.WithCheck)
	}

	return []migrate.PlannedChange{pc(sb.String(), "modify policy "+pol.Name, "")}, nil
}
