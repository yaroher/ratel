package integrations

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yaroher/ratel/dml"
	schema2 "github.com/yaroher/ratel/schema"
)

// Определяем column aliases для users и posts
type userCol string

func (c userCol) String() string { return string(c) }

const (
	userColID    userCol = "id"
	userColName  userCol = "name"
	userColEmail userCol = "email"
)

type postCol string

func (c postCol) String() string { return string(c) }

const (
	postColID      postCol = "id"
	postColUserID  postCol = "user_id"
	postColTitle   postCol = "title"
	postColContent postCol = "content"
)

type tagCol string

func (c tagCol) String() string { return string(c) }

const (
	tagColID   tagCol = "id"
	tagColName tagCol = "name"
)

// ============================================================================
// Сгенерированная схема таблиц
// ============================================================================

// UsersTable представляет схему таблицы users
type UsersTable struct {
	TableName string
	Alias     string

	// Колонки
	ID    *schema2.Column[int32, userCol]
	Name  *schema2.Column[string, userCol]
	Email *schema2.Column[string, userCol]
}

// NewUsersTable создает новую схему таблицы users
func NewUsersTable(alias string) *UsersTable {
	if alias == "" {
		alias = "users"
	}
	return &UsersTable{
		TableName: "users",
		Alias:     alias,
		ID:        schema2.IntegerCol(userColID),
		Name:      schema2.TextCol(userColName),
		Email:     schema2.TextCol(userColEmail),
	}
}

// PostsTable представляет схему таблицы posts
type PostsTable struct {
	TableName string
	Alias     string

	// Колонки
	ID      *schema2.Column[int32, postCol]
	UserID  *schema2.Column[int32, postCol]
	Title   *schema2.Column[string, postCol]
	Content *schema2.Column[string, postCol]
}

// NewPostsTable создает новую схему таблицы posts
func NewPostsTable(alias string) *PostsTable {
	if alias == "" {
		alias = "posts"
	}
	return &PostsTable{
		TableName: "posts",
		Alias:     alias,
		ID:        schema2.IntegerCol(postColID),
		UserID:    schema2.IntegerCol(postColUserID),
		Title:     schema2.TextCol(postColTitle),
		Content:   schema2.TextCol(postColContent),
	}
}

