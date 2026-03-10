package postgres

import (
	"fmt"
	"strings"

	"github.com/yaroher/ratel/pkg/migrate"
)

func (p *Planner) planAddTrigger(c migrate.AddTrigger) ([]migrate.PlannedChange, error) {
	t := c.T
	var sb strings.Builder

	fmt.Fprintf(&sb, "CREATE TRIGGER %s", q(t.Name))

	if t.Timing != "" {
		sb.WriteString(" " + strings.ToUpper(t.Timing))
	}

	if len(t.Events) > 0 {
		evts := make([]string, len(t.Events))
		for i, e := range t.Events {
			evts[i] = strings.ToUpper(e)
		}
		sb.WriteString(" " + strings.Join(evts, " OR "))
	}

	fmt.Fprintf(&sb, " ON %s", c.Table)

	if t.ForEachRow {
		sb.WriteString(" FOR EACH ROW")
	} else {
		sb.WriteString(" FOR EACH STATEMENT")
	}

	if t.When != "" {
		fmt.Fprintf(&sb, " WHEN (%s)", t.When)
	}

	// Build EXECUTE FUNCTION call
	funcCall := t.Function
	if len(t.Args) > 0 {
		funcCall += "(" + strings.Join(t.Args, ", ") + ")"
	} else {
		funcCall += "()"
	}
	fmt.Fprintf(&sb, " EXECUTE FUNCTION %s", funcCall)

	rev := fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", q(t.Name), c.Table)
	return []migrate.PlannedChange{pc(sb.String(), "add trigger "+t.Name, rev)}, nil
}

func (p *Planner) planDropTrigger(c migrate.DropTrigger) ([]migrate.PlannedChange, error) {
	t := c.T
	sql := fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", q(t.Name), c.Table)
	return []migrate.PlannedChange{pc(sql, "drop trigger "+t.Name, "")}, nil
}

func (p *Planner) planModifyTrigger(c migrate.ModifyTrigger) ([]migrate.PlannedChange, error) {
	// PostgreSQL does not support ALTER TRIGGER for most changes — drop and recreate.
	drop, err := p.planDropTrigger(migrate.DropTrigger{Table: c.Table, T: c.From})
	if err != nil {
		return nil, err
	}
	add, err := p.planAddTrigger(migrate.AddTrigger{Table: c.Table, T: c.To})
	if err != nil {
		return nil, err
	}
	return append(drop, add...), nil
}
