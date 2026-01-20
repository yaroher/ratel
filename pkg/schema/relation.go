package schema

import (
	"github.com/yaroher/ratel/pkg/clause"
	"github.com/yaroher/ratel/pkg/query"
	"github.com/yaroher/ratel/pkg/types"
)

// NewRawOnClause создает raw SQL условие для ON clause
func NewRawOnClause[C types.ColumnAlias](sql string, args ...any) types.Clause[C] {
	return &clause.RawExprClause[C]{SQL: sql, Args: args}
}

type RelationType string

const (
	HasOneRelation     RelationType = "has_one"
	HasManyRelation    RelationType = "has_many"
	BelongsToRelation  RelationType = "belongs_to"
	ManyToManyRelation RelationType = "many_to_many"
)

// Relation описывает связь между таблицами
type Relation[C types.ColumnAlias] struct {
	relationType   RelationType
	fromTable      string
	toTable        string
	toTableAlias   string
	foreignKey     string // колонка для связи (в toTable для HasOne/HasMany, в fromTable для BelongsTo)
	localKey       string // колонка в fromTable (для HasOne/HasMany) или toTable (для BelongsTo)
	throughTable   string // для many-to-many
	throughAlias   string // для many-to-many
	throughFromKey string // для many-to-many
	throughToKey   string // для many-to-many
}

// HasOne создает связь один-к-одному
// fromTable: таблица, из которой идет связь
// toTable: таблица, к которой идет связь
// foreignKey: имя колонки в toTable, которая ссылается на localKey в fromTable
// localKey: имя колонки в fromTable (обычно primary key)
func HasOne[C types.ColumnAlias](fromTable, toTable, foreignKey, localKey string) *Relation[C] {
	return &Relation[C]{
		relationType: HasOneRelation,
		fromTable:    fromTable,
		toTable:      toTable,
		toTableAlias: toTable,
		foreignKey:   foreignKey,
		localKey:     localKey,
	}
}

// HasMany создает связь один-ко-многим
// fromTable: таблица, из которой идет связь
// toTable: таблица, к которой идет связь
// foreignKey: имя колонки в toTable, которая ссылается на localKey в fromTable
// localKey: имя колонки в fromTable (обычно primary key)
func HasMany[C types.ColumnAlias](fromTable, toTable, foreignKey, localKey string) *Relation[C] {
	return &Relation[C]{
		relationType: HasManyRelation,
		fromTable:    fromTable,
		toTable:      toTable,
		toTableAlias: toTable,
		foreignKey:   foreignKey,
		localKey:     localKey,
	}
}

// BelongsTo создает обратную связь (принадлежит к)
// fromTable: таблица, из которой идет связь
// toTable: таблица, к которой идет связь
// foreignKey: имя колонки в fromTable, которая ссылается на localKey в toTable
// localKey: имя колонки в toTable (обычно primary key)
func BelongsTo[C types.ColumnAlias](fromTable, toTable, foreignKey, localKey string) *Relation[C] {
	return &Relation[C]{
		relationType: BelongsToRelation,
		fromTable:    fromTable,
		toTable:      toTable,
		toTableAlias: toTable,
		foreignKey:   foreignKey,
		localKey:     localKey,
	}
}

// ManyToMany создает связь многие-ко-многим через промежуточную таблицу
// fromTable: таблица, из которой идет связь
// toTable: таблица, к которой идет связь
// throughTable: промежуточная таблица
// throughFromKey: имя колонки в throughTable, которая ссылается на fromTable
// throughToKey: имя колонки в throughTable, которая ссылается на toTable
func ManyToMany[C types.ColumnAlias](fromTable, toTable, throughTable, throughFromKey, throughToKey string) *Relation[C] {
	return &Relation[C]{
		relationType:   ManyToManyRelation,
		fromTable:      fromTable,
		toTable:        toTable,
		toTableAlias:   toTable,
		throughTable:   throughTable,
		throughAlias:   throughTable,
		throughFromKey: throughFromKey,
		throughToKey:   throughToKey,
	}
}

// WithAlias устанавливает алиас для связанной таблицы
func (r *Relation[C]) WithAlias(alias string) *Relation[C] {
	r.toTableAlias = alias
	return r
}

// WithThroughAlias устанавливает алиас для промежуточной таблицы (только для ManyToMany)
func (r *Relation[C]) WithThroughAlias(alias string) *Relation[C] {
	r.throughAlias = alias
	return r
}

// ApplyToQuery применяет связь к SelectQuery, добавляя соответствующие JOIN
func (r *Relation[C]) ApplyToQuery(q *query.SelectQuery[C], joinType query.JoinType) *query.SelectQuery[C] {
	switch r.relationType {
	case HasOneRelation, HasManyRelation:
		// JOIN toTable ON toTable.foreignKey = fromTable.localKey
		q.Join(joinType, r.toTable, r.toTableAlias, NewRawOnClause[C](
			r.toTableAlias+"."+r.foreignKey+" = "+r.fromTable+"."+r.localKey,
		))

	case BelongsToRelation:
		// JOIN toTable ON fromTable.foreignKey = toTable.localKey
		q.Join(joinType, r.toTable, r.toTableAlias, NewRawOnClause[C](
			r.fromTable+"."+r.foreignKey+" = "+r.toTableAlias+"."+r.localKey,
		))

	case ManyToManyRelation:
		// Сначала JOIN промежуточной таблицы
		// JOIN throughTable ON fromTable.id = throughTable.throughFromKey
		q.Join(joinType, r.throughTable, r.throughAlias, NewRawOnClause[C](
			r.fromTable+".id = "+r.throughAlias+"."+r.throughFromKey,
		))
		// Потом JOIN целевой таблицы
		// JOIN toTable ON throughTable.throughToKey = toTable.id
		q.Join(joinType, r.toTable, r.toTableAlias, NewRawOnClause[C](
			r.throughAlias+"."+r.throughToKey+" = "+r.toTableAlias+".id",
		))
	}

	return q
}

// InnerJoin применяет связь с INNER JOIN
func (r *Relation[C]) InnerJoin(q *query.SelectQuery[C]) *query.SelectQuery[C] {
	return r.ApplyToQuery(q, query.InnerJoinType)
}

// LeftJoin применяет связь с LEFT JOIN
func (r *Relation[C]) LeftJoin(q *query.SelectQuery[C]) *query.SelectQuery[C] {
	return r.ApplyToQuery(q, query.LeftJoinType)
}

// RightJoin применяет связь с RIGHT JOIN
func (r *Relation[C]) RightJoin(q *query.SelectQuery[C]) *query.SelectQuery[C] {
	return r.ApplyToQuery(q, query.RightJoinType)
}

// FullJoin применяет связь с FULL JOIN
func (r *Relation[C]) FullJoin(q *query.SelectQuery[C]) *query.SelectQuery[C] {
	return r.ApplyToQuery(q, query.FullJoinType)
}
