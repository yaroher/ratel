package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/yaroher/ratel/pkg/migrate"
)

// inspectExtensions returns all non-plpgsql extensions installed in the database,
// annotated with the schema they are installed into. Extensions are database-wide,
// so schemaName is not used as a filter — it is retained for API symmetry.
func (ins *Inspector) inspectExtensions(ctx context.Context, _ string) ([]migrate.Extension, error) {
	const q = `
SELECT e.extname, n.nspname, e.extversion
FROM pg_extension e
JOIN pg_namespace n ON e.extnamespace = n.oid
WHERE e.extname != 'plpgsql'
ORDER BY e.extname
`
	rows, err := ins.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (migrate.Extension, error) {
		var ext migrate.Extension
		return ext, row.Scan(&ext.Name, &ext.Schema, &ext.Version)
	})
}
