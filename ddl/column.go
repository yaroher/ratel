package ddl

import (
	"strings"

	"github.com/yaroher/ratel/common/types"
)

// ColumnDef describes a column for CREATE/ALTER TABLE.
type ColumnDef[C types.ColumnAlias] struct {
	fa           C
	dataType     Datatype
	notNull      bool
	explicitNull bool
	unique       bool
	primaryKey   bool
	defaultValue *string
	reference    *Reference
	checkExpr    *string
}

// Reference describes a REFERENCES clause with optional actions.
type Reference struct {
	table    string
	column   string
	onDelete string
	onUpdate string
}

func (c *ColumnDef[C]) Alias() C {
	return c.fa
}

// Column creates a column definition.
func Column[C types.ColumnAlias](name C, dataType Datatype) *ColumnDef[C] {
	return &ColumnDef[C]{fa: name, dataType: dataType}
}

// NotNull sets NOT NULL constraint.
func (c *ColumnDef[C]) NotNull() *ColumnDef[C] {
	c.notNull = true
	c.explicitNull = false
	return c
}

// Nullable sets explicit NULL constraint.
func (c *ColumnDef[C]) Nullable() *ColumnDef[C] {
	c.notNull = false
	c.explicitNull = true
	return c
}

// Unique sets UNIQUE constraint.
func (c *ColumnDef[C]) Unique() *ColumnDef[C] {
	c.unique = true
	return c
}

// PrimaryKey sets PRIMARY KEY constraint.
func (c *ColumnDef[C]) PrimaryKey() *ColumnDef[C] {
	c.primaryKey = true
	return c
}

// Default sets DEFAULT value (raw SQL expression).
func (c *ColumnDef[C]) Default(value string) *ColumnDef[C] {
	c.defaultValue = &value
	return c
}

// References sets REFERENCES clause.
func (c *ColumnDef[C]) References(table, column string) *ColumnDef[C] {
	c.reference = &Reference{table: table, column: column}
	return c
}

// OnDelete sets ON DELETE action for REFERENCES.
func (c *ColumnDef[C]) OnDelete(action string) *ColumnDef[C] {
	if c.reference == nil {
		c.reference = &Reference{}
	}
	c.reference.onDelete = action
	return c
}

// OnUpdate sets ON UPDATE action for REFERENCES.
func (c *ColumnDef[C]) OnUpdate(action string) *ColumnDef[C] {
	if c.reference == nil {
		c.reference = &Reference{}
	}
	c.reference.onUpdate = action
	return c
}

// Check sets CHECK constraint (raw SQL expression).
func (c *ColumnDef[C]) Check(expr string) *ColumnDef[C] {
	c.checkExpr = &expr
	return c
}

func (c *ColumnDef[C]) Build() (string, []any) {
	var b strings.Builder
	paramIndex := 1
	args := make([]any, 0)
	c.AddToBuilder(&b, "", &paramIndex, &args)
	return b.String(), nil
}

func (c *ColumnDef[C]) AddToBuilder(buf *strings.Builder, _ string, _ *int, _ *[]any) {
	buf.WriteString(c.fa.String())
	if c.dataType.String() != "" {
		buf.WriteByte(' ')
		buf.WriteString(c.dataType.String())
	}
	if c.notNull {
		buf.WriteString(" NOT NULL")
	} else if c.explicitNull {
		buf.WriteString(" NULL")
	}
	if c.defaultValue != nil {
		buf.WriteString(" DEFAULT ")
		buf.WriteString(*c.defaultValue)
	}
	if c.unique {
		buf.WriteString(" UNIQUE")
	}
	if c.primaryKey {
		buf.WriteString(" PRIMARY KEY")
	}
	if c.reference != nil && c.reference.table != "" {
		buf.WriteString(" REFERENCES ")
		buf.WriteString(c.reference.table)
		if c.reference.column != "" {
			buf.WriteByte('(')
			buf.WriteString(c.reference.column)
			buf.WriteByte(')')
		}
		if c.reference.onDelete != "" {
			buf.WriteString(" ON DELETE ")
			buf.WriteString(c.reference.onDelete)
		}
		if c.reference.onUpdate != "" {
			buf.WriteString(" ON UPDATE ")
			buf.WriteString(c.reference.onUpdate)
		}
	}
	if c.checkExpr != nil {
		buf.WriteString(" CHECK (")
		buf.WriteString(*c.checkExpr)
		buf.WriteByte(')')
	}
}
