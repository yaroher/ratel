package models

import (
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/schema"
)

// ProductTagsAlias is the table alias type for the product_tags pivot table
type ProductTagsAlias string

func (p ProductTagsAlias) String() string { return string(p) }

const ProductTagsAliasName ProductTagsAlias = "product_tags"

// ProductTagsColumnAlias represents column names for the product_tags table
type ProductTagsColumnAlias string

func (p ProductTagsColumnAlias) String() string { return string(p) }

const (
	ProductTagsColumnProductID ProductTagsColumnAlias = "product_id"
	ProductTagsColumnTagID     ProductTagsColumnAlias = "tag_id"
)

// ProductTagsScanner is the scanner struct for product_tags rows
type ProductTagsScanner struct {
	ProductID int64
	TagID     int64
}

func (p *ProductTagsScanner) GetTarget(s string) func() any {
	switch ProductTagsColumnAlias(s) {
	case ProductTagsColumnProductID:
		return func() any { return &p.ProductID }
	case ProductTagsColumnTagID:
		return func() any { return &p.TagID }
	default:
		panic("unknown field: " + s)
	}
}

func (p *ProductTagsScanner) GetSetter(f ProductTagsColumnAlias) func() set.ValueSetter[ProductTagsColumnAlias] {
	switch f {
	case ProductTagsColumnProductID:
		return func() set.ValueSetter[ProductTagsColumnAlias] { return set.NewSetter(f, &p.ProductID) }
	case ProductTagsColumnTagID:
		return func() set.ValueSetter[ProductTagsColumnAlias] { return set.NewSetter(f, &p.TagID) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (p *ProductTagsScanner) GetValue(f ProductTagsColumnAlias) func() any {
	switch f {
	case ProductTagsColumnProductID:
		return func() any { return p.ProductID }
	case ProductTagsColumnTagID:
		return func() any { return p.TagID }
	default:
		panic("unknown field: " + string(f))
	}
}

// ProductTagsTable represents the product_tags pivot table
type ProductTagsTable struct {
	*schema.Table[ProductTagsAlias, ProductTagsColumnAlias, *ProductTagsScanner]
	ProductID schema.BigIntColumnI[ProductTagsColumnAlias]
	TagID     schema.BigIntColumnI[ProductTagsColumnAlias]
}

// ProductTags is the global product_tags table instance
var ProductTags = func() ProductTagsTable {
	productIDCol := schema.BigIntColumn(
		ProductTagsColumnProductID,
		ddl.WithNotNull[ProductTagsColumnAlias](),
		ddl.WithReferences[ProductTagsColumnAlias]("products", "product_id"),
		ddl.WithOnDelete[ProductTagsColumnAlias]("CASCADE"),
	)
	tagIDCol := schema.BigIntColumn(
		ProductTagsColumnTagID,
		ddl.WithNotNull[ProductTagsColumnAlias](),
		ddl.WithReferences[ProductTagsColumnAlias]("tags", "tag_id"),
		ddl.WithOnDelete[ProductTagsColumnAlias]("CASCADE"),
	)

	return ProductTagsTable{
		Table: schema.NewTable[ProductTagsAlias, ProductTagsColumnAlias, *ProductTagsScanner](
			ProductTagsAliasName,
			func() *ProductTagsScanner { return &ProductTagsScanner{} },
			[]*ddl.ColumnDDL[ProductTagsColumnAlias]{
				productIDCol.DDL(),
				tagIDCol.DDL(),
			},
			ddl.WithPrimaryKeyColumns[ProductTagsAlias, ProductTagsColumnAlias]([]ProductTagsColumnAlias{
				ProductTagsColumnProductID,
				ProductTagsColumnTagID,
			}),
		),
		ProductID: productIDCol,
		TagID:     tagIDCol,
	}
}()

// ProductTagsRef is a reference to the product_tags table for relations
var ProductTagsRef schema.RelationTableAlias[ProductTagsAlias] = ProductTags.Table
