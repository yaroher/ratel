package models

import (
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/schema"
)

// CategoriesAlias is the table alias type for the categories table
type CategoriesAlias string

func (c CategoriesAlias) String() string { return string(c) }

const CategoriesAliasName CategoriesAlias = "categories"

// CategoriesColumnAlias represents column names for the categories table
type CategoriesColumnAlias string

func (c CategoriesColumnAlias) String() string { return string(c) }

const (
	CategoriesColumnCategoryID CategoriesColumnAlias = "category_id"
	CategoriesColumnName       CategoriesColumnAlias = "name"
	CategoriesColumnSlug       CategoriesColumnAlias = "slug"
	CategoriesColumnParentID   CategoriesColumnAlias = "parent_id"
)

// CategoriesScanner is the scanner struct for categories rows
type CategoriesScanner struct {
	CategoryID int64
	Name       string
	Slug       string
	ParentID   *int64
	Products   []*ProductsScanner
}

func (c *CategoriesScanner) GetTarget(s string) func() any {
	switch CategoriesColumnAlias(s) {
	case CategoriesColumnCategoryID:
		return func() any { return &c.CategoryID }
	case CategoriesColumnName:
		return func() any { return &c.Name }
	case CategoriesColumnSlug:
		return func() any { return &c.Slug }
	case CategoriesColumnParentID:
		return func() any { return &c.ParentID }
	default:
		panic("unknown field: " + s)
	}
}

func (c *CategoriesScanner) GetSetter(f CategoriesColumnAlias) func() set.ValueSetter[CategoriesColumnAlias] {
	switch f {
	case CategoriesColumnCategoryID:
		return func() set.ValueSetter[CategoriesColumnAlias] { return set.NewSetter(f, &c.CategoryID) }
	case CategoriesColumnName:
		return func() set.ValueSetter[CategoriesColumnAlias] { return set.NewSetter(f, &c.Name) }
	case CategoriesColumnSlug:
		return func() set.ValueSetter[CategoriesColumnAlias] { return set.NewSetter(f, &c.Slug) }
	case CategoriesColumnParentID:
		return func() set.ValueSetter[CategoriesColumnAlias] { return set.NewSetter(f, &c.ParentID) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (c *CategoriesScanner) GetValue(f CategoriesColumnAlias) func() any {
	switch f {
	case CategoriesColumnCategoryID:
		return func() any { return c.CategoryID }
	case CategoriesColumnName:
		return func() any { return c.Name }
	case CategoriesColumnSlug:
		return func() any { return c.Slug }
	case CategoriesColumnParentID:
		return func() any { return c.ParentID }
	default:
		panic("unknown field: " + string(f))
	}
}

// CategoriesTable represents the categories table with its columns
type CategoriesTable struct {
	*schema.Table[CategoriesAlias, CategoriesColumnAlias, *CategoriesScanner]
	CategoryID schema.BigSerialColumnI[CategoriesColumnAlias]
	Name       schema.TextColumnI[CategoriesColumnAlias]
	Slug       schema.TextColumnI[CategoriesColumnAlias]
	ParentID   schema.BigIntColumnI[CategoriesColumnAlias]
}

// Categories is the global categories table instance
var Categories = func() CategoriesTable {
	categoryIDCol := schema.BigSerialColumn(CategoriesColumnCategoryID, ddl.WithPrimaryKey[CategoriesColumnAlias]())
	nameCol := schema.TextColumn(CategoriesColumnName, ddl.WithNotNull[CategoriesColumnAlias]())
	slugCol := schema.TextColumn(
		CategoriesColumnSlug,
		ddl.WithNotNull[CategoriesColumnAlias](),
		ddl.WithUnique[CategoriesColumnAlias](),
	)
	parentIDCol := schema.BigIntColumn(
		CategoriesColumnParentID,
		ddl.WithReferences[CategoriesColumnAlias]("categories", "category_id"),
	)

	return CategoriesTable{
		Table: schema.NewTable[CategoriesAlias, CategoriesColumnAlias, *CategoriesScanner](
			CategoriesAliasName,
			func() *CategoriesScanner { return &CategoriesScanner{} },
			[]*ddl.ColumnDDL[CategoriesColumnAlias]{
				categoryIDCol.DDL(),
				nameCol.DDL(),
				slugCol.DDL(),
				parentIDCol.DDL(),
			},
		),
		CategoryID: categoryIDCol,
		Name:       nameCol,
		Slug:       slugCol,
		ParentID:   parentIDCol,
	}
}()

// CategoriesRef is a reference to the categories table for relations
var CategoriesRef schema.RelationTableAlias[CategoriesAlias] = Categories.Table
