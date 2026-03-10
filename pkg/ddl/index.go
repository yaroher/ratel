package ddl

import (
	"strings"

	"github.com/yaroher/ratel/pkg/types"
)

type Index[T types.TableAlias, C types.ColumnAlias] struct {
	name         string
	table        T
	unique       bool
	concurrently bool
	columns      []C
	using        string
	where        string
}

func NewIndex[T types.TableAlias, C types.ColumnAlias](name string, table T) *Index[T, C] {
	return &Index[T, C]{
		name:  name,
		table: table,
	}
}

// OnColumns sets the columns for the index.
func (i *Index[T, C]) OnColumns(columns ...C) *Index[T, C] {
	i.columns = append(i.columns, columns...)
	return i
}

// Unique sets the index as UNIQUE.
func (i *Index[T, C]) Unique() *Index[T, C] {
	i.unique = true
	return i
}

// Concurrently sets CONCURRENTLY option for index creation.
func (i *Index[T, C]) Concurrently() *Index[T, C] {
	i.concurrently = true
	return i
}

// Using sets the index method (btree, hash, gin, gist, etc.).
func (i *Index[T, C]) Using(method string) *Index[T, C] {
	i.using = method
	return i
}

// Where sets the WHERE clause for a partial index.
func (i *Index[T, C]) Where(predicate string) *Index[T, C] {
	i.where = predicate
	return i
}

// SchemaSqlFor generates the CREATE INDEX statement using the given table name.
// Use this instead of building index SQL directly, as the table name may need
// schema-qualification (e.g., "store"."users") which the Index struct does not track.
func (i *Index[T, C]) SchemaSqlFor(tableName string) string {
	var sql strings.Builder

	sql.WriteString("CREATE ")

	if i.unique {
		sql.WriteString("UNIQUE ")
	}

	sql.WriteString("INDEX ")

	if i.concurrently {
		sql.WriteString("CONCURRENTLY ")
	}

	sql.WriteString("IF NOT EXISTS ")
	sql.WriteString(i.name)
	sql.WriteString(" ON ")
	sql.WriteString(tableName)

	if i.using != "" {
		sql.WriteString(" USING ")
		sql.WriteString(i.using)
	}

	sql.WriteString(" (")
	for j, col := range i.columns {
		if j > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(col.String())
	}
	sql.WriteString(")")

	if i.where != "" {
		sql.WriteString(" WHERE ")
		sql.WriteString(i.where)
	}

	return sql.String()
}
