package ddl

import "strings"

// ColumnDef describes a column for CREATE/ALTER TABLE.
type ColumnDef struct {
	name         string
	dataType     string
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

// Column creates a column definition.
func Column(name, dataType string) *ColumnDef {
	return &ColumnDef{name: name, dataType: dataType}
}

// NotNull sets NOT NULL constraint.
func (c *ColumnDef) NotNull() *ColumnDef {
	c.notNull = true
	c.explicitNull = false
	return c
}

// Nullable sets explicit NULL constraint.
func (c *ColumnDef) Nullable() *ColumnDef {
	c.notNull = false
	c.explicitNull = true
	return c
}

// Unique sets UNIQUE constraint.
func (c *ColumnDef) Unique() *ColumnDef {
	c.unique = true
	return c
}

// PrimaryKey sets PRIMARY KEY constraint.
func (c *ColumnDef) PrimaryKey() *ColumnDef {
	c.primaryKey = true
	return c
}

// Default sets DEFAULT value (raw SQL expression).
func (c *ColumnDef) Default(value string) *ColumnDef {
	c.defaultValue = &value
	return c
}

// References sets REFERENCES clause.
func (c *ColumnDef) References(table, column string) *ColumnDef {
	c.reference = &Reference{table: table, column: column}
	return c
}

// OnDelete sets ON DELETE action for REFERENCES.
func (c *ColumnDef) OnDelete(action string) *ColumnDef {
	if c.reference == nil {
		c.reference = &Reference{}
	}
	c.reference.onDelete = action
	return c
}

// OnUpdate sets ON UPDATE action for REFERENCES.
func (c *ColumnDef) OnUpdate(action string) *ColumnDef {
	if c.reference == nil {
		c.reference = &Reference{}
	}
	c.reference.onUpdate = action
	return c
}

// Check sets CHECK constraint (raw SQL expression).
func (c *ColumnDef) Check(expr string) *ColumnDef {
	c.checkExpr = &expr
	return c
}

func (c *ColumnDef) build() string {
	var b strings.Builder
	b.WriteString(c.name)
	if c.dataType != "" {
		b.WriteByte(' ')
		b.WriteString(c.dataType)
	}
	if c.notNull {
		b.WriteString(" NOT NULL")
	} else if c.explicitNull {
		b.WriteString(" NULL")
	}
	if c.defaultValue != nil {
		b.WriteString(" DEFAULT ")
		b.WriteString(*c.defaultValue)
	}
	if c.unique {
		b.WriteString(" UNIQUE")
	}
	if c.primaryKey {
		b.WriteString(" PRIMARY KEY")
	}
	if c.reference != nil && c.reference.table != "" {
		b.WriteString(" REFERENCES ")
		b.WriteString(c.reference.table)
		if c.reference.column != "" {
			b.WriteByte('(')
			b.WriteString(c.reference.column)
			b.WriteByte(')')
		}
		if c.reference.onDelete != "" {
			b.WriteString(" ON DELETE ")
			b.WriteString(c.reference.onDelete)
		}
		if c.reference.onUpdate != "" {
			b.WriteString(" ON UPDATE ")
			b.WriteString(c.reference.onUpdate)
		}
	}
	if c.checkExpr != nil {
		b.WriteString(" CHECK (")
		b.WriteString(*c.checkExpr)
		b.WriteByte(')')
	}
	return b.String()
}

// TableConstraint builds table-level constraints.
type TableConstraint interface {
	build() string
}

type primaryKeyConstraint struct {
	name    string
	columns []string
}

func PrimaryKey(columns ...string) TableConstraint {
	return &primaryKeyConstraint{columns: columns}
}

func NamedPrimaryKey(name string, columns ...string) TableConstraint {
	return &primaryKeyConstraint{name: name, columns: columns}
}

func (c *primaryKeyConstraint) build() string {
	var b strings.Builder
	if c.name != "" {
		b.WriteString("CONSTRAINT ")
		b.WriteString(c.name)
		b.WriteByte(' ')
	}
	b.WriteString("PRIMARY KEY (")
	for i, col := range c.columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(col)
	}
	b.WriteByte(')')
	return b.String()
}

type uniqueConstraint struct {
	name    string
	columns []string
}

func Unique(columns ...string) TableConstraint {
	return &uniqueConstraint{columns: columns}
}

func NamedUnique(name string, columns ...string) TableConstraint {
	return &uniqueConstraint{name: name, columns: columns}
}

func (c *uniqueConstraint) build() string {
	var b strings.Builder
	if c.name != "" {
		b.WriteString("CONSTRAINT ")
		b.WriteString(c.name)
		b.WriteByte(' ')
	}
	b.WriteString("UNIQUE (")
	for i, col := range c.columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(col)
	}
	b.WriteByte(')')
	return b.String()
}

type foreignKeyConstraint struct {
	name       string
	columns    []string
	refTable   string
	refColumns []string
	onDelete   string
	onUpdate   string
}

