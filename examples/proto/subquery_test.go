package storepb

import (
	"strings"
	"testing"

	"github.com/yaroher/ratel/pkg/types"
)

// TestSelectQueryImplementsQuery verifies that SelectQuery satisfies types.Query.
func TestSelectQueryImplementsQuery(t *testing.T) {
	q := Users.Select(UserColumnId).Where(Users.IsActive.Eq(true))
	var _ types.Query = q // compile-time check
	if q.TableAlias() != string(UserAliasName) {
		t.Errorf("TableAlias() = %q, want %q", q.TableAlias(), UserAliasName)
	}
}

// TestInOfSubquery verifies column.InOf(subquery) generates correct SQL.
func TestInOfSubquery(t *testing.T) {
	// SELECT ... FROM orders WHERE orders.user_id IN (SELECT users.id FROM users WHERE ...)
	subquery := Users.Select(UserColumnId).Where(Users.IsActive.Eq(true))
	query := Orders.SelectAll().Where(Orders.UserId.InOf(subquery))

	sql, args := query.Build()
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	assertContains(t, sql, "IN (SELECT")
	assertContains(t, sql, `users.id FROM "store"."users"`)
	assertContains(t, sql, "users.is_active = $")
	if len(args) != 1 || args[0] != true {
		t.Errorf("expected args=[true], got %v", args)
	}
}

// TestAnyOfSubquery verifies column.AnyOf(subquery) generates correct SQL.
func TestAnyOfSubquery(t *testing.T) {
	subquery := Users.Select(UserColumnId).Where(Users.IsActive.Eq(true))
	query := Orders.SelectAll().Where(Orders.UserId.AnyOf(subquery))

	sql, args := query.Build()
	t.Logf("SQL: %s", sql)

	assertContains(t, sql, "= ANY (SELECT")
	assertContains(t, sql, `users.id FROM "store"."users"`)
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

// TestExistsOfSubquery verifies table.ExistsOf(subquery) generates correct SQL.
func TestExistsOfSubquery(t *testing.T) {
	// SELECT ... FROM users WHERE EXISTS (SELECT 1 FROM orders WHERE ...)
	subquery := Orders.Select1().Where(Orders.UserId.Eq(int64(1)))
	query := Users.SelectAll().Where(Users.Table.ExistsOf(subquery))

	sql, args := query.Build()
	t.Logf("SQL: %s", sql)

	assertContains(t, sql, `EXISTS (SELECT 1 FROM "store"."orders"`)
	assertContains(t, sql, "orders.user_id = $")
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

// TestNotExistsOfSubquery verifies table.NotExistsOf(subquery) generates correct SQL.
func TestNotExistsOfSubquery(t *testing.T) {
	subquery := Orders.Select1().Where(Orders.UserId.Eq(int64(1)))
	query := Users.SelectAll().Where(Users.Table.NotExistsOf(subquery))

	sql, args := query.Build()
	t.Logf("SQL: %s", sql)

	assertContains(t, sql, `NOT EXISTS (SELECT 1 FROM "store"."orders"`)
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

// TestNestedSubqueries verifies multiple levels of subquery nesting.
func TestNestedSubqueries(t *testing.T) {
	// Users who have orders containing a specific product
	innerSub := OrderItems.Select(OrderItemColumnOrderId).
		Where(OrderItems.ProductId.Eq(int64(42)))
	outerSub := Orders.Select(OrderColumnUserId).
		Where(Orders.Id.InOf(innerSub))
	query := Users.SelectAll().Where(Users.Id.InOf(outerSub))

	sql, args := query.Build()
	t.Logf("SQL: %s", sql)

	// Should have two nested IN (SELECT ...) clauses
	if cnt := strings.Count(sql, "IN (SELECT"); cnt != 2 {
		t.Errorf("expected 2 nested IN (SELECT, got %d in: %s", cnt, sql)
	}
	if len(args) != 1 || args[0] != int64(42) {
		t.Errorf("expected args=[42], got %v", args)
	}
}

// TestSubqueryParameterIndexContinuity verifies parameter indices increment across subqueries.
func TestSubqueryParameterIndexContinuity(t *testing.T) {
	subquery := Users.Select(UserColumnId).Where(Users.IsActive.Eq(true))
	query := Orders.SelectAll().Where(
		Orders.Status.Eq("PENDING"),
		Orders.UserId.InOf(subquery),
	)

	sql, args := query.Build()
	t.Logf("SQL: %s", sql)

	// First param ($1) is "PENDING", second ($2) is true from the subquery
	assertContains(t, sql, "orders.status = $1")
	assertContains(t, sql, "users.is_active = $2")
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "PENDING" {
		t.Errorf("args[0] = %v, want %q", args[0], "PENDING")
	}
	if args[1] != true {
		t.Errorf("args[1] = %v, want true", args[1])
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected SQL to contain %q, got:\n%s", substr, s)
	}
}
