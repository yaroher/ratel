package exec

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yaroher/ratel/common/types"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type TableExecutor[T types.TableAlias, C types.ColumnAlias, S Scanner[C]] struct {
	alias       T
	scanFactory func() S
	allFields   []C
}

func NewTableExecutor[T types.TableAlias, C types.ColumnAlias, S Scanner[C]](
	alias T,
	allFields []C,
	scanFactory func() S,
) *TableExecutor[T, C, S] {
	return &TableExecutor[T, C, S]{
		alias:       alias,
		scanFactory: scanFactory,
		allFields:   allFields,
	}
}

func (t *TableExecutor[T, C, S]) QueryRow(ctx context.Context, db DB, query types.Scannable) (trg S, err error) {
	trg = t.scanFactory()
	scanAbleFields := query.ScanAbleFields()
	sql, args := query.Build()
	var targets []any
	targets = make([]any, 0, len(scanAbleFields))
	for _, f := range scanAbleFields {
		targets = append(targets, trg.GetTarget(f)())
	}
	err = db.QueryRow(ctx, sql, args...).Scan(targets...)
	if err != nil {
		return trg, err
	}
	return trg, nil
}

func (t *TableExecutor[T, C, S]) Query(ctx context.Context, db DB, query types.Scannable) (trgs []S, err error) {
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
		targets = make([]any, len(ScanAbleFields))
		for i, f := range ScanAbleFields {
			targets[i] = trg.GetTarget(f)()
		}
		err = rows.Scan(targets...)
		if err != nil {
			return nil, err
		}
		trgs = append(trgs, trg)
	}
	return trgs, nil
}

func (t *TableExecutor[T, C, S]) Execute(ctx context.Context, db DB, query types.Buildable) (int64, error) {
	sql, args := query.Build()
	tag, err := db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (t *TableExecutor[T, C, S]) CopyFrom(ctx context.Context, db DB, values []S, fields ...C) (int64, error) {
	if len(fields) == 0 {
		return 0, errors.New("pgx-orm: schema is empty")
	}
	if len(fields) > 0 {
		fields = t.allFields
	}
	fieldsStrings := make([]string, len(fields))
	for i, f := range fields {
		fieldsStrings[i] = f.String()
	}
	return db.CopyFrom(ctx, pgx.Identifier{t.alias.String()}, fieldsStrings, newCopyIterator[C, S](t.allFields, values))
}

type copyIterator[C types.ColumnAlias, S Scanner[C]] struct {
	allFields            []C
	rows                 []S
	skippedFirstNextCall bool
}

func newCopyIterator[C types.ColumnAlias, S Scanner[C]](
	allFields []C,
	rows []S,
) *copyIterator[C, S] {
	return &copyIterator[C, S]{
		rows: rows,
	}
}
func (r *copyIterator[C, S]) Err() error { return nil }
func (r *copyIterator[C, S]) Next() bool {
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
func (r *copyIterator[C, S]) Values() ([]interface{}, error) {
	vals := make([]interface{}, len(r.allFields))
	for i, f := range r.allFields {
		vals[i] = r.rows[0].GetTarget(f.String())()
	}
	return vals, nil
}
