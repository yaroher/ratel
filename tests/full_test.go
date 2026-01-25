package tests

import (
	"os"
	"testing"
	"time"

	"github.com/yaroher/ratel/ddl"
	"github.com/yaroher/ratel/dml"
	"github.com/yaroher/ratel/dml/set"
	"github.com/yaroher/ratel/exec"
	"github.com/yaroher/ratel/schema"
)

type CurrencyAlias string

func (c CurrencyAlias) String() string { return string(c) }

const CurrencyAliasName = "currency"

type CurrencyColumnAlias string

func (c CurrencyColumnAlias) String() string { return string(c) }

const (
	CurrencyColumnAliasCode CurrencyColumnAlias = "code"
	CurrencyColumnAliasName CurrencyColumnAlias = "name"
)

type CurrencyScanner struct {
	Code string
	Name string
}

func (c *CurrencyScanner) GetTarget(s string) func() any {
	switch CurrencyColumnAlias(s) {
	case "code":
		return func() any { return &c.Code }
	case "name":
		return func() any { return &c.Name }
	default:
		panic("unknown field: " + s)
	}
}

func (c *CurrencyScanner) GetSetter(f CurrencyColumnAlias) func() set.ValueSetter[CurrencyColumnAlias] {
	switch f {
	case "code":
		return func() set.ValueSetter[CurrencyColumnAlias] { return set.NewSetter(f, &c.Code) }
	case "name":
		return func() set.ValueSetter[CurrencyColumnAlias] { return set.NewSetter(f, &c.Name) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (c *CurrencyScanner) GetValue(f CurrencyColumnAlias) func() any {
	switch f {
	case "code":
		return func() any { return c.Code }
	case "name":
		return func() any { return c.Name }
	default:
		panic("unknown field: " + string(f))
	}
}

type CurrencyTable struct {
	*schema.Table[CurrencyAlias, CurrencyColumnAlias, *CurrencyScanner]
	Code schema.CharColumnI[CurrencyColumnAlias]
	Name schema.TextColumnI[CurrencyColumnAlias]
}

var currency = func() CurrencyTable {
	codeCol := schema.CharColumn(CurrencyColumnAliasCode, 3, ddl.WithPrimaryKey[CurrencyColumnAlias]())
	nameCol := schema.TextColumn(CurrencyColumnAliasName, ddl.WithNotNull[CurrencyColumnAlias]())
	return CurrencyTable{
		Table: schema.NewTable[CurrencyAlias, CurrencyColumnAlias, *CurrencyScanner](
			CurrencyAliasName,
			func() *CurrencyScanner { return &CurrencyScanner{} },
			[]*ddl.ColumnDDL[CurrencyColumnAlias]{
				codeCol.DDL(),
				nameCol.DDL(),
			},
		),
		Code: codeCol,
		Name: nameCol,
	}
}()

type UsersAlias string

func (u UsersAlias) String() string { return string(u) }

const UsersAliasName = "users"

type UsersColumnAlias string

func (u UsersColumnAlias) String() string { return string(u) }

const (
	UsersColumnAliasUserID    UsersColumnAlias = "user_id"
	UsersColumnAliasEmail     UsersColumnAlias = "email"
	UsersColumnAliasFullName  UsersColumnAlias = "full_name"
	UsersColumnAliasIsActive  UsersColumnAlias = "is_active"
	UsersColumnAliasCreatedAt UsersColumnAlias = "created_at"
	UsersColumnAliasUpdatedAt UsersColumnAlias = "updated_at"
)

type UsersScanner struct {
	UserID    int64
	Email     string
	FullName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Orders    []*OrdersScanner
}

func (u *UsersScanner) GetTarget(s string) func() any {
	switch UsersColumnAlias(s) {
	case "user_id":
		return func() any { return &u.UserID }
	case "email":
		return func() any { return &u.Email }
	case "full_name":
		return func() any { return &u.FullName }
	case "is_active":
		return func() any { return &u.IsActive }
	case "created_at":
		return func() any { return &u.CreatedAt }
	case "updated_at":
		return func() any { return &u.UpdatedAt }
	default:
		panic("unknown field: " + s)
	}
}

func (u *UsersScanner) GetSetter(f UsersColumnAlias) func() set.ValueSetter[UsersColumnAlias] {
	switch f {
	case "user_id":
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.UserID) }
	case "email":
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.Email) }
	case "full_name":
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.FullName) }
	case "is_active":
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.IsActive) }
	case "created_at":
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.CreatedAt) }
	case "updated_at":
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.UpdatedAt) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (u *UsersScanner) GetValue(f UsersColumnAlias) func() any {
	switch f {
	case "user_id":
		return func() any { return u.UserID }
	case "email":
		return func() any { return u.Email }
	case "full_name":
		return func() any { return u.FullName }
	case "is_active":
		return func() any { return u.IsActive }
	case "created_at":
		return func() any { return u.CreatedAt }
	case "updated_at":
		return func() any { return u.UpdatedAt }
	default:
		panic("unknown field: " + string(f))
	}
}

