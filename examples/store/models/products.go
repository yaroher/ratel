package models

import (
	"time"

	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/schema"
)

// ProductsAlias is the table alias type for the products table
type ProductsAlias string

func (p ProductsAlias) String() string { return string(p) }

const ProductsAliasName ProductsAlias = "products"

// ProductsColumnAlias represents column names for the products table
type ProductsColumnAlias string

func (p ProductsColumnAlias) String() string { return string(p) }

const (
	ProductsColumnProductID ProductsColumnAlias = "product_id"
	ProductsColumnSKU       ProductsColumnAlias = "sku"
	ProductsColumnName      ProductsColumnAlias = "name"
	ProductsColumnPrice     ProductsColumnAlias = "price"
	ProductsColumnCurrency  ProductsColumnAlias = "currency"
	ProductsColumnStockQty  ProductsColumnAlias = "stock_qty"
	ProductsColumnIsDeleted ProductsColumnAlias = "is_deleted"
	ProductsColumnCreatedAt ProductsColumnAlias = "created_at"
	ProductsColumnUpdatedAt ProductsColumnAlias = "updated_at"
)

// Index names
const ProductsIndexNotDeleted ProductsAlias = "ix_products_not_deleted"

// ProductsScanner is the scanner struct for products rows
type ProductsScanner struct {
	ProductID  int64
	SKU        string
	Name       string
	Price      float64
	Currency   string
	StockQty   int32
	IsDeleted  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Categories []*CategoriesScanner // M2M relation
	Tags       []*TagsScanner       // M2M relation
}

func (p *ProductsScanner) GetTarget(s string) func() any {
	switch ProductsColumnAlias(s) {
	case ProductsColumnProductID:
		return func() any { return &p.ProductID }
	case ProductsColumnSKU:
		return func() any { return &p.SKU }
	case ProductsColumnName:
		return func() any { return &p.Name }
	case ProductsColumnPrice:
		return func() any { return &p.Price }
	case ProductsColumnCurrency:
		return func() any { return &p.Currency }
	case ProductsColumnStockQty:
		return func() any { return &p.StockQty }
	case ProductsColumnIsDeleted:
		return func() any { return &p.IsDeleted }
	case ProductsColumnCreatedAt:
		return func() any { return &p.CreatedAt }
	case ProductsColumnUpdatedAt:
		return func() any { return &p.UpdatedAt }
	default:
		panic("unknown field: " + s)
	}
}

func (p *ProductsScanner) GetSetter(f ProductsColumnAlias) func() set.ValueSetter[ProductsColumnAlias] {
	switch f {
	case ProductsColumnProductID:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.ProductID) }
	case ProductsColumnSKU:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.SKU) }
	case ProductsColumnName:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.Name) }
	case ProductsColumnPrice:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.Price) }
	case ProductsColumnCurrency:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.Currency) }
	case ProductsColumnStockQty:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.StockQty) }
	case ProductsColumnIsDeleted:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.IsDeleted) }
	case ProductsColumnCreatedAt:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.CreatedAt) }
	case ProductsColumnUpdatedAt:
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.UpdatedAt) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (p *ProductsScanner) GetValue(f ProductsColumnAlias) func() any {
	switch f {
	case ProductsColumnProductID:
		return func() any { return p.ProductID }
	case ProductsColumnSKU:
		return func() any { return p.SKU }
	case ProductsColumnName:
		return func() any { return p.Name }
	case ProductsColumnPrice:
		return func() any { return p.Price }
	case ProductsColumnCurrency:
		return func() any { return p.Currency }
	case ProductsColumnStockQty:
		return func() any { return p.StockQty }
	case ProductsColumnIsDeleted:
		return func() any { return p.IsDeleted }
	case ProductsColumnCreatedAt:
		return func() any { return p.CreatedAt }
	case ProductsColumnUpdatedAt:
		return func() any { return p.UpdatedAt }
	default:
		panic("unknown field: " + string(f))
	}
}

