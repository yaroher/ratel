package ddl

import (
	"strings"

	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/sqlbuild/ddl/constraint"
)

// CreateTable builds CREATE TABLE statements.
type CreateTable struct {
	name        types.TableAlias
	ifNotExists bool
	columns     []types.ColumnAlias
	constraints []constraint.TableConstraint
}

func (c *CreateTable) ddlQuery() {}

// CreateTableStmt creates a new CREATE TABLE builder.
func CreateTableStmt(name types.TableAlias) *CreateTable {
	return &CreateTable{name: name}
}

// IfNotExists adds IF NOT EXISTS.
func (c *CreateTable) IfNotExists() *CreateTable {
	c.ifNotExists = true
	return c
}

// Column adds a column definition.
func (c *CreateTable) Column(column types.ColumnAlias) *CreateTable {
	c.columns = append(c.columns, column)
	return c
}

// Columns adds multiple column definitions.
func (c *CreateTable) Columns(columns ...types.ColumnAlias) *CreateTable {
	c.columns = append(c.columns, columns...)
	return c
}

// Constraint adds a table constraint.
func (c *CreateTable) Constraint(constraint constraint.TableConstraint) *CreateTable {
	c.constraints = append(c.constraints, constraint)
	return c
}

// Build returns SQL and args (DDL has no args).
func (c *CreateTable) Build() (string, []any) {
	paramIndex := 1
	buf := sbPool.Get().(*strings.Builder)
	buf.Reset()
	buf.Grow(96 + len(c.columns)*32 + len(c.constraints)*32)

	args := make([]any, 0)
	c.AddToBuilder(buf, "", &paramIndex, &args)
	sql := buf.String()
	sbPool.Put(buf)
	return sql, args
}

func (c *CreateTable) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("CREATE TABLE ")
	if c.ifNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	buf.WriteString(c.name.String())
	buf.WriteString(" (")
	first := true
	for _, col := range c.columns {
		if any(col) == nil {
			continue
		}
		if !first {
			buf.WriteString(", ")
		}
		if builder, ok := any(col).(types.Buildable); ok {
			builder.AddToBuilder(buf, ta, paramIndex, args)
		} else {
			buf.WriteString(col.String())
		}
		first = false
	}
	for _, constraint := range c.constraints {
		if constraint == nil {
			continue
		}
		if !first {
			buf.WriteString(", ")
		}
		constraint.AddToBuilder(buf, ta, paramIndex, args)
		first = false
	}
	buf.WriteString(");")
}

type alterAction interface {
	AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any)
}

type alterActionRaw string

func (a alterActionRaw) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString(string(a))
}

type alterActionColumn struct {
	prefix string
	column types.Buildable
}

func (a alterActionColumn) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString(a.prefix)
	a.column.AddToBuilder(buf, ta, paramIndex, args)
}

type alterActionConstraint struct {
	prefix     string
	constraint constraint.TableConstraint
}

func (a alterActionConstraint) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString(a.prefix)
	a.constraint.AddToBuilder(buf, ta, paramIndex, args)
}

type AlterTable struct {
	name     types.TableAlias
	ifExists bool
	actions  []alterAction
}

func (a *AlterTable) ddlQuery() {}
func AlterTableStmt(name types.TableAlias) *AlterTable {
	return &AlterTable{name: name}
}

func (a *AlterTable) IfExists() *AlterTable {
	a.ifExists = true
	return a
}

func (a *AlterTable) AddColumn(column types.Buildable) *AlterTable {
	if column != nil {
		a.actions = append(a.actions, alterActionColumn{
			prefix: "ADD COLUMN ",
			column: column,
		})
	}
	return a
}

func (a *AlterTable) AddColumnIfNotExists(column types.Buildable) *AlterTable {
	if column != nil {
		a.actions = append(a.actions, alterActionColumn{
			prefix: "ADD COLUMN IF NOT EXISTS ",
			column: column,
		})
	}
	return a
}

func (a *AlterTable) DropColumn(name string, ifExists bool) *AlterTable {
	action := "DROP COLUMN "
	if ifExists {
		action += "IF EXISTS "
	}
	action += name
	a.actions = append(a.actions, alterActionRaw(action))
	return a
}

func (a *AlterTable) RenameColumn(oldName, newName string) *AlterTable {
	a.actions = append(a.actions, alterActionRaw("RENAME COLUMN "+oldName+" TO "+newName))
	return a
}

func (a *AlterTable) RenameTo(newName string) *AlterTable {
	a.actions = append(a.actions, alterActionRaw("RENAME TO "+newName))
	return a
}

func (a *AlterTable) AddConstraint(constraint constraint.TableConstraint) *AlterTable {
	if constraint != nil {
		a.actions = append(a.actions, alterActionConstraint{
			prefix:     "ADD ",
			constraint: constraint,
		})
	}
	return a
}

func (a *AlterTable) DropConstraint(name string, ifExists bool) *AlterTable {
	action := "DROP CONSTRAINT "
	if ifExists {
		action += "IF EXISTS "
	}
	action += name
	a.actions = append(a.actions, alterActionRaw(action))
	return a
}

func (a *AlterTable) SetColumnType(name, dataType string) *AlterTable {
	a.actions = append(a.actions, alterActionRaw("ALTER COLUMN "+name+" TYPE "+dataType))
	return a
}

func (a *AlterTable) Build() (string, []any) {
	paramIndex := 1
	buf := sbPool.Get().(*strings.Builder)
	buf.Reset()
	buf.Grow(96 + len(a.actions)*32)

	args := make([]any, 0)
	a.AddToBuilder(buf, "", &paramIndex, &args)
	sql := buf.String()
	sbPool.Put(buf)
	return sql, args
}

func (a *AlterTable) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("ALTER TABLE ")
	if a.ifExists {
		buf.WriteString("IF EXISTS ")
	}
	buf.WriteString(a.name.String())
	if len(a.actions) > 0 {
		buf.WriteByte(' ')
		for i, action := range a.actions {
			if i > 0 {
				buf.WriteString(", ")
			}
			action.AddToBuilder(buf, ta, paramIndex, args)
		}
	}
	buf.WriteString(";")
}

type DropTable struct {
	names    []types.TableAlias
	ifExists bool
	cascade  bool
	restrict bool
}

func (d *DropTable) ddlQuery() {}

func DropTableStmt(names ...types.TableAlias) *DropTable {
	return &DropTable{names: names}
}

func (d *DropTable) IfExists() *DropTable {
	d.ifExists = true
	return d
}

func (d *DropTable) Cascade() *DropTable {
	d.cascade = true
	d.restrict = false
	return d
}

func (d *DropTable) Restrict() *DropTable {
	d.restrict = true
	d.cascade = false
	return d
}

func (d *DropTable) Build() (string, []any) {
	paramIndex := 1
	buf := sbPool.Get().(*strings.Builder)
	buf.Reset()
	buf.Grow(64 + len(d.names)*16)

	args := make([]any, 0)
	d.AddToBuilder(buf, "", &paramIndex, &args)
	sql := buf.String()
	sbPool.Put(buf)
	return sql, args
}

func (d *DropTable) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("DROP TABLE ")
	if d.ifExists {
		buf.WriteString("IF EXISTS ")
	}
	for i, name := range d.names {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(name.String())
	}
	if d.cascade {
		buf.WriteString(" CASCADE")
	} else if d.restrict {
		buf.WriteString(" RESTRICT")
	}
	buf.WriteString(";")
}