type UsersTable struct {
	*schema.Table[UsersAlias, UsersColumnAlias, *UsersScanner]
	UserID    schema.BigSerialColumnI[UsersColumnAlias]
	Email     schema.TextColumnI[UsersColumnAlias]
	FullName  schema.TextColumnI[UsersColumnAlias]
	IsActive  schema.BooleanColumnI[UsersColumnAlias]
	CreatedAt schema.TimestamptzColumnI[UsersColumnAlias]
	UpdatedAt schema.TimestamptzColumnI[UsersColumnAlias]
}

var users = func() UsersTable {
	userIDCol := schema.BigSerialColumn(UsersColumnAliasUserID, ddl.WithPrimaryKey[UsersColumnAlias]())
	emailCol := schema.TextColumn(
		UsersColumnAliasEmail,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithUnique[UsersColumnAlias](),
	)
	fullNameCol := schema.TextColumn(UsersColumnAliasFullName, ddl.WithNotNull[UsersColumnAlias]())
	isActiveCol := schema.BooleanColumn(
		UsersColumnAliasIsActive,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithDefault[UsersColumnAlias]("true"),
	)
	createdAtCol := schema.TimestamptzColumn(
		UsersColumnAliasCreatedAt,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithDefault[UsersColumnAlias]("now()"),
	)
	updatedAtCol := schema.TimestamptzColumn(
		UsersColumnAliasUpdatedAt,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithDefault[UsersColumnAlias]("now()"),
	)
	return UsersTable{
		Table: schema.NewTable[UsersAlias, UsersColumnAlias, *UsersScanner](
			UsersAliasName,
			func() *UsersScanner { return &UsersScanner{} },
			[]*ddl.ColumnDDL[UsersColumnAlias]{
				userIDCol.DDL(),
				emailCol.DDL(),
				fullNameCol.DDL(),
				isActiveCol.DDL(),
				createdAtCol.DDL(),
				updatedAtCol.DDL(),
			},
		),
		UserID:    userIDCol,
		Email:     emailCol,
		FullName:  fullNameCol,
		IsActive:  isActiveCol,
		CreatedAt: createdAtCol,
		UpdatedAt: updatedAtCol,
	}
}()

type ProductsAlias string

func (p ProductsAlias) String() string { return string(p) }

const ProductsAliasName = "products"

type ProductsColumnAlias string

func (p ProductsColumnAlias) String() string { return string(p) }

const (
	ProductsColumnAliasProductID ProductsColumnAlias = "product_id"
	ProductsColumnAliasSKU       ProductsColumnAlias = "sku"
	ProductsColumnAliasName      ProductsColumnAlias = "name"
	ProductsColumnAliasPrice     ProductsColumnAlias = "price"
	ProductsColumnAliasCurrency  ProductsColumnAlias = "currency"
	ProductsColumnAliasStockQty  ProductsColumnAlias = "stock_qty"
	ProductsColumnAliasIsDeleted ProductsColumnAlias = "is_deleted"
	ProductsColumnAliasCreatedAt ProductsColumnAlias = "created_at"
	ProductsColumnAliasUpdatedAt ProductsColumnAlias = "updated_at"
)

type ProductsScanner struct {
	ProductID int64
	SKU       string
	Name      string
	Price     float64
	Currency  string
	StockQty  int32
	IsDeleted bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *ProductsScanner) GetTarget(s string) func() any {
	switch ProductsColumnAlias(s) {
	case "product_id":
		return func() any { return &p.ProductID }
	case "sku":
		return func() any { return &p.SKU }
	case "name":
		return func() any { return &p.Name }
	case "price":
		return func() any { return &p.Price }
	case "currency":
		return func() any { return &p.Currency }
	case "stock_qty":
		return func() any { return &p.StockQty }
	case "is_deleted":
		return func() any { return &p.IsDeleted }
	case "created_at":
		return func() any { return &p.CreatedAt }
	case "updated_at":
		return func() any { return &p.UpdatedAt }
	default:
		panic("unknown field: " + s)
	}
}

