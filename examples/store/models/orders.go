package models

import (
	"time"

	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/schema"
)

// OrdersAlias is the table alias type for the orders table
type OrdersAlias string

func (o OrdersAlias) String() string { return string(o) }

const OrdersAliasName OrdersAlias = "orders"

// OrdersColumnAlias represents column names for the orders table
type OrdersColumnAlias string

func (o OrdersColumnAlias) String() string { return string(o) }

const (
	OrdersColumnOrderID       OrdersColumnAlias = "order_id"
	OrdersColumnUserID        OrdersColumnAlias = "user_id"
	OrdersColumnStatus        OrdersColumnAlias = "status"
	OrdersColumnCurrency      OrdersColumnAlias = "currency"
	OrdersColumnCreatedAt     OrdersColumnAlias = "created_at"
	OrdersColumnUpdatedAt     OrdersColumnAlias = "updated_at"
	OrdersColumnCreatedAtDesc OrdersColumnAlias = "created_at DESC"
)

// Index names
const OrdersIndexUserCreated OrdersAlias = "ix_orders_user_created"

// OrdersScanner is the scanner struct for orders rows
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
	case OrdersColumnOrderID:
		return func() any { return &o.OrderID }
	case OrdersColumnUserID:
		return func() any { return &o.UserID }
	case OrdersColumnStatus:
		return func() any { return &o.Status }
	case OrdersColumnCurrency:
		return func() any { return &o.Currency }
	case OrdersColumnCreatedAt:
		return func() any { return &o.CreatedAt }
	case OrdersColumnUpdatedAt:
		return func() any { return &o.UpdatedAt }
	default:
		panic("unknown field: " + s)
	}
}

func (o *OrdersScanner) GetSetter(f OrdersColumnAlias) func() set.ValueSetter[OrdersColumnAlias] {
	switch f {
	case OrdersColumnOrderID:
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.OrderID) }
	case OrdersColumnUserID:
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.UserID) }
	case OrdersColumnStatus:
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.Status) }
	case OrdersColumnCurrency:
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.Currency) }
	case OrdersColumnCreatedAt:
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.CreatedAt) }
	case OrdersColumnUpdatedAt:
		return func() set.ValueSetter[OrdersColumnAlias] { return set.NewSetter(f, &o.UpdatedAt) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (o *OrdersScanner) GetValue(f OrdersColumnAlias) func() any {
	switch f {
	case OrdersColumnOrderID:
		return func() any { return o.OrderID }
	case OrdersColumnUserID:
		return func() any { return o.UserID }
	case OrdersColumnStatus:
		return func() any { return o.Status }
	case OrdersColumnCurrency:
		return func() any { return o.Currency }
	case OrdersColumnCreatedAt:
		return func() any { return o.CreatedAt }
	case OrdersColumnUpdatedAt:
		return func() any { return o.UpdatedAt }
	default:
		panic("unknown field: " + string(f))
	}
}

// Relations returns the relation loaders for the orders table
func (o *OrdersScanner) Relations() []exec.RelationLoader[*OrdersScanner] {
	return []exec.RelationLoader[*OrdersScanner]{
		schema.BelongsToLoad(
			OrdersUser,
			Users.Table,
			OrdersColumnUserID,
			func(order *OrdersScanner, user *UsersScanner) { order.User = user },
		),
		schema.BelongsToLoad(
			OrdersCurrency,
			Currency.Table,
			OrdersColumnCurrency,
			func(order *OrdersScanner, money *CurrencyScanner) { order.Money = money },
		),
	}
}

// OrdersTable represents the orders table with its columns
type OrdersTable struct {
	*schema.Table[OrdersAlias, OrdersColumnAlias, *OrdersScanner]
	OrderID   schema.BigSerialColumnI[OrdersColumnAlias]
	UserID    schema.BigIntColumnI[OrdersColumnAlias]
	Status    schema.TextColumnI[OrdersColumnAlias]
	Currency  schema.CharColumnI[OrdersColumnAlias]
	CreatedAt schema.TimestamptzColumnI[OrdersColumnAlias]
	UpdatedAt schema.TimestamptzColumnI[OrdersColumnAlias]
}

// Orders is the global orders table instance
var Orders = func() OrdersTable {
	orderIDCol := schema.BigSerialColumn(OrdersColumnOrderID, ddl.WithPrimaryKey[OrdersColumnAlias]())
	userIDCol := schema.BigIntColumn(
		OrdersColumnUserID,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithReferences[OrdersColumnAlias]("users", "user_id"),
	)
	statusCol := schema.TextColumn(
		OrdersColumnStatus,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithCheck[OrdersColumnAlias]("status IN ('NEW', 'PAID', 'CANCELLED', 'SHIPPED')"),
	)
	currencyCol := schema.CharColumn(
		OrdersColumnCurrency,
		3,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithReferences[OrdersColumnAlias]("currency", "code"),
	)
	createdAtCol := schema.TimestamptzColumn(
		OrdersColumnCreatedAt,
		ddl.WithNotNull[OrdersColumnAlias](),
		ddl.WithDefault[OrdersColumnAlias]("now()"),
	)
	updatedAtCol := schema.TimestamptzColumn(
		OrdersColumnUpdatedAt,
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
				).OnColumns(OrdersColumnUserID, OrdersColumnCreatedAtDesc),
			),
			ddl.WithPostStatements[OrdersAlias, OrdersColumnAlias](
				"ALTER TABLE {table} ENABLE ROW LEVEL SECURITY",
				"CREATE POLICY orders_own_data ON {table} FOR ALL USING (user_id = current_setting('app.current_user_id')::bigint)",
			),
		),
		OrderID:   orderIDCol,
		UserID:    userIDCol,
		Status:    statusCol,
		Currency:  currencyCol,
		CreatedAt: createdAtCol,
		UpdatedAt: updatedAtCol,
	}
}()

// OrdersRef is a reference to the orders table for relations
var OrdersRef schema.RelationTableAlias[OrdersAlias] = Orders.Table
