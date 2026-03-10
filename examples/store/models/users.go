package models

import (
	"time"

	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/schema"
)

// UsersAlias is the table alias type for the users table
type UsersAlias string

func (u UsersAlias) String() string { return string(u) }

const UsersAliasName UsersAlias = "users"

// UsersColumnAlias represents column names for the users table
type UsersColumnAlias string

func (u UsersColumnAlias) String() string { return string(u) }

const (
	UsersColumnUserID    UsersColumnAlias = "user_id"
	UsersColumnEmail     UsersColumnAlias = "email"
	UsersColumnFullName  UsersColumnAlias = "full_name"
	UsersColumnIsActive  UsersColumnAlias = "is_active"
	UsersColumnCreatedAt UsersColumnAlias = "created_at"
	UsersColumnUpdatedAt UsersColumnAlias = "updated_at"
)

// UsersScanner is the scanner struct for users rows
type UsersScanner struct {
	UserID    int64
	Email     string
	FullName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Orders    []*OrdersScanner
}

func (u *UsersScanner) GetTarget(s string) func() any {
	switch UsersColumnAlias(s) {
	case UsersColumnUserID:
		return func() any { return &u.UserID }
	case UsersColumnEmail:
		return func() any { return &u.Email }
	case UsersColumnFullName:
		return func() any { return &u.FullName }
	case UsersColumnIsActive:
		return func() any { return &u.IsActive }
	case UsersColumnCreatedAt:
		return func() any { return &u.CreatedAt }
	case UsersColumnUpdatedAt:
		return func() any { return &u.UpdatedAt }
	default:
		panic("unknown field: " + s)
	}
}

func (u *UsersScanner) GetSetter(f UsersColumnAlias) func() set.ValueSetter[UsersColumnAlias] {
	switch f {
	case UsersColumnUserID:
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.UserID) }
	case UsersColumnEmail:
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.Email) }
	case UsersColumnFullName:
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.FullName) }
	case UsersColumnIsActive:
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.IsActive) }
	case UsersColumnCreatedAt:
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.CreatedAt) }
	case UsersColumnUpdatedAt:
		return func() set.ValueSetter[UsersColumnAlias] { return set.NewSetter(f, &u.UpdatedAt) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (u *UsersScanner) GetValue(f UsersColumnAlias) func() any {
	switch f {
	case UsersColumnUserID:
		return func() any { return u.UserID }
	case UsersColumnEmail:
		return func() any { return u.Email }
	case UsersColumnFullName:
		return func() any { return u.FullName }
	case UsersColumnIsActive:
		return func() any { return u.IsActive }
	case UsersColumnCreatedAt:
		return func() any { return u.CreatedAt }
	case UsersColumnUpdatedAt:
		return func() any { return u.UpdatedAt }
	default:
		panic("unknown field: " + string(f))
	}
}

// Relations returns the relation loaders for the users table
func (u *UsersScanner) Relations() []exec.RelationLoader[*UsersScanner] {
	return []exec.RelationLoader[*UsersScanner]{
		schema.HasManyLoad(
			UsersOrders,
			Orders.Table,
			UsersColumnUserID,
			func(user *UsersScanner, rows []*OrdersScanner) {
				for _, row := range rows {
					row.User = user
				}
				user.Orders = rows
			},
		),
	}
}

// UsersTable represents the users table with its columns
type UsersTable struct {
	*schema.Table[UsersAlias, UsersColumnAlias, *UsersScanner]
	UserID    schema.BigSerialColumnI[UsersColumnAlias]
	Email     schema.TextColumnI[UsersColumnAlias]
	FullName  schema.TextColumnI[UsersColumnAlias]
	IsActive  schema.BooleanColumnI[UsersColumnAlias]
	CreatedAt schema.TimestamptzColumnI[UsersColumnAlias]
	UpdatedAt schema.TimestamptzColumnI[UsersColumnAlias]
}

// Users is the global users table instance
var Users = func() UsersTable {
	userIDCol := schema.BigSerialColumn(UsersColumnUserID, ddl.WithPrimaryKey[UsersColumnAlias]())
	emailCol := schema.TextColumn(
		UsersColumnEmail,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithUnique[UsersColumnAlias](),
	)
	fullNameCol := schema.TextColumn(UsersColumnFullName, ddl.WithNotNull[UsersColumnAlias]())
	isActiveCol := schema.BooleanColumn(
		UsersColumnIsActive,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithDefault[UsersColumnAlias]("true"),
	)
	createdAtCol := schema.TimestamptzColumn(
		UsersColumnCreatedAt,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithDefault[UsersColumnAlias]("now()"),
	)
	updatedAtCol := schema.TimestamptzColumn(
		UsersColumnUpdatedAt,
		ddl.WithNotNull[UsersColumnAlias](),
		ddl.WithDefault[UsersColumnAlias]("now()"),
	)

	return UsersTable{
		Table: schema.NewTable[UsersAlias, UsersColumnAlias, *UsersScanner](
			UsersAliasName,
			func() *UsersScanner { return &UsersScanner{} },
			[]*ddl.ColumnDDL[UsersColumnAlias]{
				userIDCol.DDL(),
				emailCol.DDL(),
				fullNameCol.DDL(),
				isActiveCol.DDL(),
				createdAtCol.DDL(),
				updatedAtCol.DDL(),
			},
			ddl.WithPostStatements[UsersAlias, UsersColumnAlias](
				"ALTER TABLE {table} ENABLE ROW LEVEL SECURITY",
				"CREATE POLICY users_own_data ON {table} FOR ALL USING (user_id = current_setting('app.current_user_id')::bigint)",
				"CREATE POLICY users_insert ON {table} FOR INSERT WITH CHECK (true)",
			),
		),
		UserID:    userIDCol,
		Email:     emailCol,
		FullName:  fullNameCol,
		IsActive:  isActiveCol,
		CreatedAt: createdAtCol,
		UpdatedAt: updatedAtCol,
	}
}()

// UsersRef is a reference to the users table for relations
var UsersRef schema.RelationTableAlias[UsersAlias] = Users.Table
