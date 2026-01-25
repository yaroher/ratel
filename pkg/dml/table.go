package dml

import (
	"github.com/yaroher/ratel/pkg/dml/clause"
	"github.com/yaroher/ratel/pkg/types"
)

type TableDML[T types.TableAlias, C types.ColumnAlias] struct {
	alias     T
	allFields []C
}

func NewTableDML[T types.TableAlias, C types.ColumnAlias](alias T, fields ...C) *TableDML[T, C] {
	return &TableDML[T, C]{
		alias:     alias,
		allFields: fields,
	}
}

func (t *TableDML[T, C]) baseQuery(ta T, field ...C) BaseQuery[T, C] {
	return BaseQuery[T, C]{
		Ta:          ta,
		UsingFields: field,
		AllFields:   t.allFields,
	}
}

func (t *TableDML[T, C]) Select(field ...C) *SelectQuery[T, C] {
	return &SelectQuery[T, C]{
		BaseQuery: t.baseQuery(t.alias, field...),
	}
}
func (t *TableDML[T, C]) Select1() *SelectQuery[T, C] {
	return t.Select()
}
func (t *TableDML[T, C]) SelectAll() *SelectQuery[T, C] {
	return t.Select(t.allFields...)
}
func (t *TableDML[T, C]) Update() *UpdateQuery[T, C] {
	return &UpdateQuery[T, C]{
		BaseQuery: t.baseQuery(t.alias),
	}
}
func (t *TableDML[T, C]) Delete() *DeleteQuery[T, C] {
	return &DeleteQuery[T, C]{
		BaseQuery: t.baseQuery(t.alias),
	}
}
func (t *TableDML[T, C]) Insert() *InsertQuery[T, C] {
	return &InsertQuery[T, C]{
		BaseQuery: t.baseQuery(t.alias),
	}
}
func (t *TableDML[T, C]) Raw(sql string, args ...any) clause.Clause[C] {
	return &clause.RawExprClause[C]{SQL: sql, Args: args}
}
func (t *TableDML[T, C]) ExistsRaw(sql string, args ...any) clause.Clause[C] {
	return &clause.ExistsClause[C]{SubQuery: &clause.RawExprClause[C]{SQL: sql, Args: args}, Negate: false}
}
func (t *TableDML[T, C]) NotExistsRaw(sql string, args ...any) clause.Clause[C] {
	return &clause.ExistsClause[C]{SubQuery: &clause.RawExprClause[C]{SQL: sql, Args: args}, Negate: true}
}
func (t *TableDML[T, C]) ExistsOf(query types.Query) clause.Clause[C] {
	return &clause.ExistsClause[C]{SubQuery: &clause.SubQueryExprClause[C]{Query: query}, Negate: false}
}
func (t *TableDML[T, C]) NotExistsOf(query types.Query) clause.Clause[C] {
	return &clause.ExistsClause[C]{SubQuery: &clause.SubQueryExprClause[C]{Query: query}, Negate: true}
}
func (t *TableDML[T, C]) And(clauses ...clause.Clause[C]) clause.Clause[C] {
	return &clause.AndClause[C]{Clauses: clauses}
}
func (t *TableDML[T, C]) Or(clauses ...clause.Clause[C]) clause.Clause[C] {
	return &clause.OrClause[C]{Clauses: clauses}
}
