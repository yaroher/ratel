package ddl

import (
	"strings"

	"github.com/yaroher/ratel/pkg/types"
)

// ColumnDDL describes a column for CREATE/ALTER TABLE.
type ColumnDDL[C types.ColumnAlias] struct {
	fa           C
	dataType     Datatype
	notNull      bool
	explicitNull bool
	unique       bool
	primaryKey   bool
	defaultValue *string
	reference    *Reference
	checkExpr    *string
	// Constraint names for generating named constraints
	uniqueConstraintName *string
	pkConstraintName     *string
	fkConstraintName     *string
	checkConstraintName  *string
}

func NewColumnDDL[C types.ColumnAlias](
	fa C,
	dataType Datatype,
	opts ...ColumnOption[C],
) *ColumnDDL[C] {
	col := &ColumnDDL[C]{
		fa:       fa,
		dataType: dataType,
	}
	for _, opt := range opts {
		opt(col)
	}
	return col
}

// Reference describes a REFERENCES clause with optional actions.
type Reference struct {
	table    string
	column   string
	onDelete string
	onUpdate string
}

// ColumnOption is a functional option for configuring a ColumnDDL.
type ColumnOption[C types.ColumnAlias] func(*ColumnDDL[C])

// WithNotNull sets NOT NULL constraint.
func WithNotNull[C types.ColumnAlias]() ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.notNull = true
		c.explicitNull = false
	}
}

// WithNullable sets explicit NULL constraint.
func WithNullable[C types.ColumnAlias]() ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.notNull = false
		c.explicitNull = true
	}
}

// WithUnique sets UNIQUE constraint.
func WithUnique[C types.ColumnAlias]() ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.unique = true
	}
}

// WithPrimaryKey sets PRIMARY KEY constraint.
func WithPrimaryKey[C types.ColumnAlias]() ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.primaryKey = true
	}
}

// WithDefault sets DEFAULT value (raw SQL expression).
func WithDefault[C types.ColumnAlias](value string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.defaultValue = &value
	}
}

// WithReferences sets REFERENCES clause.
func WithReferences[C types.ColumnAlias](table, column string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.reference = &Reference{table: table, column: column}
	}
}

// WithOnDelete sets ON DELETE action for REFERENCES.
func WithOnDelete[C types.ColumnAlias](action string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		if c.reference == nil {
			c.reference = &Reference{}
		}
		c.reference.onDelete = action
	}
}

// WithOnUpdate sets ON UPDATE action for REFERENCES.
func WithOnUpdate[C types.ColumnAlias](action string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		if c.reference == nil {
			c.reference = &Reference{}
		}
		c.reference.onUpdate = action
	}
}

// WithCheck sets CHECK constraint (raw SQL expression).
func WithCheck[C types.ColumnAlias](expr string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.checkExpr = &expr
	}
}

// WithUniqueNamed sets UNIQUE constraint with explicit name.
func WithUniqueNamed[C types.ColumnAlias](constraintName string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.unique = true
		c.uniqueConstraintName = &constraintName
	}
}

// WithPrimaryKeyNamed sets PRIMARY KEY constraint with explicit name.
func WithPrimaryKeyNamed[C types.ColumnAlias](constraintName string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.primaryKey = true
		c.pkConstraintName = &constraintName
	}
}

// WithReferencesNamed sets REFERENCES clause with explicit constraint name.
func WithReferencesNamed[C types.ColumnAlias](constraintName, table, column string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.reference = &Reference{table: table, column: column}
		c.fkConstraintName = &constraintName
	}
}

// WithCheckNamed sets CHECK constraint with explicit name.
func WithCheckNamed[C types.ColumnAlias](constraintName, expr string) ColumnOption[C] {
	return func(c *ColumnDDL[C]) {
		c.checkExpr = &expr
		c.checkConstraintName = &constraintName
	}
}

func (c *ColumnDDL[C]) Alias() C {
	return c.fa
}

// NotNull sets NOT NULL constraint.
func (c *ColumnDDL[C]) NotNull() *ColumnDDL[C] {
	c.notNull = true
	c.explicitNull = false
	return c
}

// Nullable sets explicit NULL constraint.
func (c *ColumnDDL[C]) Nullable() *ColumnDDL[C] {
	c.notNull = false
	c.explicitNull = true
	return c
}

// Unique sets UNIQUE constraint.
func (c *ColumnDDL[C]) Unique() *ColumnDDL[C] {
	c.unique = true
	return c
}

// PrimaryKey sets PRIMARY KEY constraint.
func (c *ColumnDDL[C]) PrimaryKey() *ColumnDDL[C] {
	c.primaryKey = true
	return c
}

