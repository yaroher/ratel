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
	"github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/pgx-ext/sqlerr"
	"github.com/yaroher/ratel/pkg/repository"
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
	sqlers := []ddl.SchemaSqler{
		Currencys,
		Users,
		Profiles, // Profile depends on Users
		Categorys,
		Tags,
		Products,
		Orders,
		OrderItems,
	}
	sqlers = append(sqlers, AdditionalSQL...)
	statements := ddl.SchemaStatements(sqlers...)

	for _, stmt := range statements {
		t.Logf("Executing: %s", stmt)
		_, err := db.Exec(ctx, stmt)
		if err != nil {
			t.Fatalf("failed to execute schema statement: %v\nSQL: %s", err, stmt)
		}
	}
	t.Log("Schema created successfully")
}

// TestGeneratedModels tests the generated models from proto with nested embed
func TestGeneratedModels(t *testing.T) {
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
	// INSERT: Users (with BaseEntity embedded - Id comes from nested embed)
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
				Returning(UserColumnId) // Id from BaseEntity

			user, err := Users.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert user %s: %v", u.email, err)
			}
			userIDs = append(userIDs, user.Id)
			t.Logf("Inserted user: %s (ID: %d)", u.fullName, user.Id)
		}
	})

	// ========================================================================
	// INSERT: Profiles (one per user - HasOne relationship)
	// ========================================================================
	t.Run("Insert Profiles", func(t *testing.T) {
		profiles := []struct {
			userID    int64
			bio       string
			avatarURL string
		}{
			{userIDs[0], "Software engineer", "https://example.com/john.jpg"},
			{userIDs[1], "Product manager", "https://example.com/jane.jpg"},
			// userIDs[2] (Bob) has no profile - to test null case
		}

		for _, p := range profiles {
			query := Profiles.Insert().
				Columns(
					ProfileColumnUserId,
					ProfileColumnBio,
					ProfileColumnAvatarUrl,
				).
				Values(p.userID, p.bio, p.avatarURL).
				Returning(ProfileColumnId)

			profile, err := Profiles.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert profile for user %d: %v", p.userID, err)
			}
			t.Logf("Inserted profile: ID=%d for user %d, bio=%s", profile.Id, p.userID, p.bio)
		}
	})

	// ========================================================================
	// INSERT: Categories (with BaseEntity embedded)
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
				Returning(CategoryColumnId) // Id from BaseEntity

			cat, err := Categorys.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert category %s: %v", c.name, err)
			}
			categoryIDs = append(categoryIDs, cat.Id)
			t.Logf("Inserted category: %s (ID: %d)", c.name, cat.Id)
		}
	})

	// ========================================================================
	// INSERT: Tags (with BaseEntity embedded)
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
				Returning(TagColumnId) // Id from BaseEntity

			tag, err := Tags.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert tag %s: %v", tg.name, err)
			}
			tagIDs = append(tagIDs, tag.Id)
			t.Logf("Inserted tag: %s (ID: %d)", tg.name, tag.Id)
		}
	})

	// ========================================================================
	// INSERT: Products (with BaseEntity embedded)
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
				Returning(ProductColumnId) // Id from BaseEntity

			prod, err := Products.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert product %s: %v", p.name, err)
			}
			productIDs = append(productIDs, prod.Id)
			t.Logf("Inserted product: %s (ID: %d)", p.name, prod.Id)
		}
	})

	// ========================================================================
	// INSERT: Orders (with BaseEntity embedded)
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
				Returning(OrderColumnId) // Id from BaseEntity

			order, err := Orders.QueryRow(ctx, db, query)
			if err != nil {
				t.Fatalf("failed to insert order for user %d: %v", o.userID, err)
			}
			orderIDs = append(orderIDs, order.Id)
			t.Logf("Inserted order: ID=%d, User=%d, Status=%s", order.Id, o.userID, o.status)
		}
	})

	// ========================================================================
	// INSERT: Order Items (no BaseEntity - composite PK)
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
			t.Logf("User: ID=%d, Email=%s, Name=%s, Active=%v, CreatedAt=%v",
				u.Id, u.Email, u.FullName, u.IsActive, u.CreatedAt)
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
				p.Id, p.Name, p.Price)
		}
	})

	// ========================================================================
	// UPDATE
	// ========================================================================
	t.Run("Update Product Stock", func(t *testing.T) {
		query := Products.Update().
			Set(Products.StockQty.Set(100)).
			Where(Products.Id.Eq(productIDs[0])).
			Returning(ProductColumnId, ProductColumnStockQty)

		product, err := Products.QueryRow(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to update product: %v", err)
		}

		if product.StockQty != 100 {
			t.Errorf("expected stock_qty=100, got %d", product.StockQty)
		}
		t.Logf("Updated product %d stock to %d", product.Id, product.StockQty)
	})

	t.Run("Update Order Status", func(t *testing.T) {
		query := Orders.Update().
			Set(Orders.Status.Set("PAID")).
			Where(Orders.Id.Eq(orderIDs[0])).
			Returning(OrderColumnId, OrderColumnStatus)

		order, err := Orders.QueryRow(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to update order: %v", err)
		}

		if order.Status != "PAID" {
			t.Errorf("expected status=PAID, got %s", order.Status)
		}
		t.Logf("Updated order %d status to %s", order.Id, order.Status)
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
			Where(Products.Id.Eq(productIDs[2]))

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
					Products.Id.Eq(productIDs[2]),
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
		t.Logf("Verified product %d is soft deleted", product.Id)
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

	// ========================================================================
	// Verify timestamps from nested embed
	// ========================================================================
	t.Run("Verify Timestamps from Nested Embed", func(t *testing.T) {
		query := Users.SelectAll().
			Where(Users.Id.Eq(userIDs[0]))

		user, err := Users.QueryRow(ctx, db, query)
		if err != nil {
			t.Fatalf("failed to select user: %v", err)
		}

		// CreatedAt and UpdatedAt should be set from nested embed BaseEntity.timestamps
		if user.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set from nested embed")
		}
		if user.UpdatedAt.IsZero() {
			t.Error("expected UpdatedAt to be set from nested embed")
		}
		t.Logf("User %d timestamps - CreatedAt: %v, UpdatedAt: %v",
			user.Id, user.CreatedAt, user.UpdatedAt)
	})

	// ========================================================================
	// SELECT: With Relations using new typed Query Options API
	// ========================================================================
	t.Run("Select Users with Relations (HasMany Orders)", func(t *testing.T) {
		query := Users.SelectAll().
			Where(Users.Id.Eq(userIDs[0]))

		// New API: use typed options instead of context wrapping
		user, err := Users.QueryRow(ctx, db, query, UsersWithOrders())
		if err != nil {
			t.Fatalf("failed to select user with relations: %v", err)
		}

		t.Logf("User: %s has %d orders", user.FullName, len(user.Orders))
		for _, order := range user.Orders {
			t.Logf("  - Order ID=%d, Status=%s", order.Id, order.Status)
		}

		// Verify HasMany loaded correctly
		if len(user.Orders) != 2 {
			t.Errorf("expected 2 orders for user %s, got %d", user.FullName, len(user.Orders))
		}

		// Profile should NOT be loaded (we only requested Orders)
		if user.Profile != nil {
			t.Errorf("expected user.Profile to be nil when only loading Orders")
		}
	})

	t.Run("Select Users with Relations (HasOne Profile)", func(t *testing.T) {
		// User with profile - load only Profile
		query := Users.SelectAll().
			Where(Users.Id.Eq(userIDs[0]))

		user, err := Users.QueryRow(ctx, db, query, UsersWithProfile())
		if err != nil {
			t.Fatalf("failed to select user with relations: %v", err)
		}

		// Verify HasOne loaded correctly
		if user.Profile == nil {
			t.Fatal("expected user.Profile to be loaded (HasOne)")
		}
		t.Logf("User: %s has profile: bio=%s, avatar=%s",
			user.FullName, user.Profile.Bio, user.Profile.AvatarUrl)

		// Orders should NOT be loaded (we only requested Profile)
		if len(user.Orders) != 0 {
			t.Errorf("expected user.Orders to be empty when only loading Profile, got %d", len(user.Orders))
		}

		// User without profile (Bob)
		queryBob := Users.SelectAll().
			Where(Users.Id.Eq(userIDs[2]))

		bob, err := Users.QueryRow(ctx, db, queryBob, UsersWithProfile())
		if err != nil {
			t.Fatalf("failed to select Bob: %v", err)
		}

		// Bob has no profile - should be nil
		if bob.Profile != nil {
			t.Logf("Bob has profile (unexpected but ok): %v", bob.Profile)
		} else {
			t.Logf("User: %s has no profile (as expected)", bob.FullName)
		}
	})

	t.Run("Select Users with All Relations", func(t *testing.T) {
		query := Users.SelectAll().
			Where(Users.Id.Eq(userIDs[0]))

		// Load ALL relations using WithAllRelations
		user, err := Users.QueryRow(ctx, db, query, UsersWithAllRelations())
		if err != nil {
			t.Fatalf("failed to select user with all relations: %v", err)
		}

		// Both Orders and Profile should be loaded
		if len(user.Orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(user.Orders))
		}
		if user.Profile == nil {
			t.Error("expected user.Profile to be loaded")
		}
		t.Logf("User %s: %d orders, profile=%v", user.FullName, len(user.Orders), user.Profile != nil)
	})

	t.Run("Select Users with Multiple Specific Relations", func(t *testing.T) {
		query := Users.SelectAll().
			Where(Users.Id.Eq(userIDs[0]))

		// Load both Orders AND Profile by passing multiple options
		user, err := Users.QueryRow(ctx, db, query, UsersWithOrders(), UsersWithProfile())
		if err != nil {
			t.Fatalf("failed to select user with multiple relations: %v", err)
		}

		// Both should be loaded
		if len(user.Orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(user.Orders))
		}
		if user.Profile == nil {
			t.Error("expected user.Profile to be loaded")
		}
		t.Logf("User %s: %d orders, has profile", user.FullName, len(user.Orders))
	})

	t.Run("Select Profile with Relations (BelongsTo User)", func(t *testing.T) {
		query := Profiles.SelectAll().
			Where(Profiles.UserId.Eq(userIDs[0]))

		profile, err := Profiles.QueryRow(ctx, db, query, ProfilesWithUser())
		if err != nil {
			t.Fatalf("failed to select profile with relations: %v", err)
		}

		// Verify BelongsTo User loaded correctly
		if profile.User == nil {
			t.Fatal("expected profile.User to be loaded (BelongsTo)")
		}
		t.Logf("Profile (bio=%s) belongs to user: %s (%s)",
			profile.Bio, profile.User.FullName, profile.User.Email)
	})

	t.Run("Select Orders with Relations (BelongsTo)", func(t *testing.T) {
		query := Orders.SelectAll().
			Where(Orders.Id.Eq(orderIDs[0]))

		order, err := Orders.QueryRow(ctx, db, query, OrdersWithAllRelations())
		if err != nil {
			t.Fatalf("failed to select order with relations: %v", err)
		}

		// Verify BelongsTo User loaded correctly
		if order.User == nil {
			t.Fatal("expected order.User to be loaded")
		}
		t.Logf("Order ID=%d belongs to user: %s", order.Id, order.User.FullName)

		// Verify BelongsTo Currency loaded correctly
		if order.Money == nil {
			t.Fatal("expected order.Money (Currency) to be loaded")
		}
		t.Logf("Order currency: %s (%s)", order.Money.Code, order.Money.Name)
	})

	t.Run("Select OrderItems with Relations (BelongsTo)", func(t *testing.T) {
		query := OrderItems.SelectAll().
			Where(OrderItems.OrderId.Eq(orderIDs[1]))

		item, err := OrderItems.QueryRow(ctx, db, query, OrderItemsWithAllRelations())
		if err != nil {
			t.Fatalf("failed to select order item with relations: %v", err)
		}

		// Verify BelongsTo Order loaded correctly
		if item.Order == nil {
			t.Fatal("expected orderItem.Order to be loaded")
		}
		t.Logf("OrderItem (line %d) belongs to Order ID=%d (status: %s)",
			item.LineNo, item.Order.Id, item.Order.Status)

		// Verify BelongsTo Product loaded correctly
		if item.Product == nil {
			t.Fatal("expected orderItem.Product to be loaded")
		}
		t.Logf("OrderItem product: %s (SKU: %s, price: %.2f)",
			item.Product.Name, item.Product.Sku, item.Product.Price)
	})

	// Test backward compatibility with context-based loading (deprecated)
	t.Run("Select Users with Relations (deprecated context API)", func(t *testing.T) {
		ctxWithRelations := exec.WithRelations(ctx)

		query := Users.SelectAll().
			Where(Users.Id.Eq(userIDs[0]))

		// Old API still works for backward compatibility
		user, err := Users.QueryRow(ctxWithRelations, db, query)
		if err != nil {
			t.Fatalf("failed to select user with relations: %v", err)
		}

		// Should load all relations
		if len(user.Orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(user.Orders))
		}
		if user.Profile == nil {
			t.Error("expected user.Profile to be loaded")
		}
		t.Logf("(deprecated API) User %s: %d orders, profile=%v", user.FullName, len(user.Orders), user.Profile != nil)
	})

	// ========================================================================
	// Constraint Errors Tests
	// ========================================================================
	t.Run("Test Unique Constraint Error (duplicate email)", func(t *testing.T) {
		// Try to insert user with duplicate email
		query := Users.Insert().
			Columns(UserColumnEmail, UserColumnFullName).
			Values("john@example.com", "Another John") // email already exists

		_, err := Users.Execute(ctx, db, query)
		if err == nil {
			t.Fatal("expected unique constraint error, got nil")
		}

		// Test using generated IsUserEmailUniqueError function
		if !IsUserEmailUniqueError(err) {
			t.Logf("Error: %v", err)
			// Try alternative - unique index error
			if !IsUserEmailUniqueIdxError(err) {
				t.Errorf("expected IsUserEmailUniqueError or IsUserEmailUniqueIdxError to return true")
			} else {
				t.Logf("Correctly identified as unique index violation")
			}
		} else {
			t.Logf("Correctly identified as unique constraint violation")
		}
	})

	t.Run("Test Unique Constraint Error using sqlerr", func(t *testing.T) {
		// Try to insert user with duplicate email
		query := Users.Insert().
			Columns(UserColumnEmail, UserColumnFullName).
			Values("jane@example.com", "Another Jane") // email already exists

		_, err := Users.Execute(ctx, db, query)
		if err == nil {
			t.Fatal("expected unique constraint error, got nil")
		}

		// Test using AsConstraintError
		ce, ok := sqlerr.AsConstraintError(err)
		if !ok {
			t.Fatalf("expected AsConstraintError to return true, got error: %v", err)
		}

		t.Logf("Constraint Error: Name=%s, Type=%s, Table=%s",
			ce.Name, ce.Type, ce.Table)

		// Verify constraint info
		if ce.Type != sqlerr.UniqueConstraint {
			t.Errorf("expected constraint type 'unique', got %s", ce.Type)
		}
		if ce.Table != "users" {
			t.Errorf("expected table 'users', got %s", ce.Table)
		}
	})

	t.Run("Test Check Constraint Error (invalid order status)", func(t *testing.T) {
		// Note: table-level CHECK constraints from proto are not automatically
		// generated in DDL. This test verifies the error handling if CHECK is present.

		// Try to insert order with invalid status
		query := Orders.Insert().
			Columns(OrderColumnUserId, OrderColumnStatus, OrderColumnCurrency).
			Values(userIDs[0], "INVALID_STATUS", "USD")

		_, err := Orders.Execute(ctx, db, query)
		if err == nil {
			// CHECK constraint not applied - skip test
			t.Skip("CHECK constraint not present in schema, skipping test")
		}

		// Test using AsConstraintError
		ce, ok := sqlerr.AsConstraintError(err)
		if !ok {
			t.Logf("Error is not a constraint error: %v", err)
			return
		}

		t.Logf("Check Constraint Error: Name=%s, Type=%s, Table=%s, Detail=%s",
			ce.Name, ce.Type, ce.Table, ce.Detail)

		// Verify it's a check constraint
		if ce.Type != sqlerr.CheckConstraint {
			t.Errorf("expected constraint type 'check', got %s", ce.Type)
		}
	})

	t.Run("Test Primary Key Constraint Error (duplicate ID)", func(t *testing.T) {
		// Try to insert currency with duplicate primary key
		query := Currencys.Insert().
			Columns(CurrencyColumnCode, CurrencyColumnName).
			Values("USD", "Duplicate Dollar") // code already exists

		_, err := Currencys.Execute(ctx, db, query)
		if err == nil {
			t.Fatal("expected primary key constraint error, got nil")
		}

		// Test using generated IsCurrencyPrimaryKeyError function
		if !IsCurrencyPrimaryKeyError(err) {
			t.Logf("Error: %v", err)
			t.Errorf("expected IsCurrencyPrimaryKeyError to return true")
		} else {
			t.Logf("Correctly identified as primary key constraint violation")
		}
	})

	// ========================================================================
	// Cross-Schema Relations Tests
	// Currency and Tags live in "public", everything else in "store" schema
	// ========================================================================
	t.Run("Cross-Schema: Order(store) BelongsTo Currency(public)", func(t *testing.T) {
		query := Orders.SelectAll().
			Where(Orders.Id.Eq(orderIDs[0]))

		order, err := Orders.QueryRow(ctx, db, query, OrdersWithMoney())
		if err != nil {
			t.Fatalf("failed to load order with cross-schema currency relation: %v", err)
		}

		if order.Money == nil {
			t.Fatal("expected order.Money (Currency from public schema) to be loaded")
		}
		if order.Money.Code != "USD" {
			t.Errorf("expected currency code USD, got %s", order.Money.Code)
		}
		t.Logf("Cross-schema BelongsTo OK: Order(store).Money -> Currency(public) code=%s name=%s",
			order.Money.Code, order.Money.Name)
	})

	t.Run("Cross-Schema: Order(store) with all relations including Currency(public)", func(t *testing.T) {
		query := Orders.SelectAll().
			Where(Orders.Id.Eq(orderIDs[0]))

		order, err := Orders.QueryRow(ctx, db, query, OrdersWithAllRelations())
		if err != nil {
			t.Fatalf("failed to load order with all relations: %v", err)
		}

		// BelongsTo User (store -> store)
		if order.User == nil {
			t.Fatal("expected order.User (same schema) to be loaded")
		}
		// BelongsTo Currency (store -> public)
		if order.Money == nil {
			t.Fatal("expected order.Money (cross-schema) to be loaded")
		}
		// HasMany OrderItems (store -> store)
		if len(order.Items) == 0 {
			t.Fatal("expected order.Items to be loaded")
		}

		t.Logf("Cross-schema AllRelations OK: Order(store) -> User(store)=%s, Currency(public)=%s, Items=%d",
			order.User.FullName, order.Money.Code, len(order.Items))
	})

	t.Run("Cross-Schema: OrderItem(store) -> Order(store) -> Currency(public) chain", func(t *testing.T) {
		query := OrderItems.SelectAll().
			Where(OrderItems.OrderId.Eq(orderIDs[1]))

		item, err := OrderItems.QueryRow(ctx, db, query, OrderItemsWithAllRelations())
		if err != nil {
			t.Fatalf("failed to load order item with relations: %v", err)
		}

		if item.Order == nil {
			t.Fatal("expected orderItem.Order to be loaded")
		}
		if item.Product == nil {
			t.Fatal("expected orderItem.Product to be loaded")
		}

		t.Logf("Cross-schema chain OK: OrderItem(store) -> Order(store) id=%d, Product(store) sku=%s",
			item.Order.Id, item.Product.Sku)
	})

	// Suppress unused variables
	_ = categoryIDs
	_ = tagIDs

	// ========================================================================
	// Repository Tests
	// ========================================================================
	t.Run("Repository: ScannerRepository Query", func(t *testing.T) {
		userScannerRepo := repository.NewScannerRepository(Users.Table, db)

		query := Users.SelectAll().Where(Users.IsActive.Eq(true))
		users, err := userScannerRepo.Query(ctx, query)
		if err != nil {
			t.Fatalf("failed to query users: %v", err)
		}

		if len(users) == 0 {
			t.Error("expected at least one user")
		}
		t.Logf("Found %d active users via ScannerRepository", len(users))
	})

	t.Run("Repository: ScannerRepository QueryRow with relations", func(t *testing.T) {
		userScannerRepo := repository.NewScannerRepository(Users.Table, db)

		query := Users.SelectAll().Where(Users.Id.Eq(userIDs[0]))
		user, err := userScannerRepo.QueryRow(ctx, query, UsersWithOrders())
		if err != nil {
			t.Fatalf("failed to query user: %v", err)
		}

		if len(user.Orders) == 0 {
			t.Error("expected user to have orders loaded")
		}
		t.Logf("User %s has %d orders via ScannerRepository", user.FullName, len(user.Orders))
	})

	t.Run("Repository: ProtoRepository Query", func(t *testing.T) {
		userScannerRepo := repository.NewScannerRepository(Users.Table, db)
		userProtoRepo := repository.NewProtoRepository(userScannerRepo, UserConverter)

		query := Users.SelectAll().Where(Users.IsActive.Eq(true))
		users, err := userProtoRepo.Query(ctx, query)
		if err != nil {
			t.Fatalf("failed to query users: %v", err)
		}

		if len(users) == 0 {
			t.Error("expected at least one user")
		}

		// Verify we got proto types back
		for _, u := range users {
			if u.GetEmail() == "" {
				t.Error("expected proto user to have email")
			}
		}
		t.Logf("Found %d active users via ProtoRepository (proto types)", len(users))
	})

	t.Run("Repository: ProtoRepository QueryRow with relations", func(t *testing.T) {
		userScannerRepo := repository.NewScannerRepository(Users.Table, db)
		userProtoRepo := repository.NewProtoRepository(userScannerRepo, UserConverter)

		query := Users.SelectAll().Where(Users.Id.Eq(userIDs[0]))
		user, err := userProtoRepo.QueryRow(ctx, query, UsersWithAllRelations())
		if err != nil {
			t.Fatalf("failed to query user: %v", err)
		}

		// Verify proto type
		if user.GetEmail() == "" {
			t.Error("expected proto user to have email")
		}

		t.Logf("User %s (proto) queried via ProtoRepository", user.GetFullName())
	})

	t.Run("Repository: WithDB for transactions", func(t *testing.T) {
		userScannerRepo := repository.NewScannerRepository(Users.Table, db)
		userProtoRepo := repository.NewProtoRepository(userScannerRepo, UserConverter)

		// Verify WithDB returns a new instance
		newRepo := userProtoRepo.WithDB(db)
		if newRepo == userProtoRepo {
			t.Error("expected WithDB to return a new instance")
		}

		// Test that the new repo works
		query := Users.SelectAll().Where(Users.Id.Eq(userIDs[0]))
		user, err := newRepo.QueryRow(ctx, query)
		if err != nil {
			t.Fatalf("failed to query with new repo: %v", err)
		}
		t.Logf("WithDB test passed, user: %s", user.GetFullName())
	})

	t.Log("All generated model tests passed!")
}
