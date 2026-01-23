package examples

import (
	"fmt"

	"github.com/yaroher/ratel/ddl"
)

// ColumnDDL and TableDML aliases for examples
type Col string

func (c Col) String() string { return string(c) }

type Tab string

func (t Tab) String() string { return string(t) }

const (
	UsersTab Tab = "users"
	PostsTab Tab = "posts"
)

const (
	ColID        Col = "id"
	ColName      Col = "name"
	ColEmail     Col = "email"
	ColAge       Col = "age"
	ColCreatedAt Col = "created_at"
	ColUpdatedAt Col = "updated_at"
	ColUserID    Col = "user_id"
	ColTitle     Col = "title"
	ColContent   Col = "content"
	ColTags      Col = "tags"
)

// ColumnOptionsExample demonstrates how to use NewColumnDDL with options
func ColumnOptionsExample() {
	// 1. Простая колонка
	simpleCol := ddl.NewColumnDDL(ColName, ddl.Text())
	fmt.Println(simpleCol.Build())
	// name TEXT

	// 2. Колонка с NOT NULL
	notNullCol := ddl.NewColumnDDL(
		ColEmail,
		ddl.Text(),
		ddl.WithNotNull[Col](),
	)
	fmt.Println(notNullCol.Build())
	// email TEXT NOT NULL

	// 3. Уникальная колонка с NOT NULL
	uniqueCol := ddl.NewColumnDDL(
		ColEmail,
		ddl.Text(),
		ddl.WithNotNull[Col](),
		ddl.WithUnique[Col](),
	)
	fmt.Println(uniqueCol.Build())
	// email TEXT NOT NULL UNIQUE

	// 4. Primary Key колонка
	pkCol := ddl.NewColumnDDL(
		ColID,
		ddl.Serial(),
		ddl.WithPrimaryKey[Col](),
	)
	fmt.Println(pkCol.Build())
	// id SERIAL PRIMARY KEY

	// 5. Колонка с DEFAULT значением
	timestampCol := ddl.NewColumnDDL(
		ColCreatedAt,
		ddl.Timestamp(),
		ddl.WithNotNull[Col](),
		ddl.WithDefault[Col]("CURRENT_TIMESTAMP"),
	)
	fmt.Println(timestampCol.Build())
	// created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

	// 6. Foreign Key колонка
	fkCol := ddl.NewColumnDDL(
		ColUserID,
		ddl.Integer(),
		ddl.WithNotNull[Col](),
		ddl.WithReferences[Col]("users", "id"),
		ddl.WithOnDelete[Col]("CASCADE"),
		ddl.WithOnUpdate[Col]("CASCADE"),
	)
	fmt.Println(fkCol.Build())
	// user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE

	// 7. Колонка с CHECK constraint
	ageCol := ddl.NewColumnDDL(
		ColAge,
		ddl.Integer(),
		ddl.WithCheck[Col]("age >= 0 AND age <= 150"),
	)
	fmt.Println(ageCol.Build())
	// age INTEGER CHECK (age >= 0 AND age <= 150)

	// 8. Комбинация всех опций
	complexCol := ddl.NewColumnDDL(
		ColAge,
		ddl.Integer(),
		ddl.WithNotNull[Col](),
		ddl.WithDefault[Col]("18"),
		ddl.WithCheck[Col]("age >= 18"),
	)
	fmt.Println(complexCol.Build())
	// age INTEGER NOT NULL DEFAULT 18 CHECK (age >= 18)
}