// Default sets DEFAULT value (raw SQL expression).
func (c *ColumnDDL[C]) Default(value string) *ColumnDDL[C] {
	c.defaultValue = &value
	return c
}

// References sets REFERENCES clause.
func (c *ColumnDDL[C]) References(table, column string) *ColumnDDL[C] {
	c.reference = &Reference{table: table, column: column}
	return c
}

// OnDelete sets ON DELETE action for REFERENCES.
func (c *ColumnDDL[C]) OnDelete(action string) *ColumnDDL[C] {
	if c.reference == nil {
		c.reference = &Reference{}
	}
	c.reference.onDelete = action
	return c
}

// OnUpdate sets ON UPDATE action for REFERENCES.
func (c *ColumnDDL[C]) OnUpdate(action string) *ColumnDDL[C] {
	if c.reference == nil {
		c.reference = &Reference{}
	}
	c.reference.onUpdate = action
	return c
}

// Check sets CHECK constraint (raw SQL expression).
func (c *ColumnDDL[C]) Check(expr string) *ColumnDDL[C] {
	c.checkExpr = &expr
	return c
}

// UniqueNamed sets UNIQUE constraint with explicit name.
func (c *ColumnDDL[C]) UniqueNamed(constraintName string) *ColumnDDL[C] {
	c.unique = true
	c.uniqueConstraintName = &constraintName
	return c
}

// PrimaryKeyNamed sets PRIMARY KEY constraint with explicit name.
func (c *ColumnDDL[C]) PrimaryKeyNamed(constraintName string) *ColumnDDL[C] {
	c.primaryKey = true
	c.pkConstraintName = &constraintName
	return c
}

// ReferencesNamed sets REFERENCES clause with explicit constraint name.
func (c *ColumnDDL[C]) ReferencesNamed(constraintName, table, column string) *ColumnDDL[C] {
	c.reference = &Reference{table: table, column: column}
	c.fkConstraintName = &constraintName
	return c
}

// CheckNamed sets CHECK constraint with explicit name.
func (c *ColumnDDL[C]) CheckNamed(constraintName, expr string) *ColumnDDL[C] {
	c.checkExpr = &expr
	c.checkConstraintName = &constraintName
	return c
}

func (c *ColumnDDL[C]) SchemaSql() string {
	var sql strings.Builder

	// ColumnDDL alias and data type
	sql.WriteString(c.fa.String())
	sql.WriteString(" ")
	sql.WriteString(c.dataType.String())

	// NOT NULL / NULL
	if c.notNull {
		sql.WriteString(" NOT NULL")
	} else if c.explicitNull {
		sql.WriteString(" NULL")
	}

	// UNIQUE (with optional constraint name)
	if c.unique {
		if c.uniqueConstraintName != nil {
			sql.WriteString(" CONSTRAINT ")
			sql.WriteString(*c.uniqueConstraintName)
		}
		sql.WriteString(" UNIQUE")
	}

	// PRIMARY KEY (with optional constraint name)
	if c.primaryKey {
		if c.pkConstraintName != nil {
			sql.WriteString(" CONSTRAINT ")
			sql.WriteString(*c.pkConstraintName)
		}
		sql.WriteString(" PRIMARY KEY")
	}

	// DEFAULT
	if c.defaultValue != nil {
		sql.WriteString(" DEFAULT ")
		sql.WriteString(*c.defaultValue)
	}

	// REFERENCES (with optional constraint name)
	if c.reference != nil {
		if c.fkConstraintName != nil {
			sql.WriteString(" CONSTRAINT ")
			sql.WriteString(*c.fkConstraintName)
		}
		sql.WriteString(" REFERENCES ")
		sql.WriteString(c.reference.table)
		sql.WriteString("(")
		sql.WriteString(c.reference.column)
		sql.WriteString(")")

		if c.reference.onDelete != "" {
			sql.WriteString(" ON DELETE ")
			sql.WriteString(c.reference.onDelete)
		}

		if c.reference.onUpdate != "" {
			sql.WriteString(" ON UPDATE ")
			sql.WriteString(c.reference.onUpdate)
		}
	}

	// CHECK (with optional constraint name)
	if c.checkExpr != nil {
		if c.checkConstraintName != nil {
			sql.WriteString(" CONSTRAINT ")
			sql.WriteString(*c.checkConstraintName)
		}
		sql.WriteString(" CHECK (")
		sql.WriteString(*c.checkExpr)
		sql.WriteString(")")
	}

	return sql.String()
}
