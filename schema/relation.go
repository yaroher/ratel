package schema

import (
	"context"
	"fmt"

	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/dml"
	"github.com/yaroher/ratel/dml/clause"
	"github.com/yaroher/ratel/exec"
)

type RelationTableAlias[T types.TableAlias] interface {
	Alias() T
}

type RelationTableJoin[T types.TableAlias, C types.ColumnAlias] interface {
	RelationTableAlias[T]
	Raw(sql string, args ...any) clause.Clause[C]
}

type RelationTableQuery[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C]] interface {
	RelationTableJoin[T, C]
	SelectAll() *dml.SelectQuery[T, C]
	Query(ctx context.Context, db exec.DB, query types.Scannable) ([]S, error)
	QueryRow(ctx context.Context, db exec.DB, query types.Scannable) (S, error)
}

// ForwardRelation represents a "has many" or "has one" relationship
// From current table to related table (e.g., User -> Posts)
type ForwardRelation[
	T types.TableAlias, // Current table alias
	C types.ColumnAlias, // Current table column alias
	S exec.Scanner[C], // Current table scanner
	RT types.TableAlias, // Related table alias
	RC types.ColumnAlias, // Related table column alias
	RS exec.Scanner[RC], // Related table scanner
] struct {
	foreignKeyName string // Foreign key in related table pointing to current table (as string)
	localKeyName   string // Primary key in current table (as string)
	relatedName    RT     // Name of the related table
}

// BackwardRelation represents a "belongs to" relationship
// From current table to parent table (e.g., Post -> User)
type BackwardRelation[
	T types.TableAlias, // Current table alias
	C types.ColumnAlias, // Current table column alias
	S exec.Scanner[C], // Current table scanner
	RT types.TableAlias, // Related table alias
	RC types.ColumnAlias, // Related table column alias
	RS exec.Scanner[RC], // Related table scanner
] struct {
	foreignKeyName string // Foreign key in current table (as string)
	ownerKeyName   string // Primary key in related table (as string)
	relatedName    RT     // Name of the related table
}

// HasMany creates a forward relation (one-to-many)
// Example: User has many Posts
func HasMany[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C], RT types.TableAlias, RC types.ColumnAlias, RS exec.Scanner[RC]](
	currentTableAlias T,
	relatedTable RelationTableAlias[RT],
	foreignKey RC, // foreign key in related table
	localKey C, // primary key in current table
) *ForwardRelation[T, C, S, RT, RC, RS] {
	return &ForwardRelation[T, C, S, RT, RC, RS]{
		foreignKeyName: foreignKey.String(),
		localKeyName:   localKey.String(),
		relatedName:    relatedTable.Alias(),
	}
}

// BelongsTo creates a backward relation (many-to-one)
// Example: Post belongs to User
func BelongsTo[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C], RT types.TableAlias, RC types.ColumnAlias, RS exec.Scanner[RC]](
	currentTableAlias T,
	relatedTable RelationTableAlias[RT],
	foreignKey C, // foreign key in current table
	ownerKey RC, // primary key in related table
) *BackwardRelation[T, C, S, RT, RC, RS] {
	return &BackwardRelation[T, C, S, RT, RC, RS]{
		foreignKeyName: foreignKey.String(),
		ownerKeyName:   ownerKey.String(),
		relatedName:    relatedTable.Alias(),
	}
}

// LoadMany loads related records for a forward relation (HasMany)
func (r *ForwardRelation[T, C, S, RT, RC, RS]) LoadMany(
	ctx context.Context,
	db exec.DB,
	relatedTable RelationTableQuery[RT, RC, RS],
	localValue any, // value of the local key (e.g., user.ID)
) ([]RS, error) {
	// Build query using raw SQL for WHERE clause
	rawWhere := relatedTable.Raw(
		fmt.Sprintf("%s.%s = $1", r.relatedName.String(), r.foreignKeyName),
		localValue,
	)

	// Build query: SELECT * FROM related_table WHERE foreign_key = localValue
	query := relatedTable.SelectAll().Where(rawWhere)

	// Execute query
	results, err := relatedTable.Query(ctx, db, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load related records: %w", err)
	}

	return results, nil
}

