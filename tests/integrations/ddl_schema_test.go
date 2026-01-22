package integrations

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yaroher/ratel/next"
	schema2 "github.com/yaroher/ratel/schema"
	"github.com/yaroher/ratel/sqlscan"
)

type ddlUserCol string

func (c ddlUserCol) String() string { return string(c) }

const (
	ddlUserColID   ddlUserCol = "id"
	ddlUserColName ddlUserCol = "name"
)

type ddlPostCol string

func (c ddlPostCol) String() string { return string(c) }

const (
	ddlPostColID     ddlPostCol = "id"
	ddlPostColUserID ddlPostCol = "user_id"
	ddlPostColTitle  ddlPostCol = "title"
)

type dummyUserScanner struct {
	*sqlscan.BaseTargeter[ddlUserCol]
}

func newDummyUserScanner() *dummyUserScanner {
	return &dummyUserScanner{BaseTargeter: sqlscan.NewBaseTargeter[ddlUserCol]()}
}

type dummyPostScanner struct {
	*sqlscan.BaseTargeter[ddlPostCol]
}

func newDummyPostScanner() *dummyPostScanner {
	return &dummyPostScanner{BaseTargeter: sqlscan.NewBaseTargeter[ddlPostCol]()}
}

func TestDDLFromSchema(t *testing.T) {
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

	users := schema2.NewTableWithDDL[ddlUserCol, *dummyUserScanner](
		"users",
		newDummyUserScanner,
		[]schema2.DDLColumn{
			schema2.NewColumn[int32, ddlUserCol](ddlUserColID).Type("SERIAL").PrimaryKey(),
			schema2.NewColumn[string, ddlUserCol](ddlUserColName).Type("TEXT").NotNull(),
		},
		nil,
		ddlUserColID,
		ddlUserColName,
	)

	postUser := next.BelongsTo[ddlPostCol]("posts", "users", "user_id", "id")
	posts := schema2.NewTableWithDDL[ddlPostCol, *dummyPostScanner](
		"posts",
		newDummyPostScanner,
		[]schema2.DDLColumn{
			schema2.NewColumn[int32, ddlPostCol](ddlPostColID).Type("SERIAL").PrimaryKey(),
			schema2.NewColumn[int32, ddlPostCol](ddlPostColUserID).Type("INTEGER").NotNull(),
			schema2.NewColumn[string, ddlPostCol](ddlPostColTitle).Type("TEXT").NotNull(),
		},
		[]*next.Relation[ddlPostCol]{postUser},
		ddlPostColID,
		ddlPostColUserID,
		ddlPostColTitle,
	)

	statements := schema2.BuildSchema(users, posts)
	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to exec ddl: %s", err)
		}
	}

	_, err = pool.Exec(ctx, `INSERT INTO users (name) VALUES ('Alice')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO posts (user_id, title) VALUES (1, 'Hello')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO posts (user_id, title) VALUES (999, 'Bad')`)
	if err == nil {
		t.Fatal("expected FK violation, got nil")
	}
}
