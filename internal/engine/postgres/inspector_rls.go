package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/yaroher/ratel/pkg/migrate"
)

// inspectRLS checks whether row-level security is enabled and forced for a table.
func (ins *Inspector) inspectRLS(ctx context.Context, schemaName, tableName string) (enabled bool, forced bool, err error) {
	const q = `
SELECT relrowsecurity, relforcerowsecurity
FROM pg_class
WHERE relname = $1
  AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = $2)
`
	row := ins.pool.QueryRow(ctx, q, tableName, schemaName)
	if err = row.Scan(&enabled, &forced); err != nil {
		return false, false, err
	}
	return enabled, forced, nil
}

// polCmdToString maps a pg_policy.polcmd char to a SQL command string.
func polCmdToString(cmd string) string {
	switch cmd {
	case "*":
		return "ALL"
	case "r":
		return "SELECT"
	case "a":
		return "INSERT"
	case "w":
		return "UPDATE"
	case "d":
		return "DELETE"
	default:
		return cmd
	}
}

// inspectPolicies returns all RLS policies defined on a table.
func (ins *Inspector) inspectPolicies(ctx context.Context, schemaName, tableName string) ([]migrate.Policy, error) {
	const q = `
SELECT
    pol.polname,
    pol.polpermissive,
    pol.polcmd::text,
    COALESCE(array_agg(r.rolname) FILTER (WHERE r.rolname IS NOT NULL), '{}'),
    pg_get_expr(pol.polqual, pol.polrelid),
    pg_get_expr(pol.polwithcheck, pol.polrelid)
FROM pg_policy pol
JOIN pg_class c ON c.oid = pol.polrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_roles r ON r.oid = ANY(pol.polroles)
WHERE c.relname = $1 AND n.nspname = $2
GROUP BY pol.polname, pol.polpermissive, pol.polcmd, pol.polqual, pol.polwithcheck, pol.polrelid
ORDER BY pol.polname
`
	rows, err := ins.pool.Query(ctx, q, tableName, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (migrate.Policy, error) {
		var (
			name       string
			permissive bool
			cmd        string
			roles      []string
			using      *string
			withCheck  *string
		)
		if err := row.Scan(&name, &permissive, &cmd, &roles, &using, &withCheck); err != nil {
			return migrate.Policy{}, err
		}
		p := migrate.Policy{
			Name:       name,
			Permissive: permissive,
			Command:    polCmdToString(cmd),
			Roles:      roles,
		}
		if using != nil {
			p.Using = *using
		}
		if withCheck != nil {
			p.WithCheck = *withCheck
		}
		return p, nil
	})
	if err != nil {
		return nil, err
	}
	return policies, nil
}
