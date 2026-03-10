package postgres

// Based on Atlas (https://github.com/ariga/atlas) — Apache 2.0 License
// Copyright 2021-present The Atlas Authors

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yaroher/ratel/pkg/migrate"
)

// Inspector inspects a PostgreSQL database via system catalogs.
type Inspector struct {
	pool *pgxpool.Pool
}

// NewInspector creates a new Inspector backed by the given connection pool.
func NewInspector(pool *pgxpool.Pool) *Inspector {
	return &Inspector{pool: pool}
}

// InspectRealm inspects all user-created schemas in the database.
func (ins *Inspector) InspectRealm(ctx context.Context) (*migrate.SchemaState, error) {
	names, err := ins.listSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres inspector: list schemas: %w", err)
	}

	state := &migrate.SchemaState{}
	for _, name := range names {
		schema, err := ins.InspectSchema(ctx, name)
		if err != nil {
			return nil, err
		}
		state.Schemas = append(state.Schemas, *schema)
	}
	return state, nil
}

// InspectSchema inspects a single schema by name.
func (ins *Inspector) InspectSchema(ctx context.Context, name string) (*migrate.Schema, error) {
	schema := &migrate.Schema{Name: name}

	tables, err := ins.inspectTables(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("postgres inspector: inspect tables for schema %q: %w", name, err)
	}
	schema.Tables = tables

	extensions, err := ins.inspectExtensions(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("postgres inspector: inspect extensions for schema %q: %w", name, err)
	}
	schema.Extensions = extensions

	functions, err := ins.inspectFunctions(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("postgres inspector: inspect functions for schema %q: %w", name, err)
	}
	schema.Functions = functions

	return schema, nil
}

// listSchemas returns names of user-created schemas, excluding PostgreSQL internal schemas.
func (ins *Inspector) listSchemas(ctx context.Context) ([]string, error) {
	const q = `
SELECT nspname
FROM pg_namespace
WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
  AND nspname NOT LIKE 'pg_temp_%'
  AND nspname NOT LIKE 'pg_toast_temp_%'
ORDER BY nspname
`
	rows, err := ins.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		var name string
		return name, row.Scan(&name)
	})
}

// inspectTables returns all base tables in the given schema, fully populated.
func (ins *Inspector) inspectTables(ctx context.Context, schemaName string) ([]migrate.Table, error) {
	const q = `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = $1
  AND table_type = 'BASE TABLE'
ORDER BY table_name
`
	rows, err := ins.pool.Query(ctx, q, schemaName)
	if err != nil {
		return nil, err
	}
	tableNames, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		var n string
		return n, row.Scan(&n)
	})
	if err != nil {
		return nil, err
	}

	tables := make([]migrate.Table, 0, len(tableNames))
	for _, tname := range tableNames {
		t := migrate.Table{Name: tname, Schema: schemaName}

		t.Columns, err = ins.inspectColumns(ctx, schemaName, tname)
		if err != nil {
			return nil, fmt.Errorf("columns for %s.%s: %w", schemaName, tname, err)
		}

		pk, indexes, err := ins.inspectIndexes(ctx, schemaName, tname)
		if err != nil {
			return nil, fmt.Errorf("indexes for %s.%s: %w", schemaName, tname, err)
		}
		t.PrimaryKey = pk
		t.Indexes = indexes

		t.ForeignKeys, err = ins.inspectForeignKeys(ctx, schemaName, tname)
		if err != nil {
			return nil, fmt.Errorf("foreign keys for %s.%s: %w", schemaName, tname, err)
		}

		t.Checks, err = ins.inspectChecks(ctx, schemaName, tname)
		if err != nil {
			return nil, fmt.Errorf("checks for %s.%s: %w", schemaName, tname, err)
		}

		t.RLSEnabled, t.RLSForced, err = ins.inspectRLS(ctx, schemaName, tname)
		if err != nil {
			return nil, fmt.Errorf("rls for %s.%s: %w", schemaName, tname, err)
		}

		t.Policies, err = ins.inspectPolicies(ctx, schemaName, tname)
		if err != nil {
			return nil, fmt.Errorf("policies for %s.%s: %w", schemaName, tname, err)
		}

		t.Triggers, err = ins.inspectTriggers(ctx, schemaName, tname)
		if err != nil {
			return nil, fmt.Errorf("triggers for %s.%s: %w", schemaName, tname, err)
		}

		tables = append(tables, t)
	}
	return tables, nil
}

