package constraint

import (
	"strings"

	"github.com/yaroher/ratel/common/types"
)

// TableConstraint builds table-level constraints.
type TableConstraint interface {
	types.Buildable
	tableConstraint()
}

type PrimaryKeyConstraint struct {
	name    string
	columns []string
}

func PrimaryKey(columns ...string) *PrimaryKeyConstraint {
	return &PrimaryKeyConstraint{columns: columns}
}

func NamedPrimaryKey(name string, columns ...string) *PrimaryKeyConstraint {
	return &PrimaryKeyConstraint{name: name, columns: columns}
}

func (c *PrimaryKeyConstraint) tableConstraint() {}

func (c *PrimaryKeyConstraint) Build() (string, []any) {
	var b strings.Builder
	paramIndex := 1
	args := make([]any, 0)
	c.AddToBuilder(&b, "", &paramIndex, &args)
	return b.String(), nil
}

func (c *PrimaryKeyConstraint) AddToBuilder(buf *strings.Builder, _ string, _ *int, _ *[]any) {
	if c.name != "" {
		buf.WriteString("CONSTRAINT ")
		buf.WriteString(c.name)
		buf.WriteByte(' ')
	}
	buf.WriteString("PRIMARY KEY (")
	for i, col := range c.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(col)
	}
	buf.WriteByte(')')
}

type UniqueConstraint struct {
	name    string
	columns []string
}

func Unique(columns ...string) *UniqueConstraint {
	return &UniqueConstraint{columns: columns}
}

func NamedUnique(name string, columns ...string) *UniqueConstraint {
	return &UniqueConstraint{name: name, columns: columns}
}

func (c *UniqueConstraint) tableConstraint() {}

func (c *UniqueConstraint) Build() (string, []any) {
	var b strings.Builder
	paramIndex := 1
	args := make([]any, 0)
	c.AddToBuilder(&b, "", &paramIndex, &args)
	return b.String(), nil
}

func (c *UniqueConstraint) AddToBuilder(buf *strings.Builder, _ string, _ *int, _ *[]any) {
	if c.name != "" {
		buf.WriteString("CONSTRAINT ")
		buf.WriteString(c.name)
		buf.WriteByte(' ')
	}
	buf.WriteString("UNIQUE (")
	for i, col := range c.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(col)
	}
	buf.WriteByte(')')
}

type ForeignKeyConstraint struct {
	name       string
	columns    []string
	refTable   string
	refColumns []string
	onDelete   string
	onUpdate   string
}

func ForeignKey(columns []string, refTable string, refColumns []string) *ForeignKeyConstraint {
	return &ForeignKeyConstraint{
		columns:    columns,
		refTable:   refTable,
		refColumns: refColumns,
	}
}

func NamedForeignKey(name string, columns []string, refTable string, refColumns []string) *ForeignKeyConstraint {
	return &ForeignKeyConstraint{
		name:       name,
		columns:    columns,
		refTable:   refTable,
		refColumns: refColumns,
	}
}

func (c *ForeignKeyConstraint) OnDelete(action string) *ForeignKeyConstraint {
	c.onDelete = action
	return c
}

func (c *ForeignKeyConstraint) OnUpdate(action string) *ForeignKeyConstraint {
	c.onUpdate = action
	return c
}

func (c *ForeignKeyConstraint) tableConstraint() {}

func (c *ForeignKeyConstraint) Build() (string, []any) {
	var b strings.Builder
	paramIndex := 1
	args := make([]any, 0)
	c.AddToBuilder(&b, "", &paramIndex, &args)
	return b.String(), nil
}

func (c *ForeignKeyConstraint) AddToBuilder(buf *strings.Builder, _ string, _ *int, _ *[]any) {
	if c.name != "" {
		buf.WriteString("CONSTRAINT ")
		buf.WriteString(c.name)
		buf.WriteByte(' ')
	}
	buf.WriteString("FOREIGN KEY (")
	for i, col := range c.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(col)
	}
	buf.WriteString(") REFERENCES ")
	buf.WriteString(c.refTable)
	if len(c.refColumns) > 0 {
		buf.WriteByte('(')
		for i, col := range c.refColumns {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(col)
		}
		buf.WriteByte(')')
	}
	if c.onDelete != "" {
		buf.WriteString(" ON DELETE ")
		buf.WriteString(c.onDelete)
	}
	if c.onUpdate != "" {
		buf.WriteString(" ON UPDATE ")
		buf.WriteString(c.onUpdate)
	}
}

type CheckConstraint struct {
	name string
	expr string
}

func Check(expr string) *CheckConstraint {
	return &CheckConstraint{expr: expr}
}

func NamedCheck(name, expr string) *CheckConstraint {
	return &CheckConstraint{name: name, expr: expr}
}

func (c *CheckConstraint) tableConstraint() {}

func (c *CheckConstraint) Build() (string, []any) {
	var b strings.Builder
	paramIndex := 1
	args := make([]any, 0)
	c.AddToBuilder(&b, "", &paramIndex, &args)
	return b.String(), nil
}

func (c *CheckConstraint) AddToBuilder(buf *strings.Builder, _ string, _ *int, _ *[]any) {
	if c.name != "" {
		buf.WriteString("CONSTRAINT ")
		buf.WriteString(c.name)
		buf.WriteByte(' ')
	}
	buf.WriteString("CHECK (")
	buf.WriteString(c.expr)
	buf.WriteByte(')')
}
