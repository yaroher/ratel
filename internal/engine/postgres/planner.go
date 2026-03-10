package postgres

// Based on Atlas (https://github.com/ariga/atlas) — Apache 2.0 License
// Copyright 2021-present The Atlas Authors

import (
	"context"
	"fmt"
	"strings"

	"github.com/yaroher/ratel/pkg/migrate"
)

// Planner generates PostgreSQL DDL from a list of schema changes.
type Planner struct{}

// NewPlanner returns a new Planner.
func NewPlanner() *Planner { return &Planner{} }

// Plan converts a list of changes into a migration Plan.
func (p *Planner) Plan(_ context.Context, name string, changes []migrate.Change) (*migrate.Plan, error) {
	var planned []migrate.PlannedChange
	for _, c := range changes {
		stmts, err := p.planChange(c)
		if err != nil {
			return nil, err
		}
		planned = append(planned, stmts...)
	}
	return &migrate.Plan{Name: name, Changes: planned}, nil
}

// planChange dispatches a single Change to the appropriate plan* method.
func (p *Planner) planChange(c migrate.Change) ([]migrate.PlannedChange, error) {
	switch c := c.(type) {
	case migrate.AddTable:
		return p.planAddTable(c)
	case migrate.DropTable:
		return p.planDropTable(c)
	case migrate.ModifyTable:
		return p.planModifyTable(c)
	case migrate.AddSchema:
		return p.planAddSchema(c)
	case migrate.DropSchema:
		return p.planDropSchema(c)
	// RLS
	case migrate.EnableRLS:
		return p.planEnableRLS(c)
	case migrate.DisableRLS:
		return p.planDisableRLS(c)
	case migrate.ForceRLS:
		return p.planForceRLS(c)
	case migrate.UnforceRLS:
		return p.planUnforceRLS(c)
	case migrate.AddPolicy:
		return p.planAddPolicy(c)
	case migrate.DropPolicy:
		return p.planDropPolicy(c)
	case migrate.ModifyPolicy:
		return p.planModifyPolicy(c)
	// Extensions
	case migrate.AddExtension:
		return p.planAddExtension(c)
	case migrate.DropExtension:
		return p.planDropExtension(c)
	// Triggers
	case migrate.AddTrigger:
		return p.planAddTrigger(c)
	case migrate.DropTrigger:
		return p.planDropTrigger(c)
	case migrate.ModifyTrigger:
		return p.planModifyTrigger(c)
	// Functions
	case migrate.AddFunction:
		return p.planAddFunction(c)
	case migrate.DropFunction:
		return p.planDropFunction(c)
	case migrate.ModifyFunction:
		return p.planModifyFunction(c)
	default:
		return nil, fmt.Errorf("unsupported change type: %T", c)
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

// q double-quotes an identifier, escaping embedded double-quotes.
func q(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// qualifiedName returns "schema"."name".
func qualifiedName(schema, name string) string {
	if schema == "" {
		return q(name)
	}
	return q(schema) + "." + q(name)
}

// pc is a convenience constructor for PlannedChange.
func pc(sql, comment, reverse string) migrate.PlannedChange {
	return migrate.PlannedChange{SQL: sql, Comment: comment, Reverse: reverse}
}

// ── schema ───────────────────────────────────────────────────────────────────

func (p *Planner) planAddSchema(c migrate.AddSchema) ([]migrate.PlannedChange, error) {
	sql := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", q(c.S.Name))
	rev := fmt.Sprintf("DROP SCHEMA IF EXISTS %s", q(c.S.Name))
	return []migrate.PlannedChange{pc(sql, "create schema "+c.S.Name, rev)}, nil
}

func (p *Planner) planDropSchema(c migrate.DropSchema) ([]migrate.PlannedChange, error) {
	sql := fmt.Sprintf("DROP SCHEMA IF EXISTS %s", q(c.S.Name))
	rev := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", q(c.S.Name))
	return []migrate.PlannedChange{pc(sql, "drop schema "+c.S.Name, rev)}, nil
}

// ── table ────────────────────────────────────────────────────────────────────

func (p *Planner) planAddTable(c migrate.AddTable) ([]migrate.PlannedChange, error) {
	t := c.T
	var cols []string

	for _, col := range t.Columns {
		cols = append(cols, columnDef(col))
	}

	if t.PrimaryKey != nil {
		pkCols := make([]string, len(t.PrimaryKey.Columns))
		for i, ic := range t.PrimaryKey.Columns {
			pkCols[i] = q(ic.Name)
		}
		cols = append(cols, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pkCols, ", ")))
	}

	for _, fk := range t.ForeignKeys {
		cols = append(cols, fkConstraintDef(fk))
	}

	for _, ch := range t.Checks {
		cols = append(cols, checkConstraintDef(ch))
	}

	tname := qualifiedName(t.Schema, t.Name)
	sql := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", tname, strings.Join(cols, ",\n  "))

	stmts := make([]migrate.PlannedChange, 0, 1+len(t.Indexes))
	rev := fmt.Sprintf("DROP TABLE %s", tname)
	stmts = append(stmts, pc(sql, "create table "+t.Name, rev))

	// indexes (excluding PK which is already inlined)
	for _, idx := range t.Indexes {
		idxSQL := createIndexSQL(t.Schema, t.Name, idx)
		dropIdx := fmt.Sprintf("DROP INDEX %s", qualifiedName(t.Schema, idx.Name))
		stmts = append(stmts, pc(idxSQL, "create index "+idx.Name, dropIdx))
	}

	return stmts, nil
}

func (p *Planner) planDropTable(c migrate.DropTable) ([]migrate.PlannedChange, error) {
	t := c.T
	tname := qualifiedName(t.Schema, t.Name)
	sql := fmt.Sprintf("DROP TABLE %s", tname)
	return []migrate.PlannedChange{pc(sql, "drop table "+t.Name, "")}, nil
}

func (p *Planner) planModifyTable(c migrate.ModifyTable) ([]migrate.PlannedChange, error) {
	t := c.To
	tname := qualifiedName(t.Schema, t.Name)

	var stmts []migrate.PlannedChange
	for _, sub := range c.Changes {
		switch sc := sub.(type) {
		case migrate.AddColumn:
			col := sc.C
			sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tname, columnDef(*col))
			rev := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tname, q(col.Name))
			stmts = append(stmts, pc(sql, "add column "+col.Name, rev))

		case migrate.DropColumn:
			col := sc.C
			sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tname, q(col.Name))
			stmts = append(stmts, pc(sql, "drop column "+col.Name, ""))

		case migrate.ModifyColumn:
			ss, err := planModifyColumn(tname, sc)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)

		case migrate.AddIndex:
			idx := sc.I
			sql := createIndexSQL(t.Schema, t.Name, *idx)
			rev := fmt.Sprintf("DROP INDEX %s", qualifiedName(t.Schema, idx.Name))
			stmts = append(stmts, pc(sql, "create index "+idx.Name, rev))

		case migrate.DropIndex:
			idx := sc.I
			sql := fmt.Sprintf("DROP INDEX %s", qualifiedName(t.Schema, idx.Name))
			stmts = append(stmts, pc(sql, "drop index "+idx.Name, ""))

		case migrate.ModifyIndex:
			// Drop old, create new.
			dropSQL := fmt.Sprintf("DROP INDEX %s", qualifiedName(t.Schema, sc.From.Name))
			stmts = append(stmts, pc(dropSQL, "drop index "+sc.From.Name, ""))
			createSQL := createIndexSQL(t.Schema, t.Name, *sc.To)
			rev := fmt.Sprintf("DROP INDEX %s", qualifiedName(t.Schema, sc.To.Name))
			stmts = append(stmts, pc(createSQL, "create index "+sc.To.Name, rev))

		case migrate.AddForeignKey:
			fk := sc.FK
			sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", tname, q(fk.Name), fkConstraintDef(*fk))
			rev := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tname, q(fk.Name))
			stmts = append(stmts, pc(sql, "add foreign key "+fk.Name, rev))

		case migrate.DropForeignKey:
			fk := sc.FK
			sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tname, q(fk.Name))
			stmts = append(stmts, pc(sql, "drop foreign key "+fk.Name, ""))

		case migrate.AddCheck:
			ch := sc.C
			sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", tname, q(ch.Name), checkConstraintDef(*ch))
			rev := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tname, q(ch.Name))
			stmts = append(stmts, pc(sql, "add check "+ch.Name, rev))

		case migrate.DropCheck:
			ch := sc.C
			sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tname, q(ch.Name))
			stmts = append(stmts, pc(sql, "drop check "+ch.Name, ""))

		// RLS sub-changes: delegate to top-level planners with qualified table name.
		case migrate.EnableRLS:
			ss, err := p.planEnableRLS(migrate.EnableRLS{Table: tname})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.DisableRLS:
			ss, err := p.planDisableRLS(migrate.DisableRLS{Table: tname})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.ForceRLS:
			ss, err := p.planForceRLS(migrate.ForceRLS{Table: tname})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.UnforceRLS:
			ss, err := p.planUnforceRLS(migrate.UnforceRLS{Table: tname})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.AddPolicy:
			ss, err := p.planAddPolicy(migrate.AddPolicy{Table: tname, P: sc.P})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.DropPolicy:
			ss, err := p.planDropPolicy(migrate.DropPolicy{Table: tname, P: sc.P})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.ModifyPolicy:
			ss, err := p.planModifyPolicy(migrate.ModifyPolicy{Table: tname, From: sc.From, To: sc.To})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)

		// Trigger sub-changes.
		case migrate.AddTrigger:
			ss, err := p.planAddTrigger(migrate.AddTrigger{Table: tname, T: sc.T})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.DropTrigger:
			ss, err := p.planDropTrigger(migrate.DropTrigger{Table: tname, T: sc.T})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)
		case migrate.ModifyTrigger:
			ss, err := p.planModifyTrigger(migrate.ModifyTrigger{Table: tname, From: sc.From, To: sc.To})
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ss...)

		default:
			return nil, fmt.Errorf("unsupported table sub-change: %T", sub)
		}
	}
	return stmts, nil
}

