package models

import (
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/schema"
)

// OrderItemsAlias is the table alias type for the order_items table
type OrderItemsAlias string

func (o OrderItemsAlias) String() string { return string(o) }

const OrderItemsAliasName OrderItemsAlias = "order_items"

// OrderItemsColumnAlias represents column names for the order_items table
type OrderItemsColumnAlias string

func (o OrderItemsColumnAlias) String() string { return string(o) }

const (
	OrderItemsColumnOrderID   OrderItemsColumnAlias = "order_id"
	OrderItemsColumnLineNo    OrderItemsColumnAlias = "line_no"
	OrderItemsColumnProductID OrderItemsColumnAlias = "product_id"
	OrderItemsColumnQty       OrderItemsColumnAlias = "qty"
	OrderItemsColumnUnitPrice OrderItemsColumnAlias = "unit_price"
)

// Index names
const OrderItemsIndexProduct OrderItemsAlias = "ix_order_items_product"

// OrderItemsScanner is the scanner struct for order_items rows
type OrderItemsScanner struct {
	OrderID   int64
	LineNo    int32
	ProductID int64
	Qty       int32
	UnitPrice float64
}

func (o *OrderItemsScanner) GetTarget(s string) func() any {
	switch OrderItemsColumnAlias(s) {
	case OrderItemsColumnOrderID:
		return func() any { return &o.OrderID }
	case OrderItemsColumnLineNo:
		return func() any { return &o.LineNo }
	case OrderItemsColumnProductID:
		return func() any { return &o.ProductID }
	case OrderItemsColumnQty:
		return func() any { return &o.Qty }
	case OrderItemsColumnUnitPrice:
		return func() any { return &o.UnitPrice }
	default:
		panic("unknown field: " + s)
	}
}

func (o *OrderItemsScanner) GetSetter(f OrderItemsColumnAlias) func() set.ValueSetter[OrderItemsColumnAlias] {
	switch f {
	case OrderItemsColumnOrderID:
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.OrderID) }
	case OrderItemsColumnLineNo:
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.LineNo) }
	case OrderItemsColumnProductID:
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.ProductID) }
	case OrderItemsColumnQty:
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.Qty) }
	case OrderItemsColumnUnitPrice:
		return func() set.ValueSetter[OrderItemsColumnAlias] { return set.NewSetter(f, &o.UnitPrice) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (o *OrderItemsScanner) GetValue(f OrderItemsColumnAlias) func() any {
	switch f {
	case OrderItemsColumnOrderID:
		return func() any { return o.OrderID }
	case OrderItemsColumnLineNo:
		return func() any { return o.LineNo }
	case OrderItemsColumnProductID:
		return func() any { return o.ProductID }
	case OrderItemsColumnQty:
		return func() any { return o.Qty }
	case OrderItemsColumnUnitPrice:
		return func() any { return o.UnitPrice }
	default:
		panic("unknown field: " + string(f))
	}
}

// OrderItemsTable represents the order_items table with its columns
type OrderItemsTable struct {
	*schema.Table[OrderItemsAlias, OrderItemsColumnAlias, *OrderItemsScanner]
	OrderID   schema.BigIntColumnI[OrderItemsColumnAlias]
	LineNo    schema.IntegerColumnI[OrderItemsColumnAlias]
	ProductID schema.BigIntColumnI[OrderItemsColumnAlias]
	Qty       schema.IntegerColumnI[OrderItemsColumnAlias]
	UnitPrice schema.NumericColumnI[OrderItemsColumnAlias]
}

// OrderItems is the global order_items table instance
var OrderItems = func() OrderItemsTable {
	orderIDCol := schema.BigIntColumn(
		OrderItemsColumnOrderID,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithReferences[OrderItemsColumnAlias]("orders", "order_id"),
		ddl.WithOnDelete[OrderItemsColumnAlias]("CASCADE"),
	)
	lineNoCol := schema.IntegerColumn(
		OrderItemsColumnLineNo,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithCheck[OrderItemsColumnAlias]("line_no > 0"),
	)
	productIDCol := schema.BigIntColumn(
		OrderItemsColumnProductID,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithReferences[OrderItemsColumnAlias]("products", "product_id"),
	)
	qtyCol := schema.IntegerColumn(
		OrderItemsColumnQty,
		ddl.WithNotNull[OrderItemsColumnAlias](),
		ddl.WithCheck[OrderItemsColumnAlias]("qty > 0"),
	)
	unitPriceCol := schema.NumericColumn(
		OrderItemsColumnUnitPrice,
		12, 2,
		ddl.WithNotNull[OrderItemsColumnAlias](),
	)

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
				OrderItemsColumnOrderID,
				OrderItemsColumnProductID,
			}),
			ddl.WithIndexes[OrderItemsAlias, OrderItemsColumnAlias](
				ddl.NewIndex[OrderItemsAlias, OrderItemsColumnAlias](
					string(OrderItemsIndexProduct),
					OrderItemsAliasName,
				).OnColumns(OrderItemsColumnProductID),
			),
		),
		OrderID:   orderIDCol,
		LineNo:    lineNoCol,
		ProductID: productIDCol,
		Qty:       qtyCol,
		UnitPrice: unitPriceCol,
	}
}()

// OrderItemsRef is a reference to the order_items table for relations
var OrderItemsRef schema.RelationTableAlias[OrderItemsAlias] = OrderItems.Table
