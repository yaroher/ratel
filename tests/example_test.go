package tests

import (
	"time"

	"github.com/yaroher/ratel/ddl"
	"github.com/yaroher/ratel/dml/set"
	"github.com/yaroher/ratel/schema"
)

type OutboxAlias string

func (o OutboxAlias) String() string { return string(o) }

const OutboxAliasName = "outbox"

type OutboxColumnAlias string

func (o OutboxColumnAlias) String() string { return string(o) }

const (
	OutboxColumnAliasID        OutboxColumnAlias = "id"
	OutboxColumnAliasEvent     OutboxColumnAlias = "event"
	OutboxColumnAliasCreatedAt OutboxColumnAlias = "created_at"
	OutboxColumnAliasUpdatedAt OutboxColumnAlias = "updated_at"
	OutboxColumnAliasDeletedAt OutboxColumnAlias = "deleted_at"
	OutboxColumnAliasData      OutboxColumnAlias = "data"
	OutboxColumnAliasVersion   OutboxColumnAlias = "version"
)

type OutboxScanner struct {
	ID        string
	Event     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Data      []byte
	Version   int32
}

func (o OutboxScanner) GetTarget(s string) func() any {
	switch OutboxColumnAlias(s) {
	case "id":
		return func() any { return &o.ID }
	case "event":
		return func() any { return &o.Event }
	case "created_at":
		return func() any { return &o.CreatedAt }
	case "updated_at":
		return func() any { return &o.UpdatedAt }
	case "deleted_at":
		return func() any { return &o.DeletedAt }
	case "data":
		return func() any { return &o.Data }
	case "version":
		return func() any { return &o.Version }
	default:
		panic("unknown field: " + s)
	}
}

func (o OutboxScanner) GetSetter(f OutboxColumnAlias) func() set.ValueSetter[OutboxColumnAlias] {
	switch f {
	case "id":
		return func() set.ValueSetter[OutboxColumnAlias] { return set.NewSetter(f, &o.ID) }
	case "event":
		return func() set.ValueSetter[OutboxColumnAlias] { return set.NewSetter(f, &o.Event) }
	case "created_at":
		return func() set.ValueSetter[OutboxColumnAlias] { return set.NewSetter(f, &o.CreatedAt) }
	case "updated_at":
		return func() set.ValueSetter[OutboxColumnAlias] { return set.NewSetter(f, &o.UpdatedAt) }
	case "deleted_at":
		return func() set.ValueSetter[OutboxColumnAlias] { return set.NewSetter(f, &o.DeletedAt) }
	case "data":
		return func() set.ValueSetter[OutboxColumnAlias] { return set.NewSetter(f, &o.Data) }
	case "version":
		return func() set.ValueSetter[OutboxColumnAlias] { return set.NewSetter(f, &o.Version) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (o OutboxScanner) GetValue(f OutboxColumnAlias) func() any {
	switch f {
	case "id":
		return func() any { return o.ID }
	case "event":
		return func() any { return o.Event }
	case "created_at":
		return func() any { return o.CreatedAt }
	case "updated_at":
		return func() any { return o.UpdatedAt }
	case "deleted_at":
		return func() any { return o.DeletedAt }
	case "data":
		return func() any { return o.Data }
	case "version":
		return func() any { return o.Version }
	default:
		panic("unknown field: " + string(f))
	}
}

type OutboxTable struct {
	*schema.Table[OutboxAlias, OutboxColumnAlias, *OutboxScanner]
	ID        schema.TextColumnI[OutboxColumnAlias]
	Event     schema.TextColumnI[OutboxColumnAlias]
	CreatedAt schema.TimestamptzColumnI[OutboxColumnAlias]
	UpdatedAt schema.TimestamptzColumnI[OutboxColumnAlias]
	DeletedAt schema.NullTimestamptzColumnI[OutboxColumnAlias]
	Data      schema.ByteaColumnI[OutboxColumnAlias]
	Version   schema.IntegerColumnI[OutboxColumnAlias]
}

var outbox OutboxTable = func() OutboxTable {
	idCol := schema.TextColumn(OutboxColumnAliasID, ddl.WithPrimaryKey[OutboxColumnAlias]())
	eventCol := schema.TextColumn(OutboxColumnAliasEvent)
	createdAtCol := schema.TimestamptzColumn(OutboxColumnAliasCreatedAt, ddl.WithDefault[OutboxColumnAlias]("CURRENT_TIMESTAMP"))
	updatedAtCol := schema.TimestamptzColumn(OutboxColumnAliasUpdatedAt, ddl.WithDefault[OutboxColumnAlias]("CURRENT_TIMESTAMP"))
	deletedAtCol := schema.NullTimestamptzColumn(OutboxColumnAliasDeletedAt)
	dataCol := schema.ByteaColumn(OutboxColumnAliasData)
	versionCol := schema.IntegerColumn(OutboxColumnAliasVersion)
	return OutboxTable{
		Table: schema.NewTable[OutboxAlias, OutboxColumnAlias, *OutboxScanner](
			OutboxAliasName,
			func() *OutboxScanner { return &OutboxScanner{} },
			[]*ddl.ColumnDDL[OutboxColumnAlias]{
				idCol.DDL(),
				eventCol.DDL(),
				createdAtCol.DDL(),
				updatedAtCol.DDL(),
				deletedAtCol.DDL(),
				dataCol.DDL(),
				versionCol.DDL(),
			},
		),
		ID:        idCol,
		Event:     eventCol,
		CreatedAt: createdAtCol,
		UpdatedAt: updatedAtCol,
		DeletedAt: deletedAtCol,
		Data:      dataCol,
		Version:   versionCol,
	}
}()
