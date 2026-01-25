package models

import (
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/schema"
)

// ProductCategoriesAlias is the table alias type for the product_categories pivot table
type ProductCategoriesAlias string

func (p ProductCategoriesAlias) String() string { return string(p) }

const ProductCategoriesAliasName ProductCategoriesAlias = "product_categories"

// ProductCategoriesColumnAlias represents column names for the product_categories table
type ProductCategoriesColumnAlias string

func (p ProductCategoriesColumnAlias) String() string { return string(p) }

const (
	ProductCategoriesColumnProductID  ProductCategoriesColumnAlias = "product_id"
	ProductCategoriesColumnCategoryID ProductCategoriesColumnAlias = "category_id"
)

// ProductCategoriesScanner is the scanner struct for product_categories rows
type ProductCategoriesScanner struct {
	ProductID  int64
	CategoryID int64
}

func (p *ProductCategoriesScanner) GetTarget(s string) func() any {
	switch ProductCategoriesColumnAlias(s) {
	case ProductCategoriesColumnProductID:
		return func() any { return &p.ProductID }
	case ProductCategoriesColumnCategoryID:
		return func() any { return &p.CategoryID }
	default:
		panic("unknown field: " + s)
	}
}

func (p *ProductCategoriesScanner) GetSetter(f ProductCategoriesColumnAlias) func() set.ValueSetter[ProductCategoriesColumnAlias] {
	switch f {
	case ProductCategoriesColumnProductID:
		return func() set.ValueSetter[ProductCategoriesColumnAlias] { return set.NewSetter(f, &p.ProductID) }
	case ProductCategoriesColumnCategoryID:
		return func() set.ValueSetter[ProductCategoriesColumnAlias] { return set.NewSetter(f, &p.CategoryID) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (p *ProductCategoriesScanner) GetValue(f ProductCategoriesColumnAlias) func() any {
	switch f {
	case ProductCategoriesColumnProductID:
		return func() any { return p.ProductID }
	case ProductCategoriesColumnCategoryID:
		return func() any { return p.CategoryID }
	default:
		panic("unknown field: " + string(f))
	}
}

// ProductCategoriesTable represents the product_categories pivot table
type ProductCategoriesTable struct {
	*schema.Table[ProductCategoriesAlias, ProductCategoriesColumnAlias, *ProductCategoriesScanner]
	ProductID  schema.BigIntColumnI[ProductCategoriesColumnAlias]
	CategoryID schema.BigIntColumnI[ProductCategoriesColumnAlias]
}

// ProductCategories is the global product_categories table instance
var ProductCategories = func() ProductCategoriesTable {
	productIDCol := schema.BigIntColumn(
		ProductCategoriesColumnProductID,
		ddl.WithNotNull[ProductCategoriesColumnAlias](),
		ddl.WithReferences[ProductCategoriesColumnAlias]("products", "product_id"),
		ddl.WithOnDelete[ProductCategoriesColumnAlias]("CASCADE"),
	)
	categoryIDCol := schema.BigIntColumn(
		ProductCategoriesColumnCategoryID,
		ddl.WithNotNull[ProductCategoriesColumnAlias](),
		ddl.WithReferences[ProductCategoriesColumnAlias]("categories", "category_id"),
		ddl.WithOnDelete[ProductCategoriesColumnAlias]("CASCADE"),
	)

	return ProductCategoriesTable{
		Table: schema.NewTable[ProductCategoriesAlias, ProductCategoriesColumnAlias, *ProductCategoriesScanner](
			ProductCategoriesAliasName,
			func() *ProductCategoriesScanner { return &ProductCategoriesScanner{} },
			[]*ddl.ColumnDDL[ProductCategoriesColumnAlias]{
				productIDCol.DDL(),
				categoryIDCol.DDL(),
			},
			ddl.WithPrimaryKeyColumns[ProductCategoriesAlias, ProductCategoriesColumnAlias]([]ProductCategoriesColumnAlias{
				ProductCategoriesColumnProductID,
				ProductCategoriesColumnCategoryID,
			}),
		),
		ProductID:  productIDCol,
		CategoryID: categoryIDCol,
	}
}()

// ProductCategoriesRef is a reference to the product_categories table for relations
var ProductCategoriesRef schema.RelationTableAlias[ProductCategoriesAlias] = ProductCategories.Table