// LoadOne loads a related record for a backward relation (BelongsTo)
func (r *BackwardRelation[T, C, S, RT, RC, RS]) LoadOne(
	ctx context.Context,
	db exec.DB,
	relatedTable RelationTableQuery[RT, RC, RS],
	foreignValue any, // value of the foreign key (e.g., post.UserID)
) (RS, error) {
	// Build query using raw SQL for WHERE clause
	rawWhere := relatedTable.Raw(
		fmt.Sprintf("%s.%s = $1", r.relatedName.String(), r.ownerKeyName),
		foreignValue,
	)

	// Build query: SELECT * FROM related_table WHERE owner_key = foreignValue
	query := relatedTable.SelectAll().Where(rawWhere)

	// Execute query
	result, err := relatedTable.QueryRow(ctx, db, query)
	if err != nil {
		return result, fmt.Errorf("failed to load related record: %w", err)
	}

	return result, nil
}

// WithJoin adds a JOIN clause to a SelectQuery for forward relation
func (r *ForwardRelation[T, C, S, RT, RC, RS]) WithJoin(
	currentTable RelationTableJoin[T, C],
	query *dml.SelectQuery[T, C],
	joinType dml.JoinType,
) *dml.SelectQuery[T, C] {
	// Create ON clause using raw SQL: current_table.localKey = related_table.foreignKey
	currentTableName := currentTable.Alias().String()
	leftExpr := fmt.Sprintf("%s.%s", currentTableName, r.localKeyName)
	rightExpr := fmt.Sprintf("%s.%s", r.relatedName.String(), r.foreignKeyName)

	onClause := currentTable.Raw(
		fmt.Sprintf("%s = %s", leftExpr, rightExpr),
	)

	return query.Join(joinType, r.relatedName.String(), r.relatedName.String(), onClause)
}

// WithJoin adds a JOIN clause to a SelectQuery for backward relation
func (r *BackwardRelation[T, C, S, RT, RC, RS]) WithJoin(
	currentTable RelationTableJoin[T, C],
	query *dml.SelectQuery[T, C],
	joinType dml.JoinType,
) *dml.SelectQuery[T, C] {
	// Create ON clause using raw SQL: current_table.foreignKey = related_table.ownerKey
	currentTableName := currentTable.Alias().String()
	leftExpr := fmt.Sprintf("%s.%s", currentTableName, r.foreignKeyName)
	rightExpr := fmt.Sprintf("%s.%s", r.relatedName.String(), r.ownerKeyName)

	onClause := currentTable.Raw(
		fmt.Sprintf("%s = %s", leftExpr, rightExpr),
	)

	return query.Join(joinType, r.relatedName.String(), r.relatedName.String(), onClause)
}

type HasManyLoader[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec.Scanner[RC],
] struct {
	relation *ForwardRelation[T, C, S, RT, RC, RS]
	related  RelationTableQuery[RT, RC, RS]
	localKey C
	assign   func(S, []RS)
}

func (l HasManyLoader[T, C, S, RT, RC, RS]) Load(ctx context.Context, db exec.DB, base S) error {
	localValue := base.GetValue(l.localKey)()
	rows, err := l.relation.LoadMany(exec.WithSkipRelations(ctx), db, l.related, localValue)
	if err != nil {
		return err
	}
	if l.assign != nil {
		l.assign(base, rows)
	}
	return nil
}

func HasManyLoad[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec.Scanner[RC],
](
	relation *ForwardRelation[T, C, S, RT, RC, RS],
	related RelationTableQuery[RT, RC, RS],
	localKey C,
	assign func(S, []RS),
) exec.RelationLoader[S] {
	return HasManyLoader[T, C, S, RT, RC, RS]{
		relation: relation,
		related:  related,
		localKey: localKey,
		assign:   assign,
	}
}

type BelongsToLoader[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec.Scanner[RC],
] struct {
	relation   *BackwardRelation[T, C, S, RT, RC, RS]
	related    RelationTableQuery[RT, RC, RS]
	foreignKey C
	assign     func(S, RS)
}

func (l BelongsToLoader[T, C, S, RT, RC, RS]) Load(ctx context.Context, db exec.DB, base S) error {
	foreignValue := base.GetValue(l.foreignKey)()
	row, err := l.relation.LoadOne(exec.WithSkipRelations(ctx), db, l.related, foreignValue)
	if err != nil {
		return err
	}
	if l.assign != nil {
		l.assign(base, row)
	}
	return nil
}

func BelongsToLoad[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec.Scanner[RC],
](
	relation *BackwardRelation[T, C, S, RT, RC, RS],
	related RelationTableQuery[RT, RC, RS],
	foreignKey C,
	assign func(S, RS),
) exec.RelationLoader[S] {
	return BelongsToLoader[T, C, S, RT, RC, RS]{
		relation:   relation,
		related:    related,
		foreignKey: foreignKey,
		assign:     assign,
	}
}