// TableOptionsExample demonstrates how to use NewTableDDL with options
func TableOptionsExample() {
	// 1. Простая таблица
	simpleTable := ddl.NewTableDDL(
		UsersTab,
		ddl.WithIfNotExists(),
		ddl.WithColumn(
			ddl.NewColumnDDL(ColID, ddl.Serial(), ddl.WithPrimaryKey[Col]()),
		),
		ddl.WithColumn(
			ddl.NewColumnDDL(ColName, ddl.Text(), ddl.WithNotNull[Col]()),
		),
	)
	sql, _ := simpleTable.Build()
	fmt.Println(sql)
	// CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);

	// 2. Таблица с несколькими колонками
	usersTable := ddl.NewTableDDL(
		UsersTab,
		ddl.WithIfNotExists(),
		ddl.WithColumns(
			ddl.NewColumnDDL(ColID, ddl.Serial(), ddl.WithPrimaryKey[Col]()),
			ddl.NewColumnDDL(ColName, ddl.Text(), ddl.WithNotNull[Col]()),
			ddl.NewColumnDDL(ColEmail, ddl.Text(), ddl.WithNotNull[Col](), ddl.WithUnique[Col]()),
			ddl.NewColumnDDL(ColAge, ddl.Integer()),
			ddl.NewColumnDDL(ColCreatedAt, ddl.Timestamp(), ddl.WithDefault[Col]("CURRENT_TIMESTAMP")),
		),
	)
	sql, _ = usersTable.Build()
	fmt.Println(sql)

	// 3. Таблица с table-level constraints
	postsTable := ddl.NewTableDDL(
		PostsTab,
		ddl.WithIfNotExists(),
		ddl.WithColumns(
			ddl.NewColumnDDL(ColID, ddl.Serial()),
			ddl.NewColumnDDL(ColUserID, ddl.Integer(), ddl.WithNotNull[Col]()),
			ddl.NewColumnDDL(ColTitle, ddl.Text(), ddl.WithNotNull[Col]()),
			ddl.NewColumnDDL(ColContent, ddl.Text()),
			ddl.NewColumnDDL(ColCreatedAt, ddl.Timestamp(), ddl.WithDefault[Col]("CURRENT_TIMESTAMP")),
		),
		ddl.WithTablePrimaryKey(ColID),
		ddl.WithForeignKey([]Col{ColUserID}, "users", []string{"id"}),
		ddl.WithTableCheck("length(title) > 0"),
	)
	sql, _ = postsTable.Build()
	fmt.Println(sql)
	// CREATE TABLE IF NOT EXISTS posts (
	//   id SERIAL,
	//   user_id INTEGER NOT NULL,
	//   title TEXT NOT NULL,
	//   content TEXT,
	//   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	//   PRIMARY KEY (id),
	//   FOREIGN KEY (user_id) REFERENCES users(id),
	//   CHECK (length(title) > 0)
	// );

	// 4. Таблица с составным Primary Key
	compositeTable := ddl.NewTableDDL(
		Tab("user_roles"),
		ddl.WithIfNotExists(),
		ddl.WithColumns(
			ddl.NewColumnDDL(ColUserID, ddl.Integer(), ddl.WithNotNull[Col]()),
			ddl.NewColumnDDL(Col("role_id"), ddl.Integer(), ddl.WithNotNull[Col]()),
		),
		ddl.WithTablePrimaryKey(ColUserID, Col("role_id")),
	)
	sql, _ = compositeTable.Build()
	fmt.Println(sql)
	// CREATE TABLE IF NOT EXISTS user_roles (
	//   user_id INTEGER NOT NULL,
	//   role_id INTEGER NOT NULL,
	//   PRIMARY KEY (user_id, role_id)
	// );
}

