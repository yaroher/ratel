package migrate

// Change represents a schema change.
type Change interface {
	change() // marker method
}

// Schema-level changes
type AddSchema struct{ S *Schema }
type DropSchema struct{ S *Schema }

// Table-level changes
type AddTable struct{ T *Table }
type DropTable struct{ T *Table }
type ModifyTable struct {
	From    *Table
	To      *Table
	Changes []Change
}

// Column-level changes
type AddColumn struct{ C *Column }
type DropColumn struct{ C *Column }
type ModifyColumn struct {
	From *Column
	To   *Column
}

// Index changes
type AddIndex struct{ I *Index }
type DropIndex struct{ I *Index }
type ModifyIndex struct {
	From *Index
	To   *Index
}

// Foreign key changes
type AddForeignKey struct{ FK *ForeignKey }
type DropForeignKey struct{ FK *ForeignKey }

// Check constraint changes
type AddCheck struct{ C *Check }
type DropCheck struct{ C *Check }

// RLS changes
type EnableRLS struct{ Table string }
type DisableRLS struct{ Table string }
type ForceRLS struct{ Table string }
type UnforceRLS struct{ Table string }
type AddPolicy struct {
	Table string
	P     *Policy
}
type DropPolicy struct {
	Table string
	P     *Policy
}
type ModifyPolicy struct {
	Table string
	From  *Policy
	To    *Policy
}

// Extension changes
type AddExtension struct{ E *Extension }
type DropExtension struct{ E *Extension }

// Trigger changes
type AddTrigger struct {
	Table string
	T     *Trigger
}
type DropTrigger struct {
	Table string
	T     *Trigger
}
type ModifyTrigger struct {
	Table string
	From  *Trigger
	To    *Trigger
}

// Function changes
type AddFunction struct{ F *Function }
type DropFunction struct{ F *Function }
type ModifyFunction struct {
	From *Function
	To   *Function
}

// Marker method implementations
func (AddSchema) change()      {}
func (DropSchema) change()     {}
func (AddTable) change()       {}
func (DropTable) change()      {}
func (ModifyTable) change()    {}
func (AddColumn) change()      {}
func (DropColumn) change()     {}
func (ModifyColumn) change()   {}
func (AddIndex) change()       {}
func (DropIndex) change()      {}
func (ModifyIndex) change()    {}
func (AddForeignKey) change()  {}
func (DropForeignKey) change() {}
func (AddCheck) change()       {}
func (DropCheck) change()      {}
func (EnableRLS) change()      {}
func (DisableRLS) change()     {}
func (ForceRLS) change()       {}
func (UnforceRLS) change()     {}
func (AddPolicy) change()      {}
func (DropPolicy) change()     {}
func (ModifyPolicy) change()   {}
func (AddExtension) change()   {}
func (DropExtension) change()  {}
func (AddTrigger) change()     {}
func (DropTrigger) change()    {}
func (ModifyTrigger) change()  {}
func (AddFunction) change()    {}
func (DropFunction) change()   {}
func (ModifyFunction) change() {}
