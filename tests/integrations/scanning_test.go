package integrations

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yaroher/ratel/pkg/scanning"
	"github.com/yaroher/ratel/pkg/schema"
)

type scanUserCol string

func (c scanUserCol) String() string { return string(c) }

const (
	scanUserColID   scanUserCol = "id"
	scanUserColName scanUserCol = "name"
)

type scanUser struct {
	ID   int32
	Name string
}

type scanUserScanner struct {
	*scanning.BaseTargeter[scanUserCol]
	User *scanUser
}

func newScanUserScanner() *scanUserScanner {
	user := &scanUser{}
	bt := scanning.NewBaseTargeter[scanUserCol](
		scanning.FieldAccess[scanUserCol]{
			Name:   "id",
			Target: func() any { return &user.ID },
			Value:  func() any { return user.ID },
		},
		scanning.FieldAccess[scanUserCol]{
			Name:   "name",
			Target: func() any { return &user.Name },
			Value:  func() any { return user.Name },
		},
	)
	return &scanUserScanner{BaseTargeter: bt, User: user}
}

func TestScanTargetResolver(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (name) VALUES
			('Alice'),
			('Bob');
	`)
	if err != nil {
		t.Fatal(err)
	}

	users := schema.NewTable[scanUserCol, *scanUserScanner](
		"users",
		newScanUserScanner,
		scanUserColID,
		scanUserColName,
	)

	nameCol := schema.StringColumn[scanUserCol](scanUserColName)
	queryRow := users.Select(scanUserColName, scanUserColID)
	queryRow.Where(nameCol.Eq("Alice"))

	row, err := users.QueryRow(ctx, pool, queryRow)
	if err != nil {
		t.Fatal(err)
	}
	if row.User.ID == 0 || row.User.Name != "Alice" {
		t.Fatalf("unexpected row: id=%d name=%s", row.User.ID, row.User.Name)
	}

	queryAll := users.Select(scanUserColName, scanUserColID)
	rows, err := users.Query(ctx, pool, queryAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}