// IndexOptionsExample demonstrates how to use NewIndex with options
func IndexOptionsExample() {
	// 1. Простой индекс
	simpleIdx := ddl.NewIndex(
		Tab("idx_users_email"),
		UsersTab,
		ddl.WithIndexColumns(ColEmail),
	)
	sql, _ := simpleIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

	// 2. Уникальный индекс
	uniqueIdx := ddl.NewIndex(
		Tab("idx_users_email_unique"),
		UsersTab,
		ddl.WithIndexColumns(ColEmail),
		ddl.WithIndexUnique(),
	)
	sql, _ = uniqueIdx.Build()
	fmt.Println(sql)
	// CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (email);

	// 3. Составной индекс
	compositeIdx := ddl.NewIndex(
		Tab("idx_users_name_email"),
		UsersTab,
		ddl.WithIndexColumns(ColName, ColEmail),
	)
	sql, _ = compositeIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_users_name_email ON users (name, email);

	// 4. GIN индекс для массивов
	ginIdx := ddl.NewIndex(
		Tab("idx_posts_tags"),
		PostsTab,
		ddl.WithIndexColumns(ColTags),
		ddl.WithIndexMethod("gin"),
	)
	sql, _ = ginIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);

	// 5. Частичный индекс (partial index)
	partialIdx := ddl.NewIndex(
		Tab("idx_users_active"),
		UsersTab,
		ddl.WithIndexColumns(ColEmail),
		ddl.WithIndexPredicate("deleted_at IS NULL"),
	)
	sql, _ = partialIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX IF NOT EXISTS idx_users_active ON users (email) WHERE deleted_at IS NULL;

	// 6. Concurrent индекс
	concurrentIdx := ddl.NewIndex(
		Tab("idx_users_name"),
		UsersTab,
		ddl.WithIndexColumns(ColName),
		ddl.WithIndexConcurrently(),
	)
	sql, _ = concurrentIdx.Build()
	fmt.Println(sql)
	// CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_name ON users (name);

	// 7. Комбинация всех опций
	complexIdx := ddl.NewIndex(
		Tab("idx_users_email_active_unique"),
		UsersTab,
		ddl.WithIndexColumns(ColEmail),
		ddl.WithIndexUnique(),
		ddl.WithIndexMethod("btree"),
		ddl.WithIndexPredicate("deleted_at IS NULL"),
		ddl.WithIndexConcurrently(),
	)
	sql, _ = complexIdx.Build()
	fmt.Println(sql)
	// CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_active_unique
	// ON users USING btree (email) WHERE deleted_at IS NULL;
}

// CompleteSchemaExample shows a complete schema definition using options
func CompleteSchemaExample() {
	// Создаем таблицу users
	users := ddl.NewTableDDL(
		UsersTab,
		ddl.WithIfNotExists(),
		ddl.WithColumns(
			ddl.NewColumnDDL(
				ColID,
				ddl.Serial(),
				ddl.WithPrimaryKey[Col](),
			),
			ddl.NewColumnDDL(
				ColName,
				ddl.Varchar(255),
				ddl.WithNotNull[Col](),
			),
			ddl.NewColumnDDL(
				ColEmail,
				ddl.Varchar(255),
				ddl.WithNotNull[Col](),
			),
			ddl.NewColumnDDL(
				ColAge,
				ddl.Integer(),
				ddl.WithCheck[Col]("age >= 0"),
			),
			ddl.NewColumnDDL(
				ColCreatedAt,
				ddl.Timestamp(),
				ddl.WithNotNull[Col](),
				ddl.WithDefault[Col]("CURRENT_TIMESTAMP"),
			),
			ddl.NewColumnDDL(
				ColUpdatedAt,
				ddl.Timestamp(),
				ddl.WithNotNull[Col](),
				ddl.WithDefault[Col]("CURRENT_TIMESTAMP"),
			),
		),
		ddl.WithTableUnique(ColEmail),
	)

	// Создаем индексы
	emailIdx := ddl.NewIndex(
		Tab("idx_users_email"),
		UsersTab,
		ddl.WithIndexColumns(ColEmail),
	)

	nameIdx := ddl.NewIndex(
		Tab("idx_users_name"),
		UsersTab,
		ddl.WithIndexColumns(ColName),
	)

	// Генерируем SQL
	usersSql, _ := users.Build()
	emailIdxSql, _ := emailIdx.Build()
	nameIdxSql, _ := nameIdx.Build()

	fmt.Println(usersSql)
	fmt.Println(emailIdxSql)
	fmt.Println(nameIdxSql)
}
