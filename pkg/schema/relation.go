package schema

import (
	"context"
	"fmt"

	"github.com/yaroher/ratel/pkg/dml"
	"github.com/yaroher/ratel/pkg/dml/clause"
	exec2 "github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/pgx-ext/sqlerr"
	"github.com/yaroher/ratel/pkg/types"
)

type RelationTableAlias[T types.TableAlias] interface {
	Alias() T
}

type RelationTableJoin[T types.TableAlias, C types.ColumnAlias] interface {
	RelationTableAlias[T]
	Raw(sql string, args ...any) clause.Clause[C]
}

type RelationTableQuery[T types.TableAlias, C types.ColumnAlias, S exec2.Scanner[C]] interface {
	RelationTableJoin[T, C]
	SelectAll() *dml.SelectQuery[T, C]
	Query(ctx context.Context, db exec2.DB, query types.Scannable, opts ...exec2.QueryOption[C, S]) ([]S, error)
	QueryRow(ctx context.Context, db exec2.DB, query types.Scannable, opts ...exec2.QueryOption[C, S]) (S, error)
}

// ForwardRelation represents a "has many" or "has one" relationship
// From current table to related table (e.g., User -> Posts)
type ForwardRelation[
	T types.TableAlias, // Current table alias
	C types.ColumnAlias, // Current table column alias
	S exec2.Scanner[C], // Current table scanner
	RT types.TableAlias, // Related table alias
	RC types.ColumnAlias, // Related table column alias
	RS exec2.Scanner[RC], // Related table scanner
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
	S exec2.Scanner[C], // Current table scanner
	RT types.TableAlias, // Related table alias
	RC types.ColumnAlias, // Related table column alias
	RS exec2.Scanner[RC], // Related table scanner
] struct {
	foreignKeyName string // Foreign key in current table (as string)
	ownerKeyName   string // Primary key in related table (as string)
	relatedName    RT     // Name of the related table
}