func (p *ProductsScanner) GetSetter(f ProductsColumnAlias) func() set.ValueSetter[ProductsColumnAlias] {
	switch f {
	case "product_id":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.ProductID) }
	case "sku":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.SKU) }
	case "name":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.Name) }
	case "price":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.Price) }
	case "currency":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.Currency) }
	case "stock_qty":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.StockQty) }
	case "is_deleted":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.IsDeleted) }
	case "created_at":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.CreatedAt) }
	case "updated_at":
		return func() set.ValueSetter[ProductsColumnAlias] { return set.NewSetter(f, &p.UpdatedAt) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (p *ProductsScanner) GetValue(f ProductsColumnAlias) func() any {
	switch f {
	case "product_id":
		return func() any { return p.ProductID }
	case "sku":
		return func() any { return p.SKU }
	case "name":
		return func() any { return p.Name }
	case "price":
		return func() any { return p.Price }
	case "currency":
		return func() any { return p.Currency }
	case "stock_qty":
		return func() any { return p.StockQty }
	case "is_deleted":
		return func() any { return p.IsDeleted }
	case "created_at":
		return func() any { return p.CreatedAt }
	case "updated_at":
		return func() any { return p.UpdatedAt }
	default:
		panic("unknown field: " + string(f))
	}
}

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

var products ProductsTable = func() ProductsTable {
	productIDCol := schema.BigSerialColumn(ProductsColumnAliasProductID, ddl.WithPrimaryKey[ProductsColumnAlias]())
	skuCol := schema.TextColumn(
		ProductsColumnAliasSKU,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithUnique[ProductsColumnAlias](),
	)
	nameCol := schema.TextColumn(ProductsColumnAliasName, ddl.WithNotNull[ProductsColumnAlias]())
	priceCol := schema.NumericColumn(ProductsColumnAliasPrice, 12, 2, ddl.WithNotNull[ProductsColumnAlias]())
	currencyCol := schema.CharColumn(
		ProductsColumnAliasCurrency,
		3,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithReferences[ProductsColumnAlias]("currency", "code"),
	)
	stockQtyCol := schema.IntegerColumn(
		ProductsColumnAliasStockQty,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithDefault[ProductsColumnAlias]("0"),
		ddl.WithCheck[ProductsColumnAlias]("stock_qty >= 0"),
	)
	isDeletedCol := schema.BooleanColumn(
		ProductsColumnAliasIsDeleted,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithDefault[ProductsColumnAlias]("false"),
	)
	createdAtCol := schema.TimestamptzColumn(
		ProductsColumnAliasCreatedAt,
		ddl.WithNotNull[ProductsColumnAlias](),
		ddl.WithDefault[ProductsColumnAlias]("now()"),
	)
	updatedAtCol := schema.TimestamptzColumn(
		ProductsColumnAliasUpdatedAt,
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
			ddl.WithIndexes[ProductsAlias, ProductsColumnAlias](ddl.NewIndex[ProductsAlias, ProductsColumnAlias](
				string(ProductsIndexNotDeleted),
				ProductsAliasName,
			).OnColumns(ProductsColumnAliasProductID).Where("is_deleted = false")),
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

type OrdersAlias string

func (o OrdersAlias) String() string { return string(o) }

const OrdersAliasName = "orders"

type OrdersColumnAlias string

func (o OrdersColumnAlias) String() string { return string(o) }

const (
	OrdersColumnAliasOrderID       OrdersColumnAlias = "order_id"
	OrdersColumnAliasUserID        OrdersColumnAlias = "user_id"
	OrdersColumnAliasStatus        OrdersColumnAlias = "status"
	OrdersColumnAliasCurrency      OrdersColumnAlias = "currency"
	OrdersColumnAliasCreatedAt     OrdersColumnAlias = "created_at"
	OrdersColumnAliasUpdatedAt     OrdersColumnAlias = "updated_at"
	OrdersColumnAliasCreatedAtDesc OrdersColumnAlias = "created_at DESC"
)

type OrdersScanner struct {
	OrderID   int64
	UserID    int64
	Status    string
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *UsersScanner
	Money     *CurrencyScanner
}

func (o *OrdersScanner) GetTarget(s string) func() any {
	switch OrdersColumnAlias(s) {
	case "order_id":
		return func() any { return &o.OrderID }
	case "user_id":
		return func() any { return &o.UserID }
	case "status":
		return func() any { return &o.Status }
	case "currency":
		return func() any { return &o.Currency }
	case "created_at":
		return func() any { return &o.CreatedAt }
	case "updated_at":
		return func() any { return &o.UpdatedAt }
	default:
		panic("unknown field: " + s)
	}
}

func (o *OrdersScanner) GetSetter(f OrdersColumnAlias) func() set.ValueSetter[OrdersColumnAlias] {
	switch f {
	case "order_id":
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.OrderID) }
	case "user_id":
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.UserID) }
	case "status":
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.Status) }
	case "currency":
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.Currency) }
	case "created_at":
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.CreatedAt) }
	case "updated_at":
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.UpdatedAt) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (o *OrdersScanner) GetValue(f OrdersColumnAlias) func() any {
	switch f {
	case "order_id":
		return func() any { return o.OrderID }
	case "user_id":
		return func() any { return o.UserID }
	case "status":
		return func() any { return o.Status }
	case "currency":
		return func() any { return o.Currency }
	case "created_at":
		return func() any { return o.CreatedAt }
	case "updated_at":
		return func() any { return o.UpdatedAt }
	default:
		panic("unknown field: " + string(f))
	}
}