// TestRelations демонстрирует работу с связями
func TestRelations(t *testing.T) {
	ctx := context.Background()

	// Создаем PostgreSQL контейнер
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

	// Получаем connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	// Подключаемся к БД
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// Создаем таблицы
	_, err = pool.Exec(ctx, `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		);

		CREATE TABLE posts (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			title TEXT NOT NULL,
			content TEXT
		);

		CREATE TABLE tags (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE post_tags (
			post_id INTEGER NOT NULL REFERENCES posts(id),
			tag_id INTEGER NOT NULL REFERENCES tags(id),
			PRIMARY KEY (post_id, tag_id)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Вставляем тестовые данные
	_, err = pool.Exec(ctx, `
		INSERT INTO users (name, email) VALUES
			('Alice', 'alice@example.com'),
			('Bob', 'bob@example.com');

		INSERT INTO posts (user_id, title, content) VALUES
			(1, 'First Post', 'Content of first post'),
			(1, 'Second Post', 'Content of second post'),
			(2, 'Bob Post', 'Content from Bob');

		INSERT INTO tags (name) VALUES
			('go'),
			('db');

		INSERT INTO post_tags (post_id, tag_id) VALUES
			(1, 1),
			(1, 2),
			(2, 1),
			(3, 2);
	`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Manual INNER JOIN", func(t *testing.T) {
		// Создаем columns
		userName := next.StringColumn[userCol](userColName)
		postTitle := next.StringColumn[postCol](postColTitle)

		// Строим запрос с ручным JOIN
		selectQuery := &dml.SelectQuery[userCol]{
			BaseQuery: dml.BaseQuery[userCol]{
				Ta: "users",
				UsingFields: []userCol{
					userColName,
				},
			},
		}

		// Добавляем INNER JOIN вручную
		selectQuery.InnerJoin("posts", "posts", next.NewRawOnClause[userCol](
			"users.id = posts.user_id",
		))

		selectQuery.Where(userName.Eq("Alice"))

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)
		t.Logf("Args: %v", args)

		// Проверяем, что SQL содержит INNER JOIN
		if sql != "SELECT users.name FROM users AS users INNER JOIN posts AS posts ON users.id = posts.user_id WHERE users.name = $1;" {
			t.Fatalf("unexpected SQL: %s", sql)
		}

		// Выполняем запрос
		rows, err := pool.Query(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			count++
			t.Logf("Found user: %s", name)
		}

		if count != 2 {
			t.Fatalf("expected 2 rows (Alice has 2 posts), got %d", count)
		}

		_ = postTitle // используем для избежания warning
	})

	t.Run("LEFT JOIN", func(t *testing.T) {
		userName := next.StringColumn[userCol](userColName)

		// Создаем пользователя без постов
		_, err := pool.Exec(ctx, `INSERT INTO users (name, email) VALUES ('Charlie', 'charlie@example.com')`)
		if err != nil {
			t.Fatal(err)
		}

		selectQuery := &dml.SelectQuery[userCol]{
			BaseQuery: dml.BaseQuery[userCol]{
				Ta:          "users",
				UsingFields: []userCol{userColName},
			},
		}

		// LEFT JOIN чтобы получить всех пользователей, даже без постов
		selectQuery.LeftJoin("posts", "posts", next.NewRawOnClause[userCol](
			"users.id = posts.user_id",
		))

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)

		rows, err := pool.Query(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			count++
		}

		// 2 поста от Alice + 1 пост от Bob + 1 запись для Charlie (без поста)
		if count != 4 {
			t.Fatalf("expected 4 rows with LEFT JOIN, got %d", count)
		}

		_ = userName
	})

	t.Run("HasMany Relation", func(t *testing.T) {
		// Определяем связь: User HasMany Posts
		userPosts := next.HasMany[userCol]("users", "posts", "user_id", "id")

		selectQuery := &dml.SelectQuery[userCol]{
			BaseQuery: dml.BaseQuery[userCol]{
				Ta:          "users",
				UsingFields: []userCol{userColName},
			},
		}

		// Применяем связь с INNER JOIN
		userPosts.InnerJoin(selectQuery)

		userName := next.StringColumn[userCol](userColName)
		selectQuery.Where(userName.Eq("Alice"))

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)
		t.Logf("Args: %v", args)

		rows, err := pool.Query(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			count++
		}

		if count != 2 {
			t.Fatalf("expected 2 rows (Alice has 2 posts), got %d", count)
		}
	})

	t.Run("BelongsTo Relation", func(t *testing.T) {
		// Определяем связь: Post BelongsTo User
		postUser := next.BelongsTo[postCol]("posts", "users", "user_id", "id")

		selectQuery := &dml.SelectQuery[postCol]{
			BaseQuery: dml.BaseQuery[postCol]{
				Ta:          "posts",
				UsingFields: []postCol{postColTitle},
			},
		}

		// Применяем связь с INNER JOIN
		postUser.InnerJoin(selectQuery)

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)

		rows, err := pool.Query(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var title string
			if err := rows.Scan(&title); err != nil {
				t.Fatal(err)
			}
			count++
			t.Logf("Found post: %s", title)
		}

		if count != 3 {
			t.Fatalf("expected 3 posts, got %d", count)
		}
	})

	t.Run("Relation with Alias", func(t *testing.T) {
		// Используем HasMany с кастомным алиасом
		userPosts := next.HasMany[userCol]("users", "posts", "user_id", "id").
			WithAlias("user_posts")

		selectQuery := &dml.SelectQuery[userCol]{
			BaseQuery: dml.BaseQuery[userCol]{
				Ta:          "users",
				UsingFields: []userCol{userColName},
			},
		}

		userPosts.LeftJoin(selectQuery)

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)

		// Проверяем, что используется алиас
		if sql != "SELECT users.name FROM users AS users LEFT JOIN posts AS user_posts ON user_posts.user_id = users.id;" {
			t.Fatalf("unexpected SQL with alias: %s", sql)
		}

		_ = args
	})

	t.Run("ManyToMany Relation", func(t *testing.T) {
		// Определяем связь: Post ManyToMany Tags
		postTags := next.ManyToMany[postCol]("posts", "tags", "post_tags", "post_id", "tag_id")

		selectQuery := &dml.SelectQuery[postCol]{
			BaseQuery: dml.BaseQuery[postCol]{
				Ta:          "posts",
				UsingFields: []postCol{postColTitle},
			},
		}

		postTags.InnerJoin(selectQuery)

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)
		t.Logf("Args: %v", args)

		rows, err := pool.Query(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var title string
			if err := rows.Scan(&title); err != nil {
				t.Fatal(err)
			}
			count++
		}

		// 4 связки постов и тегов
		if count != 4 {
			t.Fatalf("expected 4 rows for many-to-many, got %d", count)
		}
	})

	t.Run("ManyToMany Relation with Aliases", func(t *testing.T) {
		// ManyToMany с алиасами для целевой и промежуточной таблицы
		postTags := next.ManyToMany[postCol]("posts", "tags", "post_tags", "post_id", "tag_id").
			WithAlias("t").
			WithThroughAlias("pt")

		selectQuery := &dml.SelectQuery[postCol]{
			BaseQuery: dml.BaseQuery[postCol]{
				Ta:          "posts",
				UsingFields: []postCol{postColTitle},
			},
		}

		postTags.LeftJoin(selectQuery)

		sql, args := selectQuery.Build()
		t.Logf("SQL: %s", sql)

		if sql != "SELECT posts.title FROM posts AS posts LEFT JOIN post_tags AS pt ON posts.id = pt.post_id LEFT JOIN tags AS t ON pt.tag_id = t.id;" {
			t.Fatalf("unexpected SQL with aliases: %s", sql)
		}

		_ = args
	})
}
