package ddl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test aliases
type testTable string

func (t testTable) String() string { return string(t) }

type testColumn string

func (t testColumn) String() string { return string(t) }

func TestTableDDL_QualifiedName_WithoutSchema(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
		},
	)

	assert.Equal(t, "users", tbl.TableName())
	assert.Equal(t, "", tbl.Schema())
}

func TestTableDDL_QualifiedName_WithSchema(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
		},
		WithSchema[testTable, testColumn]("auth"),
	)

	assert.Equal(t, `"auth"."users"`, tbl.TableName())
	assert.Equal(t, "auth", tbl.Schema())
}

func TestTableDDL_SchemaSql_WithoutSchema(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("email", TEXT, WithNotNull[testColumn]()),
		},
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "CREATE TABLE IF NOT EXISTS users")
	assert.Contains(t, stmts[0], "id BIGINT PRIMARY KEY")
	assert.Contains(t, stmts[0], "email TEXT NOT NULL")
}

func TestTableDDL_SchemaSql_WithSchema(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
		},
		WithSchema[testTable, testColumn]("auth"),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 2)
	assert.Equal(t, `CREATE SCHEMA IF NOT EXISTS "auth"`, stmts[0])
	assert.Contains(t, stmts[1], `CREATE TABLE IF NOT EXISTS "auth"."users"`)
}

func TestTableDDL_SchemaSql_WithSchemaAndIndexes(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("email", TEXT),
		},
		WithSchema[testTable, testColumn]("auth"),
		WithIndexes[testTable, testColumn](
			NewIndex[testTable, testColumn]("idx_users_email", `"auth"."users"`).
				OnColumns("email").Unique(),
		),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 3)
	assert.Equal(t, `CREATE SCHEMA IF NOT EXISTS "auth"`, stmts[0])
	assert.Contains(t, stmts[1], `CREATE TABLE IF NOT EXISTS "auth"."users"`)
	assert.Contains(t, stmts[2], "CREATE UNIQUE INDEX")
	assert.Contains(t, stmts[2], "idx_users_email")
}

func TestTableDDL_Dependencies_WithSchema(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"orders",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("user_id", BIGINT,
				WithReferences[testColumn](`"auth"."users"`, "id"),
			),
		},
		WithSchema[testTable, testColumn]("auth"),
	)

	deps := tbl.Dependencies()
	require.Len(t, deps, 1)
	assert.Equal(t, `"auth"."users"`, deps[0])
}

func TestTableDDL_Dependencies_ExcludesSelfReference(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"categories",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("parent_id", BIGINT,
				WithReferences[testColumn]("categories", "id"),
			),
		},
	)

	deps := tbl.Dependencies()
	assert.Empty(t, deps)
}

func TestTableDDL_Dependencies_SelfReferenceWithSchema(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"categories",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("parent_id", BIGINT,
				WithReferences[testColumn](`"store"."categories"`, "id"),
			),
		},
		WithSchema[testTable, testColumn]("store"),
	)

	deps := tbl.Dependencies()
	assert.Empty(t, deps)
}

func TestTableDDL_CompositePrimaryKey(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"order_items",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("order_id", BIGINT),
			NewColumnDDL[testColumn]("product_id", BIGINT),
		},
		WithPrimaryKeyColumns[testTable, testColumn]([]testColumn{"order_id", "product_id"}),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "PRIMARY KEY (order_id, product_id)")
}

func TestTableDDL_NamedPrimaryKey(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"order_items",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("order_id", BIGINT),
			NewColumnDDL[testColumn]("product_id", BIGINT),
		},
		WithTablePrimaryKeyNamed[testTable, testColumn](
			"pk_order_items", []testColumn{"order_id", "product_id"},
		),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "CONSTRAINT pk_order_items PRIMARY KEY (order_id, product_id)")
}

func TestTableDDL_NamedPrimaryKey_OverridesLegacy(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"order_items",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("order_id", BIGINT),
			NewColumnDDL[testColumn]("product_id", BIGINT),
		},
		WithPrimaryKeyColumns[testTable, testColumn]([]testColumn{"order_id"}),
		WithTablePrimaryKeyNamed[testTable, testColumn](
			"pk_order_items", []testColumn{"order_id", "product_id"},
		),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 1)
	// Named takes precedence
	assert.Contains(t, stmts[0], "CONSTRAINT pk_order_items PRIMARY KEY")
	assert.NotContains(t, stmts[0], "\n  PRIMARY KEY (order_id)")
}