type OrdersTable struct {
	*schema.Table[OrdersAlias, OrdersColumnAlias, *OrdersScanner]
	OrderID   schema.BigSerialColumnI[OrdersColumnAlias]
	UserID    schema.BigIntColumnI[OrdersColumnAlias]
	Status    schema.TextColumnI[OrdersColumnAlias]
	Currency  schema.CharColumnI[OrdersColumnAlias]
	CreatedAt schema.TimestamptzColumnI[OrdersColumnAlias]
	UpdatedAt schema.TimestamptzColumnI[OrdersColumnAlias]
}

var orders = func() OrdersTable {
	orderIDCol := schema.BigSerialColumn(OrdersColumnAliasOrderID, ddl.WithPrimaryKey[OrdersColumnAlias]())
	userIDCol := schema.BigIntColumn(
		OrdersColumnAliasUserID,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithReferences[OrdersColumnAlias]("users", "user_id"),
	)
	statusCol := schema.TextColumn(
		OrdersColumnAliasStatus,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithCheck[OrdersColumnAlias]("status IN ('NEW', 'PAID', 'CANCELLED', 'SHIPPED')"),
	)
	currencyCol := schema.CharColumn(
		OrdersColumnAliasCurrency,
		3,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithReferences[OrdersColumnAlias]("currency", "code"),
	)
	createdAtCol := schema.TimestamptzColumn(
		OrdersColumnAliasCreatedAt,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithDefault[OrdersColumnAlias]("now()"),
	)
	updatedAtCol := schema.TimestamptzColumn(
		OrdersColumnAliasUpdatedAt,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithDefault[OrdersColumnAlias]("now()"),
	)
	return OrdersTable{
		Table: schema.NewTable[OrdersAlias, OrdersColumnAlias, *OrdersScanner](
			OrdersAliasName,
			func() *OrdersScanner { return &OrdersScanner{} },
			[]*ddl.ColumnDDL[OrdersColumnAlias]{
				orderIDCol.DDL(),
				userIDCol.DDL(),
				statusCol.DDL(),
				currencyCol.DDL(),
				createdAtCol.DDL(),
				updatedAtCol.DDL(),
			},
			ddl.WithIndexes[OrdersAlias, OrdersColumnAlias](
				ddl.NewIndex[OrdersAlias, OrdersColumnAlias](
					string(OrdersIndexUserCreated),
					OrdersAliasName,
				).OnColumns(OrdersColumnAliasUserID, OrdersColumnAliasCreatedAtDesc)),
		),
		OrderID:   orderIDCol,
		UserID:    userIDCol,
		Status:    statusCol,
		Currency:  currencyCol,
		CreatedAt: createdAtCol,
		UpdatedAt: updatedAtCol,
	}
}()

type OrderItemsAlias string

