package examples

import (
	"fmt"

	"github.com/yaroher/ratel/ddl"
)

// Example table and column aliases
type TableName string

func (t TableName) String() string { return string(t) }

type ColumnName string

func (c ColumnName) String() string { return string(c) }

const (
	UsersTable TableName = "users"
	PostsTable TableName = "posts"
)

const (
	UserID    ColumnName = "id"
	UserEmail ColumnName = "email"
	UserName  ColumnName = "name"
	PostID    ColumnName = "id"
	PostTitle ColumnName = "title"
	PostTags  ColumnName = "tags"
)

// IndexExample demonstrates how to use Index
func IndexExample() {
	// 1. Простой индекс
	emailIdx := ddl.NewIndex(TableName("idx_users_email"), UsersTable).
		OnColumns(UserEmail)

	sql, _ := emailIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

	// 2. Уникальный индекс
	uniqueEmailIdx := ddl.NewIndex(TableName("idx_users_email_unique"), UsersTable).
		OnColumns(UserEmail).
		Unique()

	sql, _ = uniqueEmailIdx.Build()
	fmt.Println(sql)
	// CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (email);

	// 3. Составной индекс
	nameEmailIdx := ddl.NewIndex(TableName("idx_users_name_email"), UsersTable).
		OnColumns(UserName, UserEmail)

	sql, _ = nameEmailIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_users_name_email ON users (name, email);

	// 4. Индекс с методом (btree, hash, gin, gist)
	ginIdx := ddl.NewIndex(TableName("idx_posts_tags"), PostsTable).
		OnColumns(PostTags).
		Using("gin")

	sql, _ = ginIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);

	// 5. Частичный индекс (partial index)
	activeUsersIdx := ddl.NewIndex(TableName("idx_users_active"), UsersTable).
		OnColumns(UserEmail).
		Where("deleted_at IS NULL")

	sql, _ = activeUsersIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_users_active ON users (email) WHERE deleted_at IS NULL;

	// 6. Concurrent индекс (для больших таблиц)
	concurrentIdx := ddl.NewIndex(TableName("idx_users_name"), UsersTable).
		OnColumns(UserName).
		Concurrently()

	sql, _ = concurrentIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_name ON users (name);

	// 7. Использование Create() для большего контроля
	createStmt := emailIdx.Create().IfNotExists()
	sql, _ = createStmt.Build()
	fmt.Println(sql)

	// 8. Удаление индекса
	dropStmt := emailIdx.Drop().IfExists()
	sql, _ = dropStmt.Build()
	fmt.Println(sql)
	// DROP INDEX IF EXISTS idx_users_email;
}

// IndexBuilderExample demonstrates how to build multiple indexes at once
func IndexBuilderExample() {
	builder := ddl.NewIndexBuilder(UsersTable)

	// Добавляем обычный индекс
	builder.Add(TableName("idx_users_email"), UserEmail)

	// Добавляем уникальный индекс
	builder.AddUnique(TableName("idx_users_id"), UserID)

	// Добавляем составной индекс с дополнительными настройками
	builder.Add(TableName("idx_users_name_email"), UserName, UserEmail).
		Using("btree")

	// Добавляем частичный индекс
	builder.Add(TableName("idx_users_active"), UserEmail).
		Where("deleted_at IS NULL")

	// Генерируем все CREATE INDEX statements
	sql, _ := builder.Build()
	fmt.Println(sql)
	/*
		CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_id ON users (id);
		CREATE INDEX IF NOT EXISTS idx_users_name_email ON users USING btree (name, email);
		CREATE INDEX IF NOT EXISTS idx_users_active ON users (email) WHERE deleted_at IS NULL;
	*/

	// Получаем все индексы
	for _, idx := range builder.Indexes() {
		fmt.Printf("Index: %s on table %s\n", idx.Name(), idx.Table())
		fmt.Printf("  Columns: %v\n", idx.Columns())
		fmt.Printf("  Unique: %v\n", idx.IsUnique())
		if idx.Method() != "" {
			fmt.Printf("  Method: %s\n", idx.Method())
		}
		if idx.Predicate() != "" {
			fmt.Printf("  Where: %s\n", idx.Predicate())
		}
	}
}

// ComplexIndexExample demonstrates advanced index features
func ComplexIndexExample() {
	// GIN индекс для полнотекстового поиска
	fullTextIdx := ddl.NewIndex(TableName("idx_posts_title_fts"), PostsTable).
		OnColumns(ColumnName("to_tsvector('english', title)")).
		Using("gin")

	sql, _ := fullTextIdx.Build()
	fmt.Println(sql)

	// Уникальный частичный индекс
	uniqueActiveEmailIdx := ddl.NewIndex(TableName("idx_users_email_active_unique"), UsersTable).
		OnColumns(UserEmail).
		Unique().
		Where("deleted_at IS NULL")

	sql, _ = uniqueActiveEmailIdx.Build()
	fmt.Println(sql)
	// CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active_unique ON users (email) WHERE deleted_at IS NULL;

	// Concurrent создание с GIN для массивов
	tagsIdx := ddl.NewIndex(TableName("idx_posts_tags_gin"), PostsTable).
		OnColumns(PostTags).
		Using("gin").
		Concurrently()

	sql, _ = tagsIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_tags_gin ON posts USING gin (tags);
}
