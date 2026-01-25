package tests

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yaroher/ratel/ddl"
)

func TestRelationsLoadWithTestcontainers(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ratel"),
		postgres.WithUsername("ratel"),
		postgres.WithPassword("ratel"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Skipf("testcontainers postgres unavailable: %v", err)
		return
	}
	defer func() {
		_ = pgContainer.Terminate(ctx)
	}()

	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		t.Fatalf("pgxpool config: %v", err)
	}
	config.MaxConns = 2
	config.MinConns = 1

	var pool *pgxpool.Pool
	for attempt := 0; attempt < 10; attempt++ {
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			err = pool.Ping(ctx)
		}
		if err == nil {
			break
		}
		if pool != nil {
			pool.Close()
		}
		time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("pgxpool ping: %v", err)
	}
	defer pool.Close()

	statements := ddl.SchemaStatements(currency, users, products, orders, orderItems)

	for _, stmt := range statements {
		var execErr error
		for attempt := 0; attempt < 5; attempt++ {
			_, execErr = pool.Exec(ctx, stmt)
			if execErr == nil {
				break
			}
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
		}
		if execErr != nil {
			t.Fatalf("exec schema: %v", execErr)
		}
	}

	if _, err := currency.Execute(ctx, pool, currency.Insert().From(
		currency.Code.Set("USD"),
		currency.Name.Set("US Dollar"),
	)); err != nil {
		t.Fatalf("insert currency: %v", err)
	}

	userInsert := users.Insert().From(
		users.Email.Set("john@example.com"),
		users.FullName.Set("John Doe"),
	).Returning(UsersColumnAliasUserID)
	userSQL, userArgs := userInsert.Build()
	var userID int64
	if err := pool.QueryRow(ctx, userSQL, userArgs...).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	productInsert := products.Insert().From(
		products.SKU.Set("SKU-1"),
		products.Name.Set("Widget"),
		products.Price.Set(19.99),
		products.Currency.Set("USD"),
		products.StockQty.Set(10),
	).Returning(ProductsColumnAliasProductID)
	productSQL, productArgs := productInsert.Build()
	var productID int64
	if err := pool.QueryRow(ctx, productSQL, productArgs...).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}

	orderInsert1 := orders.Insert().From(
		orders.UserID.Set(userID),
		orders.Status.Set("NEW"),
		orders.Currency.Set("USD"),
	).Returning(OrdersColumnAliasOrderID)
	orderSQL1, orderArgs1 := orderInsert1.Build()
	var orderID1 int64
	if err := pool.QueryRow(ctx, orderSQL1, orderArgs1...).Scan(&orderID1); err != nil {
		t.Fatalf("insert order: %v", err)
	}

	orderInsert2 := orders.Insert().From(
		orders.UserID.Set(userID),
		orders.Status.Set("PAID"),
		orders.Currency.Set("USD"),
	).Returning(OrdersColumnAliasOrderID)
	orderSQL2, orderArgs2 := orderInsert2.Build()
	var orderID2 int64
	if err := pool.QueryRow(ctx, orderSQL2, orderArgs2...).Scan(&orderID2); err != nil {
		t.Fatalf("insert order 2: %v", err)
	}

	if _, err := orderItems.Execute(ctx, pool, orderItems.Insert().From(
		orderItems.OrderID.Set(orderID1),
		orderItems.LineNo.Set(1),
		orderItems.ProductID.Set(productID),
		orderItems.Qty.Set(2),
		orderItems.UnitPrice.Set(19.99),
	)); err != nil {
		t.Fatalf("insert order_items: %v", err)
	}

	if _, err := orderItems.Execute(ctx, pool, orderItems.Insert().From(
		orderItems.OrderID.Set(orderID2),
		orderItems.LineNo.Set(1),
		orderItems.ProductID.Set(productID),
		orderItems.Qty.Set(1),
		orderItems.UnitPrice.Set(19.99),
	)); err != nil {
		t.Fatalf("insert order_items 2: %v", err)
	}

	userOrders, err := usersOrders.LoadMany(ctx, pool, orders.Table, userID)
	if err != nil {
		t.Fatalf("load orders: %v", err)
	}
	if len(userOrders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(userOrders))
	}

	owner, err := ordersUser.LoadOne(ctx, pool, users.Table, userOrders[0].UserID)
	if err != nil {
		t.Fatalf("load order user: %v", err)
	}
	if owner.UserID != userID {
		t.Fatalf("expected owner %d, got %d", userID, owner.UserID)
	}

	currencyRow, err := ordersCurrency.LoadOne(ctx, pool, currency.Table, userOrders[0].Currency)
	if err != nil {
		t.Fatalf("load order currency: %v", err)
	}
	if currencyRow.Code != "USD" {
		t.Fatalf("expected currency USD, got %s", currencyRow.Code)
	}
}