func (o OrderItemsAlias) String() string { return string(o) }

const OrderItemsAliasName = "order_items"

type OrderItemsColumnAlias string

func (o OrderItemsColumnAlias) String() string { return string(o) }

const (
	OrderItemsColumnAliasOrderID   OrderItemsColumnAlias = "order_id"
	OrderItemsColumnAliasLineNo    OrderItemsColumnAlias = "line_no"
	OrderItemsColumnAliasProductID OrderItemsColumnAlias = "product_id"
	OrderItemsColumnAliasQty       OrderItemsColumnAlias = "qty"
	OrderItemsColumnAliasUnitPrice OrderItemsColumnAlias = "unit_price"
)

type OrderItemsScanner struct {
	OrderID   int64
	LineNo    int32
	ProductID int64
	Qty       int32
	UnitPrice float64
}

func (o *OrderItemsScanner) GetTarget(s string) func() any {
	switch OrderItemsColumnAlias(s) {
	case "order_id":
		return func() any { return &o.OrderID }
	case "line_no":
		return func() any { return &o.LineNo }
	case "product_id":
		return func() any { return &o.ProductID }
	case "qty":
		return func() any { return &o.Qty }
	case "unit_price":
		return func() any { return &o.UnitPrice }
	default:
		panic("unknown field: " + s)
	}
}

func (o *OrderItemsScanner) GetSetter(f OrderItemsColumnAlias) func() set.ValueSetter[OrderItemsColumnAlias] {
	switch f {
	case "order_id":
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.OrderID) }
	case "line_no":
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.LineNo) }
	case "product_id":
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.ProductID) }
	case "qty":
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.Qty) }
	case "unit_price":
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.UnitPrice) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (o *OrderItemsScanner) GetValue(f OrderItemsColumnAlias) func() any {
	switch f {
	case "order_id":
		return func() any { return o.OrderID }
	case "line_no":
		return func() any { return o.LineNo }
	case "product_id":
		return func() any { return o.ProductID }
	case "qty":
		return func() any { return o.Qty }
	case "unit_price":
		return func() any { return o.UnitPrice }
	default:
		panic("unknown field: " + string(f))
	}
}

type OrderItemsTable struct {
	*schema.Table[OrderItemsAlias, OrderItemsColumnAlias, *OrderItemsScanner]
	OrderID   schema.BigIntColumnI[OrderItemsColumnAlias]
	LineNo    schema.IntegerColumnI[OrderItemsColumnAlias]
	ProductID schema.BigIntColumnI[OrderItemsColumnAlias]
	Qty       schema.IntegerColumnI[OrderItemsColumnAlias]
	UnitPrice schema.NumericColumnI[OrderItemsColumnAlias]
}

var orderItems OrderItemsTable = func() OrderItemsTable {
	orderIDCol := schema.BigIntColumn(
		OrderItemsColumnAliasOrderID,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithReferences[OrderItemsColumnAlias]("orders", "order_id"),
		ddl.WithOnDelete[OrderItemsColumnAlias]("CASCADE"),
	)
	lineNoCol := schema.IntegerColumn(
		OrderItemsColumnAliasLineNo,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithCheck[OrderItemsColumnAlias]("line_no > 0"),
	)
	productIDCol := schema.BigIntColumn(
		OrderItemsColumnAliasProductID,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithReferences[OrderItemsColumnAlias]("products", "product_id"),
	)
	qtyCol := schema.IntegerColumn(
		OrderItemsColumnAliasQty,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithCheck[OrderItemsColumnAlias]("qty > 0"),
	)
	unitPriceCol := schema.NumericColumn(OrderItemsColumnAliasUnitPrice, 12, 2, ddl.WithNotNull[OrderItemsColumnAlias]())
	return OrderItemsTable{
		Table: schema.NewTable[OrderItemsAlias, OrderItemsColumnAlias, *OrderItemsScanner](
			OrderItemsAliasName,
			func() *OrderItemsScanner { return &OrderItemsScanner{} },
			[]*ddl.ColumnDDL[OrderItemsColumnAlias]{
				orderIDCol.DDL(),
				lineNoCol.DDL(),
				productIDCol.DDL(),
				qtyCol.DDL(),
				unitPriceCol.DDL(),
			},
			ddl.WithUniqueColumns[OrderItemsAlias, OrderItemsColumnAlias]([]OrderItemsColumnAlias{
				OrderItemsColumnAliasOrderID,
				OrderItemsColumnAliasProductID,
			}),
			ddl.WithIndexes[OrderItemsAlias, OrderItemsColumnAlias](ddl.NewIndex[OrderItemsAlias, OrderItemsColumnAlias](
				string(OrderItemsIndexProduct),
				OrderItemsAliasName,
			).OnColumns(OrderItemsColumnAliasProductID)),
		),
		OrderID:   orderIDCol,
		LineNo:    lineNoCol,
		ProductID: productIDCol,
		Qty:       qtyCol,
		UnitPrice: unitPriceCol,
	}
}()