// HasMany creates a forward relation (one-to-many)
// Example: User has many Posts
func HasMany[T types.TableAlias, C types.ColumnAlias, S exec2.Scanner[C], RT types.TableAlias, RC types.ColumnAlias, RS exec2.Scanner[RC]](
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

// HasOne creates a forward relation (one-to-one)
// Example: User has one Profile
func HasOne[T types.TableAlias, C types.ColumnAlias, S exec2.Scanner[C], RT types.TableAlias, RC types.ColumnAlias, RS exec2.Scanner[RC]](
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
func BelongsTo[T types.TableAlias, C types.ColumnAlias, S exec2.Scanner[C], RT types.TableAlias, RC types.ColumnAlias, RS exec2.Scanner[RC]](
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
	db exec2.DB,
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

// LoadOne loads a single related record for a forward relation (HasOne)
func (r *ForwardRelation[T, C, S, RT, RC, RS]) LoadOne(
	ctx context.Context,
	db exec2.DB,
	relatedTable RelationTableQuery[RT, RC, RS],
	localValue any, // value of the local key (e.g., user.ID)
) (RS, error) {
	// Build query using raw SQL for WHERE clause
	rawWhere := relatedTable.Raw(
		fmt.Sprintf("%s.%s = $1", r.relatedName.String(), r.foreignKeyName),
		localValue,
	)

	// Build query: SELECT * FROM related_table WHERE foreign_key = localValue LIMIT 1
	query := relatedTable.SelectAll().Where(rawWhere)

	// Execute query - returns single row
	result, err := relatedTable.QueryRow(ctx, db, query)
	if err != nil {
		return result, fmt.Errorf("failed to load related record: %w", err)
	}

	return result, nil
}

// LoadOne loads a related record for a backward relation (BelongsTo)
func (r *BackwardRelation[T, C, S, RT, RC, RS]) LoadOne(
	ctx context.Context,
	db exec2.DB,
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
	S exec2.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
] struct {
	relation *ForwardRelation[T, C, S, RT, RC, RS]
	related  RelationTableQuery[RT, RC, RS]
	localKey C
	assign   func(S, []RS)
}

func (l HasManyLoader[T, C, S, RT, RC, RS]) Load(ctx context.Context, db exec2.DB, base S) error {
	localValue := base.GetValue(l.localKey)()
	rows, err := l.relation.LoadMany(exec2.WithSkipRelations(ctx), db, l.related, localValue)
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
	S exec2.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
](
	relation *ForwardRelation[T, C, S, RT, RC, RS],
	related RelationTableQuery[RT, RC, RS],
	localKey C,
	assign func(S, []RS),
) exec2.RelationLoader[S] {
	return HasManyLoader[T, C, S, RT, RC, RS]{
		relation: relation,
		related:  related,
		localKey: localKey,
		assign:   assign,
	}
}

// HasOneLoader implements RelationLoader for one-to-one forward relations
type HasOneLoader[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec2.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
] struct {
	relation *ForwardRelation[T, C, S, RT, RC, RS]
	related  RelationTableQuery[RT, RC, RS]
	localKey C
	assign   func(S, RS)
}

func (l HasOneLoader[T, C, S, RT, RC, RS]) Load(ctx context.Context, db exec2.DB, base S) error {
	localValue := base.GetValue(l.localKey)()
	row, err := l.relation.LoadOne(exec2.WithSkipRelations(ctx), db, l.related, localValue)
	if err != nil {
		// HasOne relation may not exist - that's OK, just leave the field nil
		if sqlerr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if l.assign != nil {
		l.assign(base, row)
	}
	return nil
}

// HasOneLoad creates a loader for one-to-one forward relations
func HasOneLoad[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec2.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
](
	relation *ForwardRelation[T, C, S, RT, RC, RS],
	related RelationTableQuery[RT, RC, RS],
	localKey C,
	assign func(S, RS),
) exec2.RelationLoader[S] {
	return HasOneLoader[T, C, S, RT, RC, RS]{
		relation: relation,
		related:  related,
		localKey: localKey,
		assign:   assign,
	}
}

type BelongsToLoader[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec2.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
] struct {
	relation   *BackwardRelation[T, C, S, RT, RC, RS]
	related    RelationTableQuery[RT, RC, RS]
	foreignKey C
	assign     func(S, RS)
}

func (l BelongsToLoader[T, C, S, RT, RC, RS]) Load(ctx context.Context, db exec2.DB, base S) error {
	foreignValue := base.GetValue(l.foreignKey)()
	row, err := l.relation.LoadOne(exec2.WithSkipRelations(ctx), db, l.related, foreignValue)
	if err != nil {
		// BelongsTo relation may not exist (e.g. NULL foreign key) - that's OK
		if sqlerr.IsNotFound(err) {
			return nil
		}
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
	S exec2.Scanner[C],
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
](
	relation *BackwardRelation[T, C, S, RT, RC, RS],
	related RelationTableQuery[RT, RC, RS],
	foreignKey C,
	assign func(S, RS),
) exec2.RelationLoader[S] {
	return BelongsToLoader[T, C, S, RT, RC, RS]{
		relation:   relation,
		related:    related,
		foreignKey: foreignKey,
		assign:     assign,
	}
}

// ManyToManyRelation represents a many-to-many relationship through a pivot table
// Example: Product <-> Categories through product_categories pivot table
type ManyToManyRelation[
	T types.TableAlias, // Current table alias (e.g., products)
	C types.ColumnAlias, // Current table column alias
	S exec2.Scanner[C], // Current table scanner
	PT types.TableAlias, // Pivot table alias (e.g., product_categories)
	PC types.ColumnAlias, // Pivot table column alias
	RT types.TableAlias, // Related table alias (e.g., categories)
	RC types.ColumnAlias, // Related table column alias
	RS exec2.Scanner[RC], // Related table scanner
] struct {
	localKeyName        string // Primary key in current table (e.g., product_id)
	pivotLocalKeyName   string // Foreign key in pivot table pointing to current table
	pivotRelatedKeyName string // Foreign key in pivot table pointing to related table
	relatedKeyName      string // Primary key in related table (e.g., category_id)
	pivotName           PT     // Name of the pivot table
	relatedName         RT     // Name of the related table
}

// ManyToMany creates a many-to-many relation through a pivot table
// Example: Product has many Categories through product_categories
func ManyToMany[
	T types.TableAlias, C types.ColumnAlias, S exec2.Scanner[C],
	PT types.TableAlias, PC types.ColumnAlias,
	RT types.TableAlias, RC types.ColumnAlias, RS exec2.Scanner[RC],
](
	currentTableAlias T,
	pivotTable RelationTableAlias[PT],
	relatedTable RelationTableAlias[RT],
	localKey C, // primary key in current table
	pivotLocalKey PC, // foreign key in pivot table pointing to current table
	pivotRelatedKey PC, // foreign key in pivot table pointing to related table
	relatedKey RC, // primary key in related table
) *ManyToManyRelation[T, C, S, PT, PC, RT, RC, RS] {
	return &ManyToManyRelation[T, C, S, PT, PC, RT, RC, RS]{
		localKeyName:        localKey.String(),
		pivotLocalKeyName:   pivotLocalKey.String(),
		pivotRelatedKeyName: pivotRelatedKey.String(),
		relatedKeyName:      relatedKey.String(),
		pivotName:           pivotTable.Alias(),
		relatedName:         relatedTable.Alias(),
	}
}

// WithJoin adds JOIN clauses to a SelectQuery for many-to-many relation
// Joins: current_table -> pivot_table -> related_table
func (r *ManyToManyRelation[T, C, S, PT, PC, RT, RC, RS]) WithJoin(
	currentTable RelationTableJoin[T, C],
	query *dml.SelectQuery[T, C],
	joinType dml.JoinType,
) *dml.SelectQuery[T, C] {
	currentTableName := currentTable.Alias().String()

	// First JOIN: current_table -> pivot_table
	// ON current_table.localKey = pivot_table.pivotLocalKey
	pivotOnClause := currentTable.Raw(
		fmt.Sprintf("%s.%s = %s.%s",
			currentTableName, r.localKeyName,
			r.pivotName.String(), r.pivotLocalKeyName),
	)
	query = query.Join(joinType, r.pivotName.String(), r.pivotName.String(), pivotOnClause)

	// Second JOIN: pivot_table -> related_table
	// ON pivot_table.pivotRelatedKey = related_table.relatedKey
	relatedOnClause := currentTable.Raw(
		fmt.Sprintf("%s.%s = %s.%s",
			r.pivotName.String(), r.pivotRelatedKeyName,
			r.relatedName.String(), r.relatedKeyName),
	)
	query = query.Join(joinType, r.relatedName.String(), r.relatedName.String(), relatedOnClause)

	return query
}

// LoadMany loads related records for a many-to-many relation
func (r *ManyToManyRelation[T, C, S, PT, PC, RT, RC, RS]) LoadMany(
	ctx context.Context,
	db exec2.DB,
	relatedTable RelationTableQuery[RT, RC, RS],
	localValue any, // value of the local key (e.g., product.ID)
) ([]RS, error) {
	// Build subquery to get related IDs from pivot table
	// SELECT related_key FROM pivot_table WHERE pivot_local_key = $1
	subQuery := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = $1",
		r.pivotRelatedKeyName,
		r.pivotName.String(),
		r.pivotLocalKeyName,
	)

	// Build main query: SELECT * FROM related_table WHERE related_key IN (subquery)
	rawWhere := relatedTable.Raw(
		fmt.Sprintf("%s.%s IN (%s)", r.relatedName.String(), r.relatedKeyName, subQuery),
		localValue,
	)

	query := relatedTable.SelectAll().Where(rawWhere)

	// Execute query
	results, err := relatedTable.Query(ctx, db, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load related records: %w", err)
	}

	return results, nil
}

// ManyToManyLoader implements RelationLoader for many-to-many relations
type ManyToManyLoader[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec2.Scanner[C],
	PT types.TableAlias,
	PC types.ColumnAlias,
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
] struct {
	relation *ManyToManyRelation[T, C, S, PT, PC, RT, RC, RS]
	related  RelationTableQuery[RT, RC, RS]
	localKey C
	assign   func(S, []RS)
}

func (l ManyToManyLoader[T, C, S, PT, PC, RT, RC, RS]) Load(ctx context.Context, db exec2.DB, base S) error {
	localValue := base.GetValue(l.localKey)()
	rows, err := l.relation.LoadMany(exec2.WithSkipRelations(ctx), db, l.related, localValue)
	if err != nil {
		return err
	}
	if l.assign != nil {
		l.assign(base, rows)
	}
	return nil
}

// ManyToManyLoad creates a loader for many-to-many relations
func ManyToManyLoad[
	T types.TableAlias,
	C types.ColumnAlias,
	S exec2.Scanner[C],
	PT types.TableAlias,
	PC types.ColumnAlias,
	RT types.TableAlias,
	RC types.ColumnAlias,
	RS exec2.Scanner[RC],
](
	relation *ManyToManyRelation[T, C, S, PT, PC, RT, RC, RS],
	related RelationTableQuery[RT, RC, RS],
	localKey C,
	assign func(S, []RS),
) exec2.RelationLoader[S] {
	return ManyToManyLoader[T, C, S, PT, PC, RT, RC, RS]{
		relation: relation,
		related:  related,
		localKey: localKey,
		assign:   assign,
	}
}
