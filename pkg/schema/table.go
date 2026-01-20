package schema

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yaroher/ratel/pkg/clause"
	"github.com/yaroher/ratel/pkg/query"
	"github.com/yaroher/ratel/pkg/scanning"
	"github.com/yaroher/ratel/pkg/types"
)

// DB is the interface for database operations
type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type TableI[F types.ColumnAlias, T scanning.Targeter[F]] interface {
	AllFields() []F
	AllFieldsExcept(field ...F) []F
	Name() string
	NewScanner() T
	Select(field ...F) *query.SelectQuery[F]
	Select1() *query.SelectQuery[F]
	SelectAll() *query.SelectQuery[F]
	Insert() *query.InsertQuery[F]
	Update() *query.UpdateQuery[F]
	Delete() *query.DeleteQuery[F]
	Query(ctx context.Context, db DB, q types.OrmQuery) ([]T, error)
	QueryRow(ctx context.Context, db DB, q types.OrmQuery) (T, error)
	Execute(ctx context.Context, db DB, q types.OrmQuery) (int64, error)
}

type table[F types.ColumnAlias, T scanning.Targeter[F]] struct {
	alias       string
	allFields   []F
	scanFactory func() T
}

// NewTable creates a table with a scanner factory.
// This is intended for generated schema/scanner code.
func NewTable[F types.ColumnAlias, T scanning.Targeter[F]](
	alias string,
	scanFactory func() T,
	fields ...F,
) TableI[F, T] {
	return newTable[F, T](alias, scanFactory, fields...)
}

func newTable[F types.ColumnAlias, T scanning.Targeter[F]](
	alias string,
	scanFactory func() T,
	fields ...F,
) *table[F, T] {
	return &table[F, T]{
		alias:       alias,
		allFields:   fields,
		scanFactory: scanFactory,
	}
}

func (t *table[F, T]) baseQuery(ta string, field ...F) query.BaseQuery[F] {
	return query.BaseQuery[F]{
		Ta:          ta,
		UsingFields: field,
		AllFields:   t.allFields,
	}
}
func (t *table[F, T]) Name() string {
	return t.alias
}
func (t *table[F, T]) AllFields() []F {
	return t.allFields
}
func (t *table[F, T]) AllFieldsExcept(field ...F) []F {
	ret := make([]F, 0)
	for _, f := range t.allFields {
		needed := true
		for _, skip := range field {
			if skip.String() == f.String() {
				needed = false
				break
			}
		}
		if needed {
			ret = append(ret, f)
		}
	}
	return ret
}
func (t *table[F, T]) NewScanner() T {
	return t.scanFactory()
}
func (t *table[F, T]) Select(field ...F) *query.SelectQuery[F] {
	return &query.SelectQuery[F]{
		BaseQuery: t.baseQuery(t.alias, field...),
	}
}
func (t *table[F, T]) Select1() *query.SelectQuery[F] {
	return t.Select()
}
func (t *table[F, T]) SelectAll() *query.SelectQuery[F] {
	return t.Select(t.allFields...)
}
func (t *table[F, T]) Update() *query.UpdateQuery[F] {
	return &query.UpdateQuery[F]{
		BaseQuery: t.baseQuery(t.alias),
	}
}
func (t *table[F, T]) Delete() *query.DeleteQuery[F] {
	return &query.DeleteQuery[F]{
		BaseQuery: t.baseQuery(t.alias),
	}
}
func (t *table[F, T]) Insert() *query.InsertQuery[F] {
	return &query.InsertQuery[F]{
		BaseQuery: t.baseQuery(t.alias),
	}
}

