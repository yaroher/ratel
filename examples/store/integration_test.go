package main

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yaroher/ratel/examples/store/models"
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml"
	"github.com/yaroher/ratel/pkg/exec"
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
		models.Currency,
		models.Users,
		models.Categories,
		models.Tags,
		models.Products,
		models.Orders,
		models.OrderItems,
		models.ProductCategories,
		models.ProductTags,
	)

	for _, stmt := range statements {
		_, err := db.Exec(ctx, stmt)
		if err != nil {
			t.Fatalf("failed to execute schema statement: %v\nSQL: %s", err, stmt)
		}
	}
	t.Log("Schema created successfully")
}

// TestFullCycle tests the complete CRUD cycle with relations
func TestFullCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
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
			query := models.Currency.Insert().
				Columns(models.CurrencyColumnCode, models.CurrencyColumnName).
				Values(c.code, c.name)

			_, err := models.Currency.Execute(ctx, db, query)
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
			query := models.Users.Insert().
				Columns(
					models.UsersColumnEmail,
					models.UsersColumnFullName,
				).
				Values(u.email, u.fullName).
				Returning(models.UsersColumnUserID)

			user, err := models.Users.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert user %s: %v", u.email, err)
			}
			userIDs = append(userIDs, user.UserID)
			t.Logf("Inserted user: %s (ID: %d)", u.fullName, user.UserID)
		}
	})

	// ========================================================================
	// INSERT: Categories (with hierarchy)
	// ========================================================================
	var categoryIDs []int64
	t.Run("Insert Categories", func(t *testing.T) {
		// Root categories
		rootCategories := []struct {
			name string
			slug string
		}{
			{"Electronics", "electronics"},
			{"Clothing", "clothing"},
		}

		for _, c := range rootCategories {
			query := models.Categories.Insert().
				Columns(
					models.CategoriesColumnName,
					models.CategoriesColumnSlug,
				).
				Values(c.name, c.slug).
				Returning(models.CategoriesColumnCategoryID)

			cat, err := models.Categories.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert category %s: %v", c.name, err)
			}
			categoryIDs = append(categoryIDs, cat.CategoryID)
			t.Logf("Inserted category: %s (ID: %d)", c.name, cat.CategoryID)
		}

		// Sub-categories (under Electronics)
		subCategories := []struct {
			name     string
			slug     string
			parentID int64
		}{
			{"Phones", "phones", categoryIDs[0]},
			{"Laptops", "laptops", categoryIDs[0]},
		}

		for _, c := range subCategories {
			query := models.Categories.Insert().
				Columns(
					models.CategoriesColumnName,
					models.CategoriesColumnSlug,
					models.CategoriesColumnParentID,
				).
				Values(c.name, c.slug, c.parentID).
				Returning(models.CategoriesColumnCategoryID)

			cat, err := models.Categories.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert sub-category %s: %v", c.name, err)
			}
			categoryIDs = append(categoryIDs, cat.CategoryID)
			t.Logf("Inserted sub-category: %s (ID: %d, parent: %d)", c.name, cat.CategoryID, c.parentID)
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
			{"Premium", "premium"},
		}

		for _, tg := range tags {
			query := models.Tags.Insert().
				Columns(models.TagsColumnName, models.TagsColumnSlug).
				Values(tg.name, tg.slug).
				Returning(models.TagsColumnTagID)

			tag, err := models.Tags.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert tag %s: %v", tg.name, err)
			}
			tagIDs = append(tagIDs, tag.TagID)
			t.Logf("Inserted tag: %s (ID: %d)", tg.name, tag.TagID)
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
			{"PHONE-002", "Samsung Galaxy S24", 899.99, "USD", 75},
			{"LAPTOP-001", "MacBook Pro 16", 2499.99, "USD", 25},
			{"SHIRT-001", "Cotton T-Shirt", 29.99, "USD", 200},
		}

		for _, p := range products {
			query := models.Products.Insert().
				Columns(
					models.ProductsColumnSKU,
					models.ProductsColumnName,
					models.ProductsColumnPrice,
					models.ProductsColumnCurrency,
					models.ProductsColumnStockQty,
				).
				Values(p.sku, p.name, p.price, p.currency, p.stockQty).
				Returning(models.ProductsColumnProductID)

			prod, err := models.Products.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert product %s: %v", p.name, err)
			}
			productIDs = append(productIDs, prod.ProductID)
			t.Logf("Inserted product: %s (ID: %d)", p.name, prod.ProductID)
		}
	})

	// ========================================================================
	// INSERT: Product-Category M2M relations
	// ========================================================================
	t.Run("Link Products to Categories", func(t *testing.T) {
		// Phones -> Electronics, Phones
		// Laptops -> Electronics, Laptops
		// Shirt -> Clothing
		links := []struct {
			productID  int64
			categoryID int64
		}{
			{productIDs[0], categoryIDs[0]}, // iPhone -> Electronics
			{productIDs[0], categoryIDs[2]}, // iPhone -> Phones
			{productIDs[1], categoryIDs[0]}, // Samsung -> Electronics
			{productIDs[1], categoryIDs[2]}, // Samsung -> Phones
			{productIDs[2], categoryIDs[0]}, // MacBook -> Electronics
			{productIDs[2], categoryIDs[3]}, // MacBook -> Laptops
			{productIDs[3], categoryIDs[1]}, // Shirt -> Clothing
		}

		for _, l := range links {
			query := models.ProductCategories.Insert().
				Columns(
					models.ProductCategoriesColumnProductID,
					models.ProductCategoriesColumnCategoryID,
				).
				Values(l.productID, l.categoryID)

			_, err := models.ProductCategories.Execute(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to link product %d to category %d: %v", l.productID, l.categoryID, err)
			}
		}
		t.Log("Product-Category links created")
	})

	// ========================================================================
	// INSERT: Product-Tag M2M relations
	// ========================================================================
	t.Run("Link Products to Tags", func(t *testing.T) {
		links := []struct {
			productID int64
			tagID     int64
		}{
			{productIDs[0], tagIDs[0]}, // iPhone -> New
			{productIDs[0], tagIDs[2]}, // iPhone -> Popular
			{productIDs[0], tagIDs[3]}, // iPhone -> Premium
			{productIDs[1], tagIDs[0]}, // Samsung -> New
			{productIDs[1], tagIDs[1]}, // Samsung -> Sale
			{productIDs[2], tagIDs[3]}, // MacBook -> Premium
			{productIDs[3], tagIDs[1]}, // Shirt -> Sale
		}

		for _, l := range links {
			query := models.ProductTags.Insert().
				Columns(
					models.ProductTagsColumnProductID,
					models.ProductTagsColumnTagID,
				).
				Values(l.productID, l.tagID)

			_, err := models.ProductTags.Execute(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to link product %d to tag %d: %v", l.productID, l.tagID, err)
			}
		}
		t.Log("Product-Tag links created")
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
			query := models.Orders.Insert().
				Columns(
					models.OrdersColumnUserID,
					models.OrdersColumnStatus,
					models.OrdersColumnCurrency,
				).
				Values(o.userID, o.status, o.currency).
				Returning(models.OrdersColumnOrderID)

			order, err := models.Orders.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert order for user %d: %v", o.userID, err)
			}
			orderIDs = append(orderIDs, order.OrderID)
			t.Logf("Inserted order: ID=%d, User=%d, Status=%s", order.OrderID, o.userID, o.status)
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
			{orderIDs[0], 1, productIDs[0], 1, 999.99},  // Order 1: iPhone
			{orderIDs[0], 2, productIDs[3], 2, 29.99},   // Order 1: 2x Shirt
			{orderIDs[1], 1, productIDs[2], 1, 2499.99}, // Order 2: MacBook
			{orderIDs[2], 1, productIDs[1], 1, 899.99},  // Order 3: Samsung
		}

		for _, item := range items {
			query := models.OrderItems.Insert().
				Columns(
					models.OrderItemsColumnOrderID,
					models.OrderItemsColumnLineNo,
					models.OrderItemsColumnProductID,
					models.OrderItemsColumnQty,
					models.OrderItemsColumnUnitPrice,
				).
				Values(item.orderID, item.lineNo, item.productID, item.qty, item.unitPrice)

			_, err := models.OrderItems.Execute(ctx, db, query)
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
		query := models.Users.SelectAll()
		users, err := models.Users.Query(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to select users: %v", err)
		}

		if len(users) != 3 {
			t.Errorf("expected 3 users, got %d", len(users))
		}

		for _, u := range users {
			t.Logf("User: ID=%d, Email=%s, Name=%s, Active=%v",
				u.UserID, u.Email, u.FullName, u.IsActive)
		}
	})

	t.Run("Select Products with WHERE", func(t *testing.T) {
		// Select products with price > 500
		query := models.Products.SelectAll().
			Where(models.Products.Price.Gt(500.0))

		products, err := models.Products.Query(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to select products: %v", err)
		}

		if len(products) != 3 {
			t.Errorf("expected 3 expensive products, got %d", len(products))
		}

		for _, p := range products {
			t.Logf("Expensive Product: ID=%d, Name=%s, Price=%.2f",
				p.ProductID, p.Name, p.Price)
		}
	})

	// ========================================================================
	// SELECT: One-to-Many with JOIN
	// ========================================================================
	t.Run("Select Users with Orders (LEFT JOIN)", func(t *testing.T) {
		query := models.UsersOrders.WithJoin(
			models.Users.Table,
			models.Users.SelectAll(),
			dml.LeftJoinType,
		)

		sql, args := query.Build()
		t.Logf("Query: %s, Args: %v", sql, args)

		// Execute raw query to see join results
		rows, err := db.Query(ctx, sql, args...)
		if err != nil {
			t.Fatalf("failed to execute join query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		t.Logf("Join returned %d rows", count)
	})

	// ========================================================================
	// SELECT: BelongsTo with JOIN
	// ========================================================================
	t.Run("Select Orders with User (LEFT JOIN)", func(t *testing.T) {
		query := models.OrdersUser.WithJoin(
			models.Orders.Table,
			models.Orders.SelectAll(),
			dml.LeftJoinType,
		)

		sql, args := query.Build()
		t.Logf("Query: %s, Args: %v", sql, args)
	})

	// ========================================================================
	// SELECT: Many-to-Many with JOIN
	// ========================================================================
	t.Run("Select Products with Categories (M2M JOIN)", func(t *testing.T) {
		query := models.ProductsCategories.WithJoin(
			models.Products.Table,
			models.Products.SelectAll(),
			dml.LeftJoinType,
		)

		sql, args := query.Build()
		t.Logf("M2M Query: %s, Args: %v", sql, args)

		// Execute to verify it works
		rows, err := db.Query(ctx, sql, args...)
		if err != nil {
			t.Fatalf("failed to execute M2M join query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		t.Logf("M2M Join returned %d rows (products × categories)", count)
	})

	t.Run("Select Products with Tags (M2M JOIN)", func(t *testing.T) {
		query := models.ProductsTags.WithJoin(
			models.Products.Table,
			models.Products.SelectAll(),
			dml.LeftJoinType,
		)

		sql, _ := query.Build()
		t.Logf("M2M Tags Query: %s", sql)
	})

	// ========================================================================
	// SELECT: With Relations Auto-Loading
	// ========================================================================
	t.Run("Select Users with Relations (Auto-load)", func(t *testing.T) {
		ctxWithRelations := exec.WithRelations(ctx)

		query := models.Users.SelectAll().
			Where(models.Users.UserID.Eq(userIDs[0]))

		user, err := models.Users.QueryRow(ctxWithRelations, db, query)
		if err != nil {
			t.Fatalf("failed to select user with relations: %v", err)
		}

		t.Logf("User: %s has %d orders", user.FullName, len(user.Orders))
		for _, order := range user.Orders {
			t.Logf("  - Order ID=%d, Status=%s", order.OrderID, order.Status)
		}
	})

	t.Run("Select Orders with Relations (Auto-load)", func(t *testing.T) {
		ctxWithRelations := exec.WithRelations(ctx)

		query := models.Orders.SelectAll().
			Where(models.Orders.OrderID.Eq(orderIDs[0]))

		order, err := models.Orders.QueryRow(ctxWithRelations, db, query)
		if err != nil {
			t.Fatalf("failed to select order with relations: %v", err)
		}

		t.Logf("Order ID=%d belongs to user: %s", order.OrderID, order.User.FullName)
		t.Logf("Order currency: %s (%s)", order.Money.Code, order.Money.Name)
	})

	t.Run("Select Products with M2M Relations (Auto-load)", func(t *testing.T) {
		ctxWithRelations := exec.WithRelations(ctx)

		query := models.Products.SelectAll().
			Where(models.Products.ProductID.Eq(productIDs[0]))

		product, err := models.Products.QueryRow(ctxWithRelations, db, query)
		if err != nil {
			t.Fatalf("failed to select product with relations: %v", err)
		}

		t.Logf("Product: %s", product.Name)
		t.Logf("  Categories (%d):", len(product.Categories))
		for _, cat := range product.Categories {
			t.Logf("    - %s (%s)", cat.Name, cat.Slug)
		}
		t.Logf("  Tags (%d):", len(product.Tags))
		for _, tag := range product.Tags {
			t.Logf("    - %s", tag.Name)
		}
	})

	// ========================================================================
	// UPDATE
	// ========================================================================
	t.Run("Update Product Stock", func(t *testing.T) {
		query := models.Products.Update().
			Set(models.Products.StockQty.Set(100)).
			Where(models.Products.ProductID.Eq(productIDs[0])).
			Returning(models.ProductsColumnProductID, models.ProductsColumnStockQty)

		product, err := models.Products.QueryRow(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to update product: %v", err)
		}

		if product.StockQty != 100 {
			t.Errorf("expected stock_qty=100, got %d", product.StockQty)
		}
		t.Logf("Updated product %d stock to %d", product.ProductID, product.StockQty)
	})

	t.Run("Update Order Status", func(t *testing.T) {
		query := models.Orders.Update().
			Set(models.Orders.Status.Set("PAID")).
			Where(models.Orders.OrderID.Eq(orderIDs[0])).
			Returning(models.OrdersColumnOrderID, models.OrdersColumnStatus)

		order, err := models.Orders.QueryRow(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to update order: %v", err)
		}

		if order.Status != "PAID" {
			t.Errorf("expected status=PAID, got %s", order.Status)
		}
		t.Logf("Updated order %d status to %s", order.OrderID, order.Status)
	})

	// ========================================================================
	// DELETE
	// ========================================================================
	t.Run("Delete Order Item", func(t *testing.T) {
		query := models.OrderItems.Delete().
			Where(
				models.OrderItems.Table.And(
					models.OrderItems.OrderID.Eq(orderIDs[0]),
					models.OrderItems.LineNo.Eq(int32(2)),
				),
			)

		affected, err := models.OrderItems.Execute(ctx, db, query)
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
		query := models.Products.Update().
			Set(models.Products.IsDeleted.Set(true)).
			Where(models.Products.ProductID.Eq(productIDs[3]))

		affected, err := models.Products.Execute(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to soft delete product: %v", err)
		}

		if affected != 1 {
			t.Errorf("expected 1 row affected, got %d", affected)
		}
		t.Logf("Soft deleted product %d", productIDs[3])

		// Verify with select
		selectQuery := models.Products.SelectAll().
			Where(
				models.Products.Table.And(
					models.Products.ProductID.Eq(productIDs[3]),
					models.Products.IsDeleted.Eq(true),
				),
			)

		product, err := models.Products.QueryRow(ctx, db, selectQuery)
		if err != nil {
			t.Fatalf("failed to verify soft deleted product: %v", err)
		}

		if !product.IsDeleted {
			t.Error("product should be marked as deleted")
		}
		t.Logf("Verified product %d is soft deleted", product.ProductID)
	})

	// ========================================================================
	// Complex Query: Products not deleted with categories
	// ========================================================================
	t.Run("Select Active Products Count", func(t *testing.T) {
		query := models.Products.SelectAll().
			Where(models.Products.IsDeleted.Eq(false))

		products, err := models.Products.Query(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to select active products: %v", err)
		}

		// Should be 3 (one was soft deleted)
		if len(products) != 3 {
			t.Errorf("expected 3 active products, got %d", len(products))
		}
		t.Logf("Found %d active products", len(products))
	})

	t.Log("All integration tests passed!")
}
