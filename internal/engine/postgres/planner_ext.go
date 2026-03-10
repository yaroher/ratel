package postgres

import (
	"fmt"
	"strings"

	"github.com/yaroher/ratel/pkg/migrate"
)

func (p *Planner) planAddExtension(c migrate.AddExtension) ([]migrate.PlannedChange, error) {
	ext := c.E
	var sb strings.Builder

	fmt.Fprintf(&sb, "CREATE EXTENSION IF NOT EXISTS %s", q(ext.Name))

	if ext.Schema != "" {
		fmt.Fprintf(&sb, " SCHEMA %s", q(ext.Schema))
	}

	if ext.Version != "" {
		fmt.Fprintf(&sb, " VERSION '%s'", strings.ReplaceAll(ext.Version, "'", "''"))
	}

	rev := fmt.Sprintf("DROP EXTENSION IF EXISTS %s", q(ext.Name))
	return []migrate.PlannedChange{pc(sb.String(), "add extension "+ext.Name, rev)}, nil
}

func (p *Planner) planDropExtension(c migrate.DropExtension) ([]migrate.PlannedChange, error) {
	ext := c.E
	sql := fmt.Sprintf("DROP EXTENSION IF EXISTS %s", q(ext.Name))
	rev := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", q(ext.Name))
	return []migrate.PlannedChange{pc(sql, "drop extension "+ext.Name, rev)}, nil
}