func TestTableDDL_UniqueConstraint(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("email", TEXT),
			NewColumnDDL[testColumn]("tenant_id", BIGINT),
		},
		WithUniqueColumns[testTable, testColumn](
			[]testColumn{"email", "tenant_id"},
		),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "UNIQUE (email, tenant_id)")
}

func TestTableDDL_NamedUniqueConstraint(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("email", TEXT),
			NewColumnDDL[testColumn]("tenant_id", BIGINT),
		},
		WithTableUniqueNamed[testTable, testColumn](
			"uq_users_email_tenant", []testColumn{"email", "tenant_id"},
		),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "CONSTRAINT uq_users_email_tenant UNIQUE (email, tenant_id)")
}

func TestTableDDL_CheckConstraint(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"products",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("price", REAL),
		},
		WithTableCheckConstraint[testTable, testColumn](
			"chk_products_price_positive", "price > 0",
		),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "CONSTRAINT chk_products_price_positive CHECK (price > 0)")
}

func TestTableDDL_AllConstraintsTogether(t *testing.T) {
	tbl := NewTableDDL[testTable, testColumn](
		"order_items",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("order_id", BIGINT),
			NewColumnDDL[testColumn]("product_id", BIGINT),
			NewColumnDDL[testColumn]("quantity", INTEGER),
		},
		WithSchema[testTable, testColumn]("store"),
		WithTablePrimaryKeyNamed[testTable, testColumn](
			"pk_order_items", []testColumn{"order_id", "product_id"},
		),
		WithTableUniqueNamed[testTable, testColumn](
			"uq_order_product", []testColumn{"order_id", "product_id"},
		),
		WithTableCheckConstraint[testTable, testColumn](
			"chk_quantity_positive", "quantity > 0",
		),
	)

	stmts := tbl.SchemaSql()
	require.Len(t, stmts, 2)
	assert.Equal(t, `CREATE SCHEMA IF NOT EXISTS "store"`, stmts[0])

	createTable := stmts[1]
	assert.Contains(t, createTable, `CREATE TABLE IF NOT EXISTS "store"."order_items"`)
	assert.Contains(t, createTable, "CONSTRAINT pk_order_items PRIMARY KEY (order_id, product_id)")
	assert.Contains(t, createTable, "CONSTRAINT uq_order_product UNIQUE (order_id, product_id)")
	assert.Contains(t, createTable, "CONSTRAINT chk_quantity_positive CHECK (quantity > 0)")
}

func TestSchemaSortedStatements_WithSchemaQualifiedDeps(t *testing.T) {
	users := NewTableDDL[testTable, testColumn](
		"users",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
		},
		WithSchema[testTable, testColumn]("auth"),
	)

	orders := NewTableDDL[testTable, testColumn](
		"orders",
		[]*ColumnDDL[testColumn]{
			NewColumnDDL[testColumn]("id", BIGINT, WithPrimaryKey[testColumn]()),
			NewColumnDDL[testColumn]("user_id", BIGINT,
				WithReferences[testColumn](`"auth"."users"`, "id"),
			),
		},
		WithSchema[testTable, testColumn]("auth"),
	)

	// orders depends on users, so users should come first regardless of input order
	stmts, err := SchemaSortedStatements(orders, users)
	require.NoError(t, err)

	// Find which CREATE TABLE statement appears first
	// Use "EXISTS" prefix to avoid matching REFERENCES "auth"."users"
	usersIdx := -1
	ordersIdx := -1
	for i, s := range stmts {
		if strings.Contains(s, `EXISTS "auth"."users"`) {
			usersIdx = i
		}
		if strings.Contains(s, `EXISTS "auth"."orders"`) {
			ordersIdx = i
		}
	}
	require.NotEqual(t, -1, usersIdx, "users CREATE TABLE not found")
	require.NotEqual(t, -1, ordersIdx, "orders CREATE TABLE not found")
	assert.Less(t, usersIdx, ordersIdx, "users table should be created before orders")
}
