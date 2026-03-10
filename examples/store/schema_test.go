package main

import (
	"os"
	"strings"
	"testing"

	"github.com/yaroher/ratel/examples/store/models"
	"github.com/yaroher/ratel/pkg/ddl"
)

func TestGenerateSchema(t *testing.T) {
	// Generate schema SQL
	// Order matters for foreign key dependencies:
	// 1. Base tables (currency, users, categories, tags)
	// 2. Tables with FK to base tables (products, orders)
	// 3. Pivot/junction tables (product_categories, product_tags, order_items)
	sqlString := ddl.SchemaSQL(
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

	// Write to generated.sql
	err := os.WriteFile("generated.sql", []byte(sqlString), 0644)
	if err != nil {
		t.Fatalf("failed to write generated.sql: %v", err)
	}

	// Read gold.sql for comparison
	goldSQL, err := os.ReadFile("gold.sql")
	if err != nil {
		t.Fatalf("failed to read gold.sql: %v", err)
	}

	// Read generated.sql
	generatedSQL, err := os.ReadFile("generated.sql")
	if err != nil {
		t.Fatalf("failed to read generated.sql: %v", err)
	}

	// Compare (note: gold.sql has different formatting, so we just log both)
	t.Logf("Generated SQL:\n%s", generatedSQL)
	t.Logf("Gold SQL:\n%s", goldSQL)

	// Basic check that key elements are present
	generated := string(generatedSQL)
	expectedTables := []string{
		"CREATE TABLE IF NOT EXISTS currency",
		"CREATE TABLE IF NOT EXISTS users",
		"CREATE TABLE IF NOT EXISTS products",
		"CREATE TABLE IF NOT EXISTS orders",
		"CREATE TABLE IF NOT EXISTS order_items",
		"CREATE TABLE IF NOT EXISTS categories",
		"CREATE TABLE IF NOT EXISTS tags",
		"CREATE TABLE IF NOT EXISTS product_categories",
		"CREATE TABLE IF NOT EXISTS product_tags",
	}

	for _, table := range expectedTables {
		if !strings.Contains(generated, table) {
			t.Errorf("expected table creation not found: %s", table)
		}
	}

	expectedIndexes := []string{
		"CREATE INDEX IF NOT EXISTS ix_products_not_deleted",
		"CREATE INDEX IF NOT EXISTS ix_orders_user_created",
		"CREATE INDEX IF NOT EXISTS ix_order_items_product",
	}

	for _, idx := range expectedIndexes {
		if !strings.Contains(generated, idx) {
			t.Errorf("expected index creation not found: %s", idx)
		}
	}

	// Verify RLS post_statements
	expectedRLS := []string{
		"ALTER TABLE users ENABLE ROW LEVEL SECURITY",
		"CREATE POLICY users_own_data ON users",
		"CREATE POLICY users_insert ON users",
		"ALTER TABLE orders ENABLE ROW LEVEL SECURITY",
		"CREATE POLICY orders_own_data ON orders",
		"ALTER TABLE order_items ENABLE ROW LEVEL SECURITY",
		"CREATE POLICY order_items_own_data ON order_items",
	}

	for _, rls := range expectedRLS {
		if !strings.Contains(generated, rls) {
			t.Errorf("expected RLS statement not found: %s", rls)
		}
	}
}
