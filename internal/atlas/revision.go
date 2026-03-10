package atlas

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/yaroher/ratel/internal/pgtypecast"
	"github.com/yaroher/ratel/pkg/pgx-ext/sqlexec"

	"ariga.io/atlas/sql/migrate"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	createTableSQL = `CREATE TABLE IF NOT EXISTS atlas_migrations (
		version TEXT NOT NULL PRIMARY KEY,
		description TEXT NOT NULL,
		type BIGINT NOT NULL,
		applied BIGINT NOT NULL,
		total BIGINT NOT NULL,
		executed_at TIMESTAMPTZ NOT NULL,
		execution_time BIGINT NOT NULL,
		error TEXT NOT NULL,
		error_stmt TEXT NOT NULL,
		hash TEXT NOT NULL,
		partial_hashes TEXT[] DEFAULT '{}',
		operator_version TEXT NOT NULL
	)`

	selectAllRevisionsSQL = `SELECT 
		version, description, type, applied, total, 
		executed_at, execution_time, error, error_stmt, 
		hash, partial_hashes, operator_version 
	FROM atlas_migrations`

	selectRevisionByVersionSQL = `SELECT 
		version, description, type, applied, total, 
		executed_at, execution_time, error, error_stmt, 
		hash, partial_hashes, operator_version 
	FROM atlas_migrations 
	WHERE version = $1`

	upsertRevisionSQL = `INSERT INTO atlas_migrations (
		version, description, type, applied, total, 
		executed_at, execution_time, error, error_stmt, 
		hash, partial_hashes, operator_version
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT (version) DO UPDATE SET
		description = EXCLUDED.description,
		type = EXCLUDED.type,
		applied = EXCLUDED.applied,
		total = EXCLUDED.total,
		executed_at = EXCLUDED.executed_at,
		execution_time = EXCLUDED.execution_time,
		error = EXCLUDED.error,
		error_stmt = EXCLUDED.error_stmt,
		hash = EXCLUDED.hash,
		partial_hashes = EXCLUDED.partial_hashes,
		operator_version = EXCLUDED.operator_version`

	deleteRevisionSQL = `DELETE FROM atlas_migrations WHERE version = $1`
)

func arrtoToPg(s []string) pgtype.Array[string] {
	if s == nil {
		return pgtype.Array[string]{}
	}
	return pgtype.Array[string]{Elements: s, Valid: true}
}

func intToPg(i int) pgtype.Int8 {
	return pgtype.Int8{Int64: int64(i), Valid: true}
}

type Revision struct {
	Version         pgtype.Text          `db:"version"`          // Version of the migration.
	Description     pgtype.Text          `db:"description"`      // Description of this migration.
	Type            pgtype.Int8          `db:"type"`             // Type of the migration.
	Applied         pgtype.Int8          `db:"applied"`          // Applied amount of statements in the migration.
	Total           pgtype.Int8          `db:"total"`            // Total amount of statements in the migration.
	ExecutedAt      pgtype.Timestamptz   `db:"executed_at"`      // ExecutedAt is the starting point of execution.
	ExecutionTime   pgtype.Int8          `db:"execution_time"`   // ExecutionTime of the migration.
	Error           pgtype.Text          `db:"error"`            // Error of the migration, if any occurred.
	ErrorStmt       pgtype.Text          `db:"error_stmt"`       // ErrorStmt is the statement that raised Error.
	Hash            pgtype.Text          `db:"hash"`             // Hash of migration file.
	PartialHashes   pgtype.Array[string] `db:"partial_hashes"`   // PartialHashes is the hashes of applied statements.
	OperatorVersion pgtype.Text          `db:"operator_version"` // OperatorVersion that executed this migration.
}

func (r *Revision) ToRevision() *migrate.Revision {
	return &migrate.Revision{
		Version:         pgtypecast.PgTextToStr(r.Version),
		Description:     pgtypecast.PgTextToStr(r.Description),
		Type:            migrate.RevisionType(r.Type.Int64),
		Applied:         int(r.Applied.Int64),
		Total:           int(r.Total.Int64),
		ExecutedAt:      pgtypecast.PgTimestamptzToTime(r.ExecutedAt),
		ExecutionTime:   time.Duration(r.ExecutionTime.Int64),
		Error:           pgtypecast.PgTextToStr(r.Error),
		ErrorStmt:       pgtypecast.PgTextToStr(r.ErrorStmt),
		Hash:            pgtypecast.PgTextToStr(r.Hash),
		PartialHashes:   r.PartialHashes.Elements,
		OperatorVersion: pgtypecast.PgTextToStr(r.OperatorVersion),
	}
}

func revisionToDb(r *migrate.Revision) *Revision {
	return &Revision{
		Version:         pgtypecast.StrToPgText(r.Version),
		Description:     pgtypecast.StrToPgText(r.Description),
		Type:            intToPg(int(r.Type)),
		Applied:         intToPg(r.Applied),
		Total:           intToPg(r.Total),
		ExecutedAt:      pgtypecast.TimeToPgTimestamptz(r.ExecutedAt),
		ExecutionTime:   intToPg(int(r.ExecutionTime.Milliseconds())),
		Error:           pgtypecast.StrToPgText(r.Error),
		ErrorStmt:       pgtypecast.StrToPgText(r.ErrorStmt),
		Hash:            pgtypecast.StrToPgText(r.Hash),
		PartialHashes:   arrtoToPg(r.PartialHashes),
		OperatorVersion: pgtypecast.StrToPgText(r.OperatorVersion),
	}
}

const (
	migrationsTableName = "atlas_migrations"
	publicSchema        = "public"
)

type RevisionReaderWriter struct {
	exec sqlexec.Executor
}

func NewRevisionReaderWriter(exec sqlexec.Executor) (*RevisionReaderWriter, error) {
	_, err := exec.Exec(context.Background(), createTableSQL)
	if err != nil {
		return nil, err
	}
	return &RevisionReaderWriter{
		exec: exec,
	}, nil
}

func (r *RevisionReaderWriter) Ident() *migrate.TableIdent {
	return &migrate.TableIdent{
		Name: migrationsTableName, Schema: publicSchema,
	}
}

func (r *RevisionReaderWriter) ReadRevisions(ctx context.Context) ([]*migrate.Revision, error) {
	rows, err := r.exec.Query(ctx, selectAllRevisionsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revisions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Revision])
	if err != nil {
		return nil, err
	}

	ret := make([]*migrate.Revision, 0, len(revisions))
	for _, rev := range revisions {
		ret = append(ret, rev.ToRevision())
	}
	return ret, nil
}

func (r *RevisionReaderWriter) ReadRevision(ctx context.Context, version string) (*migrate.Revision, error) {
	rows, err := r.exec.Query(ctx, selectRevisionByVersionSQL, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revision, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Revision])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, migrate.ErrRevisionNotExist
		}
		return nil, err
	}
	return revision.ToRevision(), nil
}

func (r *RevisionReaderWriter) WriteRevision(ctx context.Context, revision *migrate.Revision) error {
	rev := revisionToDb(revision)
	_, err := r.exec.Exec(ctx, upsertRevisionSQL,
		rev.Version,
		rev.Description,
		rev.Type,
		rev.Applied,
		rev.Total,
		rev.ExecutedAt,
		rev.ExecutionTime,
		rev.Error,
		rev.ErrorStmt,
		rev.Hash,
		rev.PartialHashes,
		rev.OperatorVersion,
	)
	return err
}

func (r *RevisionReaderWriter) DeleteRevision(ctx context.Context, version string) error {
	_, err := r.exec.Exec(ctx, deleteRevisionSQL, version)
	return err
}