var _ schema.RelationTableAlias[CurrencyAlias] = currency.Table
var _ schema.RelationTableAlias[UsersAlias] = users.Table
var _ schema.RelationTableAlias[OrdersAlias] = orders.Table
var _ schema.RelationTableJoin[OrdersAlias, OrdersColumnAlias] = orders.Table
var _ schema.RelationTableQuery[OrdersAlias, OrdersColumnAlias, *OrdersScanner] = orders.Table

var currencyRef schema.RelationTableAlias[CurrencyAlias] = currency.Table
var usersRef schema.RelationTableAlias[UsersAlias] = users.Table
var ordersRef schema.RelationTableAlias[OrdersAlias] = orders.Table

var usersOrders = schema.HasMany[UsersAlias, UsersColumnAlias, *UsersScanner, OrdersAlias, OrdersColumnAlias, *OrdersScanner](
	UsersAliasName,
	ordersRef,
	OrdersColumnAliasUserID,
	UsersColumnAliasUserID,
)

var ordersUser = schema.BelongsTo[OrdersAlias, OrdersColumnAlias, *OrdersScanner, UsersAlias, UsersColumnAlias, *UsersScanner](
	OrdersAliasName,
	usersRef,
	OrdersColumnAliasUserID,
	UsersColumnAliasUserID,
)

var ordersCurrency = schema.BelongsTo[OrdersAlias, OrdersColumnAlias, *OrdersScanner, CurrencyAlias, CurrencyColumnAlias, *CurrencyScanner](
	OrdersAliasName,
	currencyRef,
	OrdersColumnAliasCurrency,
	CurrencyColumnAliasCode,
)

func (u *UsersScanner) Relations() []exec.RelationLoader[*UsersScanner] {
	return []exec.RelationLoader[*UsersScanner]{
		schema.HasManyLoad(
			usersOrders,
			orders.Table,
			UsersColumnAliasUserID,
			func(user *UsersScanner, rows []*OrdersScanner) {
				for _, row := range rows {
					row.User = user
				}
				user.Orders = rows
			},
		),
	}
}

func (o *OrdersScanner) Relations() []exec.RelationLoader[*OrdersScanner] {
	return []exec.RelationLoader[*OrdersScanner]{
		schema.BelongsToLoad(
			ordersUser,
			users.Table,
			OrdersColumnAliasUserID,
			func(order *OrdersScanner, user *UsersScanner) { order.User = user },
		),
		schema.BelongsToLoad(
			ordersCurrency,
			currency.Table,
			OrdersColumnAliasCurrency,
			func(order *OrdersScanner, money *CurrencyScanner) { order.Money = money },
		),
	}
}

const OrdersIndexUserCreated OrdersAlias = "ix_orders_user_created"
const OrderItemsIndexProduct OrderItemsAlias = "ix_order_items_product"
const ProductsIndexNotDeleted ProductsAlias = "ix_products_not_deleted"

func TestDumpSchemaSQL(t *testing.T) {
	sqlString := ddl.SchemaSQL(currency, users, products, orders, orderItems)
	err := os.WriteFile("generated.sql", []byte(sqlString), 0644)
	if err != nil {
		panic(err)
	}
}

func TestRelationsUsage(t *testing.T) {
	usersQuery := usersOrders.WithJoin(users.Table, users.SelectAll(), dml.LeftJoinType)
	usersSQL, _ := usersQuery.Build()
	t.Log(usersSQL)

	ordersQuery := ordersUser.WithJoin(orders.Table, orders.SelectAll(), dml.LeftJoinType)
	ordersQuery = ordersCurrency.WithJoin(orders.Table, ordersQuery, dml.LeftJoinType)
	ordersSQL, _ := ordersQuery.Build()
	t.Log(ordersSQL)
}