// ── column helpers ────────────────────────────────────────────────────────────

// columnDef returns the full column definition for a CREATE TABLE or ADD COLUMN.
func columnDef(col migrate.Column) string {
	var sb strings.Builder
	sb.WriteString(q(col.Name))
	sb.WriteString(" ")
	sb.WriteString(col.Type)

	if col.Identity != nil {
		fmt.Fprintf(&sb, " GENERATED %s AS IDENTITY", strings.ToUpper(col.Identity.Generation))
	} else {
		if !col.Nullable {
			sb.WriteString(" NOT NULL")
		}
		if col.Default != "" {
			sb.WriteString(" DEFAULT ")
			sb.WriteString(col.Default)
		}
	}
	return sb.String()
}

// planModifyColumn emits one or more ALTER COLUMN statements.
func planModifyColumn(tname string, mc migrate.ModifyColumn) ([]migrate.PlannedChange, error) {
	from, to := mc.From, mc.To
	var stmts []migrate.PlannedChange

	prefix := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s", tname, q(to.Name))

	// Type change
	if from.Type != to.Type {
		sql := fmt.Sprintf("%s TYPE %s", prefix, to.Type)
		stmts = append(stmts, pc(sql, "alter column type "+to.Name, ""))
	}

	// Nullability change
	if from.Nullable != to.Nullable {
		var sql string
		if to.Nullable {
			sql = fmt.Sprintf("%s DROP NOT NULL", prefix)
		} else {
			sql = fmt.Sprintf("%s SET NOT NULL", prefix)
		}
		stmts = append(stmts, pc(sql, "alter column nullability "+to.Name, ""))
	}

	// Default change
	if from.Default != to.Default {
		var sql string
		if to.Default == "" {
			sql = fmt.Sprintf("%s DROP DEFAULT", prefix)
		} else {
			sql = fmt.Sprintf("%s SET DEFAULT %s", prefix, to.Default)
		}
		stmts = append(stmts, pc(sql, "alter column default "+to.Name, ""))
	}

	return stmts, nil
}

