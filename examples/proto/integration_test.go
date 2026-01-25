package storepb

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yaroher/ratel/pkg/ddl"
)

// setupTestDB creates a PostgreSQL testcontainer and returns a connection pool
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute),
		),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	cleanup := func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

// createSchema creates all tables in the database
func createSchema(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	statements := ddl.SchemaStatements(
		Currencys,
		Users,
		Categorys,
		Tags,
		Products,
		Orders,
		OrderItems,
	)

	for _, stmt := range statements {
		t.Logf("Executing: %s", stmt)
		_, err := db.Exec(ctx, stmt)
		if err != nil {
			t.Fatalf("failed to execute schema statement: %v\nSQL: %s", err, stmt)
		}
	}
	t.Log("Schema created successfully")
}

// TestGeneratedModels tests the generated models from proto
func TestGeneratedModels(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	createSchema(t, ctx, db)

	// ========================================================================
	// INSERT: Currency
	// ========================================================================
	t.Run("Insert Currency", func(t *testing.T) {
		currencies := []struct {
			code string
			name string
		}{
			{"USD", "US Dollar"},
			{"EUR", "Euro"},
			{"RUB", "Russian Ruble"},
		}

		for _, c := range currencies {
			query := Currencys.Insert().
				Columns(CurrencyColumnCode, CurrencyColumnName).
				Values(c.code, c.name)

			_, err := Currencys.Execute(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert currency %s: %v", c.code, err)
			}
		}
		t.Log("Currencies inserted")
	})

	// ========================================================================
	// INSERT: Users
	// ========================================================================
	var userIDs []int64
	t.Run("Insert Users", func(t *testing.T) {
		users := []struct {
			email    string
			fullName string
		}{
			{"john@example.com", "John Doe"},
			{"jane@example.com", "Jane Smith"},
			{"bob@example.com", "Bob Wilson"},
		}

		for _, u := range users {
			query := Users.Insert().
				Columns(
					UserColumnEmail,
					UserColumnFullName,
				).
				Values(u.email, u.fullName).
				Returning(UserColumnUserId)

			user, err := Users.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert user %s: %v", u.email, err)
			}
			userIDs = append(userIDs, user.UserId)
			t.Logf("Inserted user: %s (ID: %d)", u.fullName, user.UserId)
		}
	})

	// ========================================================================
	// INSERT: Categories
	// ========================================================================
	var categoryIDs []int64
	t.Run("Insert Categories", func(t *testing.T) {
		categories := []struct {
			name string
			slug string
		}{
			{"Electronics", "electronics"},
			{"Clothing", "clothing"},
		}

		for _, c := range categories {
			query := Categorys.Insert().
				Columns(
					CategoryColumnName,
					CategoryColumnSlug,
				).
				Values(c.name, c.slug).
				Returning(CategoryColumnCategoryId)

			cat, err := Categorys.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert category %s: %v", c.name, err)
			}
			categoryIDs = append(categoryIDs, cat.CategoryId)
			t.Logf("Inserted category: %s (ID: %d)", c.name, cat.CategoryId)
		}
	})

	// ========================================================================
	// INSERT: Tags
	// ========================================================================
	var tagIDs []int64
	t.Run("Insert Tags", func(t *testing.T) {
		tags := []struct {
			name string
			slug string
		}{
			{"New", "new"},
			{"Sale", "sale"},
			{"Popular", "popular"},
		}

		for _, tg := range tags {
			query := Tags.Insert().
				Columns(TagColumnName, TagColumnSlug).
				Values(tg.name, tg.slug).
				Returning(TagColumnTagId)

			tag, err := Tags.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert tag %s: %v", tg.name, err)
			}
			tagIDs = append(tagIDs, tag.TagId)
			t.Logf("Inserted tag: %s (ID: %d)", tg.name, tag.TagId)
		}
	})

	// ========================================================================
	// INSERT: Products
	// ========================================================================
	var productIDs []int64
	t.Run("Insert Products", func(t *testing.T) {
		products := []struct {
			sku      string
			name     string
			price    float64
			currency string
			stockQty int32
		}{
			{"PHONE-001", "iPhone 15 Pro", 999.99, "USD", 50},
			{"LAPTOP-001", "MacBook Pro 16", 2499.99, "USD", 25},
			{"SHIRT-001", "Cotton T-Shirt", 29.99, "USD", 200},
		}

		for _, p := range products {
			query := Products.Insert().
				Columns(
					ProductColumnSku,
					ProductColumnName,
					ProductColumnPrice,
					ProductColumnCurrency,
					ProductColumnStockQty,
				).
				Values(p.sku, p.name, p.price, p.currency, p.stockQty).
				Returning(ProductColumnProductId)

			prod, err := Products.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert product %s: %v", p.name, err)
			}
			productIDs = append(productIDs, prod.ProductId)
			t.Logf("Inserted product: %s (ID: %d)", p.name, prod.ProductId)
		}
	})

	// ========================================================================
	// INSERT: Orders
	// ========================================================================
	var orderIDs []int64
	t.Run("Insert Orders", func(t *testing.T) {
		orders := []struct {
			userID   int64
			status   string
			currency string
		}{
			{userIDs[0], "NEW", "USD"},
			{userIDs[0], "PAID", "USD"},
			{userIDs[1], "SHIPPED", "USD"},
		}

		for _, o := range orders {
			query := Orders.Insert().
				Columns(
					OrderColumnUserId,
					OrderColumnStatus,
					OrderColumnCurrency,
				).
				Values(o.userID, o.status, o.currency).
				Returning(OrderColumnOrderId)

			order, err := Orders.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert order for user %d: %v", o.userID, err)
			}
			orderIDs = append(orderIDs, order.OrderId)
			t.Logf("Inserted order: ID=%d, User=%d, Status=%s", order.OrderId, o.userID, o.status)
		}
	})

	// ========================================================================
	// INSERT: Order Items
	// ========================================================================
	t.Run("Insert Order Items", func(t *testing.T) {
		items := []struct {
			orderID   int64
			lineNo    int32
			productID int64
			qty       int32
			unitPrice float64
		}{
			{orderIDs[0], 1, productIDs[0], 1, 999.99},
			{orderIDs[0], 2, productIDs[2], 2, 29.99},
			{orderIDs[1], 1, productIDs[1], 1, 2499.99},
		}

		for _, item := range items {
			query := OrderItems.Insert().
				Columns(
					OrderItemColumnOrderId,
					OrderItemColumnLineNo,
					OrderItemColumnProductId,
					OrderItemColumnQty,
					OrderItemColumnUnitPrice,
				).
				Values(item.orderID, item.lineNo, item.productID, item.qty, item.unitPrice)

			_, err := OrderItems.Execute(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert order item: %v", err)
			}
		}
		t.Log("Order items inserted")
	})

	// ========================================================================
	// SELECT: Basic queries
	// ========================================================================
	t.Run("Select All Users", func(t *testing.T) {
		query := Users.SelectAll()
		users, err := Users.Query(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to select users: %v", err)
		}

		if len(users) != 3 {
			t.Errorf("expected 3 users, got %d", len(users))
		}

		for _, u := range users {
			t.Logf("User: ID=%d, Email=%s, Name=%s, Active=%v",
				u.UserId, u.Email, u.FullName, u.IsActive)
		}
	})

	t.Run("Select Products with WHERE", func(t *testing.T) {
		// Select products with price > 500
		query := Products.SelectAll().
			Where(Products.Price.Gt(500.0))

		products, err := Products.Query(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to select products: %v", err)
		}

		if len(products) != 2 {
			t.Errorf("expected 2 expensive products, got %d", len(products))
		}

		for _, p := range products {
			t.Logf("Expensive Product: ID=%d, Name=%s, Price=%.2f",
				p.ProductId, p.Name, p.Price)
		}
	})

	// ========================================================================
	// UPDATE
	// ========================================================================
	t.Run("Update Product Stock", func(t *testing.T) {
		query := Products.Update().
			Set(Products.StockQty.Set(100)).
			Where(Products.ProductId.Eq(productIDs[0])).
			Returning(ProductColumnProductId, ProductColumnStockQty)

		product, err := Products.QueryRow(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to update product: %v", err)
		}

		if product.StockQty != 100 {
			t.Errorf("expected stock_qty=100, got %d", product.StockQty)
		}
		t.Logf("Updated product %d stock to %d", product.ProductId, product.StockQty)
	})

	t.Run("Update Order Status", func(t *testing.T) {
		query := Orders.Update().
			Set(Orders.Status.Set("PAID")).
			Where(Orders.OrderId.Eq(orderIDs[0])).
			Returning(OrderColumnOrderId, OrderColumnStatus)

		order, err := Orders.QueryRow(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to update order: %v", err)
		}

		if order.Status != "PAID" {
			t.Errorf("expected status=PAID, got %s", order.Status)
		}
		t.Logf("Updated order %d status to %s", order.OrderId, order.Status)
	})

	// ========================================================================
	// DELETE
	// ========================================================================
	t.Run("Delete Order Item", func(t *testing.T) {
		query := OrderItems.Delete().
			Where(
				OrderItems.Table.And(
					OrderItems.OrderId.Eq(orderIDs[0]),
					OrderItems.LineNo.Eq(int32(2)),
				),
			)

		affected, err := OrderItems.Execute(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to delete order item: %v", err)
		}

		if affected != 1 {
			t.Errorf("expected 1 row affected, got %d", affected)
		}
		t.Logf("Deleted %d order item(s)", affected)
	})

	t.Run("Soft Delete Product", func(t *testing.T) {
		// Soft delete by setting is_deleted = true
		query := Products.Update().
			Set(Products.IsDeleted.Set(true)).
			Where(Products.ProductId.Eq(productIDs[2]))

		affected, err := Products.Execute(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to soft delete product: %v", err)
		}

		if affected != 1 {
			t.Errorf("expected 1 row affected, got %d", affected)
		}
		t.Logf("Soft deleted product %d", productIDs[2])

		// Verify with select
		selectQuery := Products.SelectAll().
			Where(
				Products.Table.And(
					Products.ProductId.Eq(productIDs[2]),
					Products.IsDeleted.Eq(true),
				),
			)

		product, err := Products.QueryRow(ctx, db, selectQuery)
		if err != nil {
			t.Fatalf("failed to verify soft deleted product: %v", err)
		}

		if !product.IsDeleted {
			t.Error("product should be marked as deleted")
		}
		t.Logf("Verified product %d is soft deleted", product.ProductId)
	})

	// ========================================================================
	// Complex Query: Products not deleted
	// ========================================================================
	t.Run("Select Active Products Count", func(t *testing.T) {
		query := Products.SelectAll().
			Where(Products.IsDeleted.Eq(false))

		products, err := Products.Query(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to select active products: %v", err)
		}

		// Should be 2 (one was soft deleted)
		if len(products) != 2 {
			t.Errorf("expected 2 active products, got %d", len(products))
		}
		t.Logf("Found %d active products", len(products))
	})

	// Suppress unused variables
	_ = categoryIDs
	_ = tagIDs

	t.Log("All generated model tests passed!")
}
