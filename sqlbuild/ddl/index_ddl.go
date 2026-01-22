package ddl

import (
	"strings"

	"github.com/yaroher/ratel/common/types"
)

// CreateIndex builds CREATE INDEX statements.
type CreateIndex struct {
	name         types.TableAlias
	table        types.TableAlias
	ifNotExists  bool
	unique       bool
	concurrently bool
	columns      []types.ColumnAlias
	using        string
	where        string
}

// CreateIndexStmt creates a new CREATE INDEX builder.
func CreateIndexStmt(name types.TableAlias, table types.TableAlias) *CreateIndex {
	return &CreateIndex{name: name, table: table}
}

// IfNotExists adds IF NOT EXISTS.
func (c *CreateIndex) IfNotExists() *CreateIndex {
	c.ifNotExists = true
	return c
}

// Unique adds UNIQUE.
func (c *CreateIndex) Unique() *CreateIndex {
	c.unique = true
	return c
}

// Concurrently adds CONCURRENTLY.
func (c *CreateIndex) Concurrently() *CreateIndex {
	c.concurrently = true
	return c
}

// Using sets index method (raw SQL, e.g. "btree").
func (c *CreateIndex) Using(method string) *CreateIndex {
	c.using = method
	return c
}

// Columns adds index columns/expressions (raw SQL).
func (c *CreateIndex) Columns(columns ...types.ColumnAlias) *CreateIndex {
	c.columns = append(c.columns, columns...)
	return c
}

// Where adds WHERE predicate (partial index).
func (c *CreateIndex) Where(expr string) *CreateIndex {
	c.where = expr
	return c
}

// Build returns SQL and args (DDL has no args).
func (c *CreateIndex) Build() (string, []any) {
	paramIndex := 1
	buf := sbPool.Get().(*strings.Builder)
	buf.Reset()
	buf.Grow(96 + len(c.columns)*16 + len(c.where))

	args := make([]any, 0)
	c.AddToBuilder(buf, "", &paramIndex, &args)
	sql := buf.String()
	sbPool.Put(buf)
	return sql, args
}

func (c *CreateIndex) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("CREATE ")
	if c.unique {
		buf.WriteString("UNIQUE ")
	}
	buf.WriteString("INDEX ")
	if c.concurrently {
		buf.WriteString("CONCURRENTLY ")
	}
	if c.ifNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	buf.WriteString(c.name.String())
	buf.WriteString(" ON ")
	buf.WriteString(c.table.String())
	if c.using != "" {
		buf.WriteString(" USING ")
		buf.WriteString(c.using)
	}
	buf.WriteString(" (")
	for i, col := range c.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		if builder, ok := any(col).(types.Buildable); ok {
			builder.AddToBuilder(buf, ta, paramIndex, args)
		} else {
			buf.WriteString(col.String())
		}
	}
	buf.WriteByte(')')
	if c.where != "" {
		buf.WriteString(" WHERE ")
		buf.WriteString(c.where)
	}
	buf.WriteString(";")
}

// DropIndex builds DROP INDEX statements.
type DropIndex struct {
	names        []types.TableAlias
	ifExists     bool
	cascade      bool
	restrict     bool
	concurrently bool
}

// DropIndexStmt creates a new DROP INDEX builder.
func DropIndexStmt(names ...types.TableAlias) *DropIndex {
	return &DropIndex{names: names}
}

// IfExists adds IF EXISTS.
func (d *DropIndex) IfExists() *DropIndex {
	d.ifExists = true
	return d
}

// Cascade adds CASCADE.
func (d *DropIndex) Cascade() *DropIndex {
	d.cascade = true
	d.restrict = false
	return d
}

// Restrict adds RESTRICT.
func (d *DropIndex) Restrict() *DropIndex {
	d.restrict = true
	d.cascade = false
	return d
}

// Concurrently adds CONCURRENTLY.
func (d *DropIndex) Concurrently() *DropIndex {
	d.concurrently = true
	return d
}

// Build returns SQL and args (DDL has no args).
func (d *DropIndex) Build() (string, []any) {
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

func (d *DropIndex) AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("DROP INDEX ")
	if d.concurrently {
		buf.WriteString("CONCURRENTLY ")
	}
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