// inspectColumns returns all columns for a table, including identity and comment info.
func (ins *Inspector) inspectColumns(ctx context.Context, schemaName, tableName string) ([]migrate.Column, error) {
	const q = `
SELECT
    c.column_name,
    c.udt_name,
    c.is_nullable,
    c.column_default,
    c.is_identity,
    c.identity_generation,
    d.description
FROM information_schema.columns c
LEFT JOIN pg_class pc ON pc.relname = c.table_name
    AND pc.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = c.table_schema)
LEFT JOIN pg_attribute pa ON pa.attrelid = pc.oid AND pa.attname = c.column_name
LEFT JOIN pg_description d ON d.objoid = pa.attrelid AND d.objsubid = pa.attnum
WHERE c.table_schema = $1
  AND c.table_name  = $2
ORDER BY c.ordinal_position
`
	rows, err := ins.pool.Query(ctx, q, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []migrate.Column
	for rows.Next() {
		var (
			colName       string
			udtName       string
			isNullableStr string
			colDefault    *string
			isIdentityStr string
			identityGen   *string
			comment       *string
		)
		if err := rows.Scan(
			&colName, &udtName, &isNullableStr,
			&colDefault, &isIdentityStr, &identityGen,
			&comment,
		); err != nil {
			return nil, err
		}

		col := migrate.Column{
			Name:     colName,
			Type:     udtName,
			Nullable: strings.EqualFold(isNullableStr, "YES"),
		}
		if colDefault != nil {
			col.Default = *colDefault
		}
		if comment != nil {
			col.Comment = *comment
		}
		if strings.EqualFold(isIdentityStr, "YES") && identityGen != nil {
			col.Identity = &migrate.Identity{
				Generation: *identityGen,
			}
		}

		cols = append(cols, col)
	}
	return cols, rows.Err()
}

// inspectIndexes returns the primary key and secondary indexes for a table.
// Returns (pk *migrate.Index, indexes []migrate.Index, err error).
func (ins *Inspector) inspectIndexes(ctx context.Context, schemaName, tableName string) (*migrate.Index, []migrate.Index, error) {
	// Query all indexes (including primary key) for the table.
	const q = `
SELECT
    i.indexrelid,
    ic.relname                                                        AS index_name,
    i.indisunique,
    i.indisprimary,
    am.amname,
    pg_get_expr(i.indpred, i.indrelid)                               AS predicate,
    i.indnkeyatts,
    array_agg(a.attname ORDER BY array_position(i.indkey::int[], a.attnum)) AS key_columns,
    array_agg(a.attnum   ORDER BY array_position(i.indkey::int[], a.attnum)) AS key_attnums
FROM pg_index i
JOIN pg_class ic  ON ic.oid  = i.indexrelid
JOIN pg_class tc  ON tc.oid  = i.indrelid
JOIN pg_am    am  ON ic.relam = am.oid
JOIN pg_attribute a ON a.attrelid = tc.oid AND a.attnum = ANY(i.indkey)
WHERE tc.relname       = $1
  AND tc.relnamespace  = (SELECT oid FROM pg_namespace WHERE nspname = $2)
  AND a.attnum > 0
GROUP BY i.indexrelid, ic.relname, i.indisunique, i.indisprimary, am.amname, i.indpred, i.indrelid, i.indnkeyatts
ORDER BY ic.relname
`

	// For INCLUDE columns, we need the full indkey array.
	// We query include columns separately via a second query using indexrelid.
	rows, err := ins.pool.Query(ctx, q, tableName, schemaName)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	type rawIndex struct {
		indexrelid  uint32
		name        string
		unique      bool
		primary     bool
		method      string
		predicate   *string
		indnkeyatts int
		keyCols     []string
		keyAttnums  []int32
	}

	var raws []rawIndex
	for rows.Next() {
		var r rawIndex
		if err := rows.Scan(
			&r.indexrelid,
			&r.name,
			&r.unique,
			&r.primary,
			&r.method,
			&r.predicate,
			&r.indnkeyatts,
			&r.keyCols,
			&r.keyAttnums,
		); err != nil {
			return nil, nil, err
		}
		raws = append(raws, r)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var pk *migrate.Index
	var indexes []migrate.Index

	for _, r := range raws {
		// Key columns: first indnkeyatts entries are key columns; the rest are INCLUDE columns.
		keyCols := r.keyCols
		var includeCols []string
		if r.indnkeyatts > 0 && r.indnkeyatts < len(r.keyCols) {
			includeCols = r.keyCols[r.indnkeyatts:]
			keyCols = r.keyCols[:r.indnkeyatts]
		}

		idxCols := make([]migrate.IndexColumn, len(keyCols))
		for j, c := range keyCols {
			idxCols[j] = migrate.IndexColumn{Name: c}
		}

		idx := migrate.Index{
			Name:    r.name,
			Columns: idxCols,
			Unique:  r.unique,
			Method:  r.method,
			Include: includeCols,
		}
		if r.predicate != nil {
			idx.Where = *r.predicate
		}

		if r.primary {
			pk = &idx
		} else {
			indexes = append(indexes, idx)
		}
	}

	return pk, indexes, nil
}

// fkAction maps PostgreSQL FK action chars to human-readable strings.
func fkAction(ch string) string {
	switch ch {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return ch
	}
}

// inspectForeignKeys returns all foreign keys defined on a table.
func (ins *Inspector) inspectForeignKeys(ctx context.Context, schemaName, tableName string) ([]migrate.ForeignKey, error) {
	const q = `
SELECT
    c.conname,
    array_agg(a.attname  ORDER BY array_position(c.conkey::int[],  a.attnum))  AS cols,
    ref_class.relname                                                            AS ref_table,
    ref_ns.nspname                                                               AS ref_schema,
    array_agg(ra.attname ORDER BY array_position(c.confkey::int[], ra.attnum)) AS ref_cols,
    c.confupdtype::text,
    c.confdeltype::text
FROM pg_constraint c
JOIN pg_class       ON pg_class.oid       = c.conrelid
JOIN pg_namespace   ON pg_namespace.oid   = pg_class.relnamespace
JOIN pg_attribute   a  ON a.attrelid      = c.conrelid  AND a.attnum  = ANY(c.conkey)
JOIN pg_class       ref_class ON ref_class.oid = c.confrelid
JOIN pg_namespace   ref_ns    ON ref_ns.oid    = ref_class.relnamespace
JOIN pg_attribute   ra ON ra.attrelid     = c.confrelid AND ra.attnum = ANY(c.confkey)
WHERE c.contype        = 'f'
  AND pg_class.relname = $1
  AND pg_namespace.nspname = $2
GROUP BY c.conname, ref_class.relname, ref_ns.nspname, c.confupdtype, c.confdeltype
ORDER BY c.conname
`
	rows, err := ins.pool.Query(ctx, q, tableName, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []migrate.ForeignKey
	for rows.Next() {
		var (
			name      string
			cols      []string
			refTable  string
			refSchema string
			refCols   []string
			onUpdate  string
			onDelete  string
		)
		if err := rows.Scan(&name, &cols, &refTable, &refSchema, &refCols, &onUpdate, &onDelete); err != nil {
			return nil, err
		}
		fks = append(fks, migrate.ForeignKey{
			Name:       name,
			Columns:    cols,
			RefTable:   refTable,
			RefSchema:  refSchema,
			RefColumns: refCols,
			OnUpdate:   fkAction(onUpdate),
			OnDelete:   fkAction(onDelete),
		})
	}
	return fks, rows.Err()
}

// inspectChecks returns all CHECK constraints on a table.
func (ins *Inspector) inspectChecks(ctx context.Context, schemaName, tableName string) ([]migrate.Check, error) {
	const q = `
SELECT conname, pg_get_constraintdef(oid), connoinherit
FROM pg_constraint
WHERE contype = 'c'
  AND conrelid = (
      SELECT oid FROM pg_class
      WHERE relname = $1
        AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = $2)
  )
ORDER BY conname
`
	rows, err := ins.pool.Query(ctx, q, tableName, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []migrate.Check
	for rows.Next() {
		var (
			name      string
			expr      string
			noInherit bool
		)
		if err := rows.Scan(&name, &expr, &noInherit); err != nil {
			return nil, err
		}
		checks = append(checks, migrate.Check{
			Name:      name,
			Expr:      expr,
			NoInherit: noInherit,
		})
	}
	return checks, rows.Err()
}