// ── index helpers ─────────────────────────────────────────────────────────────

func createIndexSQL(schema, table string, idx migrate.Index) string {
	var sb strings.Builder
	sb.WriteString("CREATE ")
	if idx.Unique {
		sb.WriteString("UNIQUE ")
	}
	fmt.Fprintf(&sb, "INDEX %s ON %s", q(idx.Name), qualifiedName(schema, table))

	method := idx.Method
	if method == "" {
		method = "btree"
	}
	fmt.Fprintf(&sb, " USING %s", method)

	cols := make([]string, len(idx.Columns))
	for i, ic := range idx.Columns {
		col := q(ic.Name)
		if ic.Desc || strings.EqualFold(ic.Order, "DESC") {
			col += " DESC"
		}
		cols[i] = col
	}
	fmt.Fprintf(&sb, " (%s)", strings.Join(cols, ", "))

	if len(idx.Include) > 0 {
		incl := make([]string, len(idx.Include))
		for i, c := range idx.Include {
			incl[i] = q(c)
		}
		fmt.Fprintf(&sb, " INCLUDE (%s)", strings.Join(incl, ", "))
	}

	if idx.Where != "" {
		fmt.Fprintf(&sb, " WHERE %s", idx.Where)
	}

	return sb.String()
}

// ── FK / check helpers ────────────────────────────────────────────────────────

func fkConstraintDef(fk migrate.ForeignKey) string {
	cols := make([]string, len(fk.Columns))
	for i, c := range fk.Columns {
		cols[i] = q(c)
	}
	refCols := make([]string, len(fk.RefColumns))
	for i, c := range fk.RefColumns {
		refCols[i] = q(c)
	}

	def := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
		strings.Join(cols, ", "),
		qualifiedName(fk.RefSchema, fk.RefTable),
		strings.Join(refCols, ", "))

	if fk.OnUpdate != "" {
		def += " ON UPDATE " + fk.OnUpdate
	}
	if fk.OnDelete != "" {
		def += " ON DELETE " + fk.OnDelete
	}
	return def
}

func checkConstraintDef(ch migrate.Check) string {
	def := fmt.Sprintf("CHECK (%s)", ch.Expr)
	if ch.NoInherit {
		def += " NO INHERIT"
	}
	return def
}