// Relations returns the relation loaders for the products table
func (p *ProductsScanner) Relations() []exec.RelationLoader[*ProductsScanner] {
	return []exec.RelationLoader[*ProductsScanner]{
		schema.ManyToManyLoad(
			ProductsCategories,
			Categories.Table,
			ProductsColumnProductID,
			func(product *ProductsScanner, categories []*CategoriesScanner) {
				product.Categories = categories
			},
		),
		schema.ManyToManyLoad(
			ProductsTags,
			Tags.Table,
			ProductsColumnProductID,
			func(product *ProductsScanner, tags []*TagsScanner) {
				product.Tags = tags
			},
		),
	}
}

// ProductsTable represents the products table with its columns
type ProductsTable struct {
	*schema.Table[ProductsAlias, ProductsColumnAlias, *ProductsScanner]
	ProductID schema.BigSerialColumnI[ProductsColumnAlias]
	SKU       schema.TextColumnI[ProductsColumnAlias]
	Name      schema.TextColumnI[ProductsColumnAlias]
	Price     schema.NumericColumnI[ProductsColumnAlias]
	Currency  schema.CharColumnI[ProductsColumnAlias]
	StockQty  schema.IntegerColumnI[ProductsColumnAlias]
	IsDeleted schema.BooleanColumnI[ProductsColumnAlias]
	CreatedAt schema.TimestamptzColumnI[ProductsColumnAlias]
	UpdatedAt schema.TimestamptzColumnI[ProductsColumnAlias]
}

// Products is the global products table instance
var Products = func() ProductsTable {
	productIDCol := schema.BigSerialColumn(ProductsColumnProductID, ddl.WithPrimaryKey[ProductsColumnAlias]())
	skuCol := schema.TextColumn(
		ProductsColumnSKU,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithUnique[ProductsColumnAlias](),
	)
	nameCol := schema.TextColumn(ProductsColumnName, ddl.WithNotNull[ProductsColumnAlias]())
	priceCol := schema.NumericColumn(ProductsColumnPrice, 12, 2, ddl.WithNotNull[ProductsColumnAlias]())
	currencyCol := schema.CharColumn(
		ProductsColumnCurrency,
		3,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithReferences[ProductsColumnAlias]("currency", "code"),
	)
	stockQtyCol := schema.IntegerColumn(
		ProductsColumnStockQty,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithDefault[ProductsColumnAlias]("0"),
		ddl.WithCheck[ProductsColumnAlias]("stock_qty >= 0"),
	)
	isDeletedCol := schema.BooleanColumn(
		ProductsColumnIsDeleted,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithDefault[ProductsColumnAlias]("false"),
	)
	createdAtCol := schema.TimestamptzColumn(
		ProductsColumnCreatedAt,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithDefault[ProductsColumnAlias]("now()"),
	)
	updatedAtCol := schema.TimestamptzColumn(
		ProductsColumnUpdatedAt,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithDefault[ProductsColumnAlias]("now()"),
	)

	return ProductsTable{
		Table: schema.NewTable[ProductsAlias, ProductsColumnAlias, *ProductsScanner](
			ProductsAliasName,
			func() *ProductsScanner { return &ProductsScanner{} },
			[]*ddl.ColumnDDL[ProductsColumnAlias]{
				productIDCol.DDL(),
				skuCol.DDL(),
				nameCol.DDL(),
				priceCol.DDL(),
				currencyCol.DDL(),
				stockQtyCol.DDL(),
				isDeletedCol.DDL(),
				createdAtCol.DDL(),
				updatedAtCol.DDL(),
			},
			ddl.WithIndexes[ProductsAlias, ProductsColumnAlias](
				ddl.NewIndex[ProductsAlias, ProductsColumnAlias](
					string(ProductsIndexNotDeleted),
					ProductsAliasName,
				).OnColumns(ProductsColumnProductID).Where("is_deleted = false"),
			),
		),
		ProductID: productIDCol,
		SKU:       skuCol,
		Name:      nameCol,
		Price:     priceCol,
		Currency:  currencyCol,
		StockQty:  stockQtyCol,
		IsDeleted: isDeletedCol,
		CreatedAt: createdAtCol,
		UpdatedAt: updatedAtCol,
	}
}()

// ProductsRef is a reference to the products table for relations
var ProductsRef schema.RelationTableAlias[ProductsAlias] = Products.Table