func (t *table[F, T]) QueryRow(ctx context.Context, db DB, query types.OrmQuery) (trg T, err error) {
	trg = t.scanFactory()
	ScanAbleFields := query.ScanAbleFields()
	sql, args := query.Build()
	var targets []any
	if resolver, ok := any(trg).(scanning.TargetResolver); ok {
		targets, err = resolver.Targets(ScanAbleFields)
		if err != nil {
			return trg, err
		}
	} else {
		targets = make([]any, 0, len(ScanAbleFields))
		for _, f := range ScanAbleFields {
			targets = append(targets, trg.GetTarget(f)())
		}
	}
	err = db.QueryRow(ctx, sql, args...).Scan(targets...)
	if err != nil {
		return trg, err
	}
	return trg, nil
}

func (t *table[F, T]) Query(ctx context.Context, db DB, query types.OrmQuery) (trgs []T, err error) {
	ScanAbleFields := query.ScanAbleFields()
	sql, args := query.Build()
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		trg := t.scanFactory()
		var targets []any
		if resolver, ok := any(trg).(scanning.TargetResolver); ok {
			targets, err = resolver.Targets(ScanAbleFields)
			if err != nil {
				return nil, err
			}
		} else {
			targets = make([]any, len(ScanAbleFields))
			for i, f := range ScanAbleFields {
				targets[i] = trg.GetTarget(f)()
			}
		}
		err = rows.Scan(targets...)
		if err != nil {
			return nil, err
		}
		trgs = append(trgs, trg)
	}
	return trgs, nil
}

func (t *table[F, T]) Execute(ctx context.Context, db DB, query types.OrmQuery) (int64, error) {
	sql, args := query.Build()
	tag, err := db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
func (t *table[F, T]) Raw(sql string, args ...any) types.Clause[F] {
	return &clause.RawExprClause[F]{SQL: sql, Args: args}
}
func (t *table[F, T]) ExistsRaw(sql string, args ...any) types.Clause[F] {
	return &clause.ExistsClause[F]{SubQuery: &clause.RawExprClause[F]{SQL: sql, Args: args}, Negate: false}
}
func (t *table[F, T]) NotExistsRaw(sql string, args ...any) types.Clause[F] {
	return &clause.ExistsClause[F]{SubQuery: &clause.RawExprClause[F]{SQL: sql, Args: args}, Negate: true}
}
func (t *table[F, T]) ExistsOf(query types.OrmQuery) types.Clause[F] {
	return &clause.ExistsClause[F]{SubQuery: &clause.SubQueryExprClause[F]{Query: query}, Negate: false}
}
func (t *table[F, T]) NotExistsOf(query types.OrmQuery) types.Clause[F] {
	return &clause.ExistsClause[F]{SubQuery: &clause.SubQueryExprClause[F]{Query: query}, Negate: true}
}
func (t *table[F, T]) And(clauses ...types.Clause[F]) types.Clause[F] {
	return &clause.AndClause[F]{Clauses: clauses}
}
func (t *table[F, T]) Or(clauses ...types.Clause[F]) types.Clause[F] {
	return &clause.OrClause[F]{Clauses: clauses}
}

type copyIterator[T scanning.Valuer] struct {
	rows                 []T
	skippedFirstNextCall bool
}

func newCopyIterator[T scanning.Valuer](rows []T) *copyIterator[T] {
	return &copyIterator[T]{
		rows: rows,
	}
}
func (r *copyIterator[T]) Err() error { return nil }
func (r *copyIterator[T]) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}
func (r *copyIterator[F]) Values() ([]interface{}, error) {
	return r.rows[0].Values(), nil
}

func (t *table[F, T]) CopyFrom(ctx context.Context, db DB, values []T, fields ...F) (int64, error) {
	if len(fields) == 0 {
		return 0, errors.New("pgx-orm: fields is empty")
	}
	if len(fields) > 0 {
		fields = t.allFields
	}
	fieldsStrings := make([]string, len(fields))
	for i, f := range fields {
		fieldsStrings[i] = f.String()
	}
	return db.CopyFrom(ctx, pgx.Identifier{t.alias}, fieldsStrings, newCopyIterator[T](values))
}