func ForeignKey(columns []string, refTable string, refColumns []string) TableConstraint {
	return &foreignKeyConstraint{
		columns:    columns,
		refTable:   refTable,
		refColumns: refColumns,
	}
}

func NamedForeignKey(name string, columns []string, refTable string, refColumns []string) TableConstraint {
	return &foreignKeyConstraint{
		name:       name,
		columns:    columns,
		refTable:   refTable,
		refColumns: refColumns,
	}
}

func (c *foreignKeyConstraint) OnDelete(action string) *foreignKeyConstraint {
	c.onDelete = action
	return c
}

func (c *foreignKeyConstraint) OnUpdate(action string) *foreignKeyConstraint {
	c.onUpdate = action
	return c
}

func (c *foreignKeyConstraint) build() string {
	var b strings.Builder
	if c.name != "" {
		b.WriteString("CONSTRAINT ")
		b.WriteString(c.name)
		b.WriteByte(' ')
	}
	b.WriteString("FOREIGN KEY (")
	for i, col := range c.columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(col)
	}
	b.WriteString(") REFERENCES ")
	b.WriteString(c.refTable)
	if len(c.refColumns) > 0 {
		b.WriteByte('(')
		for i, col := range c.refColumns {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(col)
		}
		b.WriteByte(')')
	}
	if c.onDelete != "" {
		b.WriteString(" ON DELETE ")
		b.WriteString(c.onDelete)
	}
	if c.onUpdate != "" {
		b.WriteString(" ON UPDATE ")
		b.WriteString(c.onUpdate)
	}
	return b.String()
}

type checkConstraint struct {
	name string
	expr string
}

func Check(expr string) TableConstraint {
	return &checkConstraint{expr: expr}
}

func NamedCheck(name, expr string) TableConstraint {
	return &checkConstraint{name: name, expr: expr}
}

func (c *checkConstraint) build() string {
	var b strings.Builder
	if c.name != "" {
		b.WriteString("CONSTRAINT ")
		b.WriteString(c.name)
		b.WriteByte(' ')
	}
	b.WriteString("CHECK (")
	b.WriteString(c.expr)
	b.WriteByte(')')
	return b.String()
}

// CreateTable builds CREATE TABLE statements.
type CreateTable struct {
	name        string
	ifNotExists bool
	columns     []*ColumnDef
	constraints []TableConstraint
}

// CreateTableStmt creates a new CREATE TABLE builder.
func CreateTableStmt(name string) *CreateTable {
	return &CreateTable{name: name}
}

// IfNotExists adds IF NOT EXISTS.
func (c *CreateTable) IfNotExists() *CreateTable {
	c.ifNotExists = true
	return c
}

// Column adds a column definition.
func (c *CreateTable) Column(column *ColumnDef) *CreateTable {
	c.columns = append(c.columns, column)
	return c
}

// Columns adds multiple column definitions.
func (c *CreateTable) Columns(columns ...*ColumnDef) *CreateTable {
	c.columns = append(c.columns, columns...)
	return c
}

// Constraint adds a table constraint.
func (c *CreateTable) Constraint(constraint TableConstraint) *CreateTable {
	c.constraints = append(c.constraints, constraint)
	return c
}

// Build returns SQL and args (DDL has no args).
func (c *CreateTable) Build() (string, []any) {
	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	if c.ifNotExists {
		b.WriteString("IF NOT EXISTS ")
	}
	b.WriteString(c.name)
	b.WriteString(" (")
	first := true
	for _, col := range c.columns {
		if col == nil {
			continue
		}
		if !first {
			b.WriteString(", ")
		}
		b.WriteString(col.build())
		first = false
	}
	for _, constraint := range c.constraints {
		if constraint == nil {
			continue
		}
		if !first {
			b.WriteString(", ")
		}
		b.WriteString(constraint.build())
		first = false
	}
	b.WriteString(");")
	return b.String(), nil
}

// AlterTable builds ALTER TABLE statements.
type AlterTable struct {
	name     string
	ifExists bool
	actions  []string
}

// AlterTableStmt creates a new ALTER TABLE builder.
func AlterTableStmt(name string) *AlterTable {
	return &AlterTable{name: name}
}

// IfExists adds IF EXISTS.
func (a *AlterTable) IfExists() *AlterTable {
	a.ifExists = true
	return a
}

// AddColumn adds ADD COLUMN action.
func (a *AlterTable) AddColumn(column *ColumnDef) *AlterTable {
	if column != nil {
		a.actions = append(a.actions, "ADD COLUMN "+column.build())
	}
	return a
}

// AddColumnIfNotExists adds ADD COLUMN IF NOT EXISTS action.
func (a *AlterTable) AddColumnIfNotExists(column *ColumnDef) *AlterTable {
	if column != nil {
		a.actions = append(a.actions, "ADD COLUMN IF NOT EXISTS "+column.build())
	}
	return a
}

// DropColumn adds DROP COLUMN action.
func (a *AlterTable) DropColumn(name string, ifExists bool) *AlterTable {
	action := "DROP COLUMN "
	if ifExists {
		action += "IF EXISTS "
	}
	action += name
	a.actions = append(a.actions, action)
	return a
}

