package sqlerr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ============================================================================
// Constraint Error Types
// ============================================================================

// ConstraintType represents the type of database constraint
type ConstraintType string

const (
	UniqueConstraint     ConstraintType = "unique"
	PrimaryKeyConstraint ConstraintType = "primary_key"
	ForeignKeyConstraint ConstraintType = "foreign_key"
	CheckConstraint      ConstraintType = "check"
	UnknownConstraint    ConstraintType = "unknown"
)

// PostgreSQL error codes for constraint violations
const (
	PgCodeUniqueViolation     = "23505" // unique_violation
	PgCodeForeignKeyViolation = "23503" // foreign_key_violation
	PgCodeCheckViolation      = "23514" // check_violation
	PgCodeNotNullViolation    = "23502" // not_null_violation
	PgCodeExclusionViolation  = "23P01" // exclusion_violation
)

// ConstraintError contains information about a violated database constraint
type ConstraintError struct {
	Name       string         // constraint name (e.g., "users_email_key")
	Table      string         // table name
	Column     string         // column name (if available)
	Type       ConstraintType // UNIQUE, FK, PK, CHECK
	Detail     string         // error details from PostgreSQL
	Message    string         // error message
	underlying error
}

func (e *ConstraintError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("%s constraint violated: %s", e.Type, e.Name)
	}
	return fmt.Sprintf("%s constraint violated on table %s", e.Type, e.Table)
}

func (e *ConstraintError) Unwrap() error {
	return e.underlying
}

// Is implements errors.Is for ConstraintError
func (e *ConstraintError) Is(target error) bool {
	if target == nil {
		return false
	}
	t, ok := target.(*ConstraintError)
	if !ok {
		return false
	}
	// Match by constraint name if both have names
	if e.Name != "" && t.Name != "" {
		return e.Name == t.Name
	}
	// Match by type and table
	return e.Type == t.Type && e.Table == t.Table
}

// AsConstraintError extracts constraint information from a PostgreSQL error.
// Returns the ConstraintError and true if the error is a constraint violation,
// otherwise returns nil and false.
func AsConstraintError(err error) (*ConstraintError, bool) {
	if err == nil {
		return nil, false
	}

	// Try to get the underlying PgError
	var pgErr *pgconn.PgError

	// First try direct conversion
	if errors.As(err, &pgErr) {
		return pgErrorToConstraintError(pgErr), true
	}

	// Try unwrapping executorError
	var execErr *executorError
	if errors.As(err, &execErr) {
		if errors.As(execErr.err, &pgErr) {
			return pgErrorToConstraintError(pgErr), true
		}
	}

	return nil, false
}

// pgErrorToConstraintError converts a PgError to ConstraintError
func pgErrorToConstraintError(pgErr *pgconn.PgError) *ConstraintError {
	ce := &ConstraintError{
		Name:       pgErr.ConstraintName,
		Table:      pgErr.TableName,
		Column:     pgErr.ColumnName,
		Detail:     pgErr.Detail,
		Message:    pgErr.Message,
		underlying: pgErr,
	}

	// Determine constraint type from error code
	switch pgErr.Code {
	case PgCodeUniqueViolation:
		ce.Type = UniqueConstraint
	case PgCodeForeignKeyViolation:
		ce.Type = ForeignKeyConstraint
	case PgCodeCheckViolation:
		ce.Type = CheckConstraint
	case PgCodeNotNullViolation:
		ce.Type = UnknownConstraint // NOT NULL is not a named constraint
	case PgCodeExclusionViolation:
		ce.Type = UnknownConstraint
	default:
		ce.Type = UnknownConstraint
	}

	// Try to determine if it's a primary key violation (special case of unique)
	if ce.Type == UniqueConstraint && strings.HasSuffix(ce.Name, "_pkey") {
		ce.Type = PrimaryKeyConstraint
	}

	return ce
}

// IsConstraintNamed checks if the error is a constraint violation with the specified name
func IsConstraintNamed(err error, constraintName string) bool {
	ce, ok := AsConstraintError(err)
	if !ok {
		return false
	}
	return ce.Name == constraintName
}

// IsConstraintType checks if the error is a constraint violation of the specified type
func IsConstraintType(err error, constraintType ConstraintType) bool {
	ce, ok := AsConstraintError(err)
	if !ok {
		return false
	}
	return ce.Type == constraintType
}

const (
	sqlBuildStageErr = "sql build stage: "
	sqlExecStageErr  = "sql exec stage: "
	sqlScanStageErr  = "sql sqlscan stage: "
)

type ErroredRow struct {
	BuilderErr error
}

func (b *ErroredRow) Scan(...any) error {
	return b.BuilderErr
}

type executorError struct {
	err   error
	stage string
}

func (e *executorError) Error() string {
	return fmt.Sprintf("%s %s", e.stage, e.err)
}

func (e *executorError) Unwrap() error {
	return e.err
}

func SqlBuildErr(err error) error {
	return &executorError{
		err:   err,
		stage: sqlBuildStageErr,
	}
}

func SqlExecErr(err error) error {
	return &executorError{
		err:   err,
		stage: sqlExecStageErr,
	}
}

func SqlScanErr(err error) error {
	return &executorError{
		err:   err,
		stage: sqlScanStageErr,
	}
}

func IsPgxErr(err, target error) bool {
	var cast *executorError
	ok := errors.As(err, &cast)
	if ok {
		return errors.Is(cast.err, target)
	}

	return errors.Is(err, target)
}

func IsNotFound(err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	return IsPgxErr(err, pgx.ErrNoRows)
}

func IsConstraintError(err error) bool {
	return IsUniqueConstraintError(err) || IsForeignKeyConstraintError(err)
}

// IsUniqueConstraintError reports if the error resulted from a DB uniqueness constraint violation.
// e.g. duplicate value in unique index.
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	for _, s := range []string{
		"violates unique constraint", // Postgres
		"UNIQUE constraint failed",   // SQLite
	} {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}
	return false
}

// IsForeignKeyConstraintError reports if the error resulted from a database foreign-key constraint violation.
// e.g. parent row does not exist.
func IsForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}
	for _, s := range []string{
		"violates foreign key constraint", // Postgres
		"FOREIGN KEY constraint failed",   // SQLite
	} {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}
	return false
}
