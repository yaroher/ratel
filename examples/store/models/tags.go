package models

import (
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/schema"
)

// TagsAlias is the table alias type for the tags table
type TagsAlias string

func (t TagsAlias) String() string { return string(t) }

const TagsAliasName TagsAlias = "tags"

// TagsColumnAlias represents column names for the tags table
type TagsColumnAlias string

func (t TagsColumnAlias) String() string { return string(t) }

const (
	TagsColumnTagID TagsColumnAlias = "tag_id"
	TagsColumnName  TagsColumnAlias = "name"
	TagsColumnSlug  TagsColumnAlias = "slug"
)

// TagsScanner is the scanner struct for tags rows
type TagsScanner struct {
	TagID    int64
	Name     string
	Slug     string
	Products []*ProductsScanner
}

func (t *TagsScanner) GetTarget(s string) func() any {
	switch TagsColumnAlias(s) {
	case TagsColumnTagID:
		return func() any { return &t.TagID }
	case TagsColumnName:
		return func() any { return &t.Name }
	case TagsColumnSlug:
		return func() any { return &t.Slug }
	default:
		panic("unknown field: " + s)
	}
}

func (t *TagsScanner) GetSetter(f TagsColumnAlias) func() set.ValueSetter[TagsColumnAlias] {
	switch f {
	case TagsColumnTagID:
		return func() set.ValueSetter[TagsColumnAlias] { return set.NewSetter(f, &t.TagID) }
	case TagsColumnName:
		return func() set.ValueSetter[TagsColumnAlias] { return set.NewSetter(f, &t.Name) }
	case TagsColumnSlug:
		return func() set.ValueSetter[TagsColumnAlias] { return set.NewSetter(f, &t.Slug) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (t *TagsScanner) GetValue(f TagsColumnAlias) func() any {
	switch f {
	case TagsColumnTagID:
		return func() any { return t.TagID }
	case TagsColumnName:
		return func() any { return t.Name }
	case TagsColumnSlug:
		return func() any { return t.Slug }
	default:
		panic("unknown field: " + string(f))
	}
}

// TagsTable represents the tags table with its columns
type TagsTable struct {
	*schema.Table[TagsAlias, TagsColumnAlias, *TagsScanner]
	TagID schema.BigSerialColumnI[TagsColumnAlias]
	Name  schema.TextColumnI[TagsColumnAlias]
	Slug  schema.TextColumnI[TagsColumnAlias]
}

// Tags is the global tags table instance
var Tags = func() TagsTable {
	tagIDCol := schema.BigSerialColumn(TagsColumnTagID, ddl.WithPrimaryKey[TagsColumnAlias]())
	nameCol := schema.TextColumn(TagsColumnName, ddl.WithNotNull[TagsColumnAlias]())
	slugCol := schema.TextColumn(
		TagsColumnSlug,
		ddl.WithNotNull[TagsColumnAlias](),
		ddl.WithUnique[TagsColumnAlias](),
	)

	return TagsTable{
		Table: schema.NewTable[TagsAlias, TagsColumnAlias, *TagsScanner](
			TagsAliasName,
			func() *TagsScanner { return &TagsScanner{} },
			[]*ddl.ColumnDDL[TagsColumnAlias]{
				tagIDCol.DDL(),
				nameCol.DDL(),
				slugCol.DDL(),
			},
		),
		TagID: tagIDCol,
		Name:  nameCol,
		Slug:  slugCol,
	}
}()

// TagsRef is a reference to the tags table for relations
var TagsRef schema.RelationTableAlias[TagsAlias] = Tags.Table