// RenameColumn adds RENAME COLUMN action.
func (a *AlterTable) RenameColumn(oldName, newName string) *AlterTable {
	a.actions = append(a.actions, "RENAME COLUMN "+oldName+" TO "+newName)
	return a
}

// RenameTo adds RENAME TO action.
func (a *AlterTable) RenameTo(newName string) *AlterTable {
	a.actions = append(a.actions, "RENAME TO "+newName)
	return a
}

// AddConstraint adds ADD constraint action.
func (a *AlterTable) AddConstraint(constraint TableConstraint) *AlterTable {
	if constraint != nil {
		a.actions = append(a.actions, "ADD "+constraint.build())
	}
	return a
}

// DropConstraint adds DROP CONSTRAINT action.
func (a *AlterTable) DropConstraint(name string, ifExists bool) *AlterTable {
	action := "DROP CONSTRAINT "
	if ifExists {
		action += "IF EXISTS "
	}
	action += name
	a.actions = append(a.actions, action)
	return a
}

// SetColumnType adds ALTER COLUMN ... TYPE action.
func (a *AlterTable) SetColumnType(name, dataType string) *AlterTable {
	a.actions = append(a.actions, "ALTER COLUMN "+name+" TYPE "+dataType)
	return a
}

// Build returns SQL and args (DDL has no args).
func (a *AlterTable) Build() (string, []any) {
	var b strings.Builder
	b.WriteString("ALTER TABLE ")
	if a.ifExists {
		b.WriteString("IF EXISTS ")
	}
	b.WriteString(a.name)
	if len(a.actions) > 0 {
		b.WriteByte(' ')
		for i, action := range a.actions {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(action)
		}
	}
	b.WriteString(";")
	return b.String(), nil
}

// DropTable builds DROP TABLE statements.
type DropTable struct {
	names    []string
	ifExists bool
	cascade  bool
	restrict bool
}

// DropTableStmt creates a new DROP TABLE builder.
func DropTableStmt(names ...string) *DropTable {
	return &DropTable{names: names}
}

// IfExists adds IF EXISTS.
func (d *DropTable) IfExists() *DropTable {
	d.ifExists = true
	return d
}

// Cascade adds CASCADE.
func (d *DropTable) Cascade() *DropTable {
	d.cascade = true
	d.restrict = false
	return d
}

// Restrict adds RESTRICT.
func (d *DropTable) Restrict() *DropTable {
	d.restrict = true
	d.cascade = false
	return d
}

// Build returns SQL and args (DDL has no args).
func (d *DropTable) Build() (string, []any) {
	var b strings.Builder
	b.WriteString("DROP TABLE ")
	if d.ifExists {
		b.WriteString("IF EXISTS ")
	}
	for i, name := range d.names {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(name)
	}
	if d.cascade {
		b.WriteString(" CASCADE")
	} else if d.restrict {
		b.WriteString(" RESTRICT")
	}
	b.WriteString(";")
	return b.String(), nil
}

// CreateIndex builds CREATE INDEX statements.
type CreateIndex struct {
	name         string
	table        string
	ifNotExists  bool
	unique       bool
	concurrently bool
	columns      []string
	using        string
	where        string
}

// CreateIndexStmt creates a new CREATE INDEX builder.
func CreateIndexStmt(name, table string) *CreateIndex {
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
func (c *CreateIndex) Columns(columns ...string) *CreateIndex {
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
	var b strings.Builder
	b.WriteString("CREATE ")
	if c.unique {
		b.WriteString("UNIQUE ")
	}
	b.WriteString("INDEX ")
	if c.concurrently {
		b.WriteString("CONCURRENTLY ")
	}
	if c.ifNotExists {
		b.WriteString("IF NOT EXISTS ")
	}
	b.WriteString(c.name)
	b.WriteString(" ON ")
	b.WriteString(c.table)
	if c.using != "" {
		b.WriteString(" USING ")
		b.WriteString(c.using)
	}
	b.WriteString(" (")
	for i, col := range c.columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(col)
	}
	b.WriteByte(')')
	if c.where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(c.where)
	}
	b.WriteString(";")
	return b.String(), nil
}

// DropIndex builds DROP INDEX statements.
type DropIndex struct {
	names        []string
	ifExists     bool
	cascade      bool
	restrict     bool
	concurrently bool
}

// DropIndexStmt creates a new DROP INDEX builder.
func DropIndexStmt(names ...string) *DropIndex {
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
	var b strings.Builder
	b.WriteString("DROP INDEX ")
	if d.concurrently {
		b.WriteString("CONCURRENTLY ")
	}
	if d.ifExists {
		b.WriteString("IF EXISTS ")
	}
	for i, name := range d.names {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(name)
	}
	if d.cascade {
		b.WriteString(" CASCADE")
	} else if d.restrict {
		b.WriteString(" RESTRICT")
	}
	b.WriteString(";")
	return b.String(), nil
}
