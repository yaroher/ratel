package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/yaroher/ratel/pkg/migrate"
)

// tgtype bitmask constants (int16 from pg_trigger.tgtype).
const (
	tgtypeRow       = 1  // bit 0: ROW level
	tgtypeBefore    = 2  // bit 1: BEFORE
	tgtypeInsert    = 4  // bit 2: INSERT
	tgtypeDelete    = 8  // bit 3: DELETE
	tgtypeUpdate    = 16 // bit 4: UPDATE
	tgtypeTruncate  = 32 // bit 5: TRUNCATE
	tgtypeInsteadOf = 64 // bit 6: INSTEAD OF
)

// inspectTriggers returns all non-internal triggers defined on a table.
func (ins *Inspector) inspectTriggers(ctx context.Context, schemaName, tableName string) ([]migrate.Trigger, error) {
	const q = `
SELECT
    t.tgname,
    t.tgtype::int,
    p.proname,
    pg_get_triggerdef(t.oid)
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_proc p ON t.tgfoid = p.oid
WHERE c.relname = $1
  AND n.nspname = $2
  AND NOT t.tgisinternal
ORDER BY t.tgname
`
	rows, err := ins.pool.Query(ctx, q, tableName, schemaName)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (migrate.Trigger, error) {
		var (
			name       string
			tgtype     int
			proname    string
			triggerdef string
		)
		if err := row.Scan(&name, &tgtype, &proname, &triggerdef); err != nil {
			return migrate.Trigger{}, err
		}

		trig := migrate.Trigger{
			Name:     name,
			Table:    tableName,
			Function: proname,
		}

		// ForEachRow: bit 0
		trig.ForEachRow = (tgtype & tgtypeRow) != 0

		// Timing: INSTEAD OF > BEFORE > AFTER
		switch {
		case (tgtype & tgtypeInsteadOf) != 0:
			trig.Timing = "INSTEAD OF"
		case (tgtype & tgtypeBefore) != 0:
			trig.Timing = "BEFORE"
		default:
			trig.Timing = "AFTER"
		}

		// Events: bits 2,3,4,5
		if (tgtype & tgtypeInsert) != 0 {
			trig.Events = append(trig.Events, "INSERT")
		}
		if (tgtype & tgtypeDelete) != 0 {
			trig.Events = append(trig.Events, "DELETE")
		}
		if (tgtype & tgtypeUpdate) != 0 {
			trig.Events = append(trig.Events, "UPDATE")
		}
		if (tgtype & tgtypeTruncate) != 0 {
			trig.Events = append(trig.Events, "TRUNCATE")
		}

		// Extract WHEN clause from triggerdef, e.g. "... WHEN (...) EXECUTE ..."
		trig.When = extractTriggerWhen(triggerdef)

		// Extract args from triggerdef: text after EXECUTE FUNCTION/PROCEDURE name(...)
		trig.Args = extractTriggerArgs(triggerdef)

		return trig, nil
	})
}

// extractTriggerWhen parses the WHEN clause out of a pg_get_triggerdef string.
// The definition looks like: "... WHEN (<expr>) EXECUTE FUNCTION ..."
func extractTriggerWhen(def string) string {
	const whenMarker = " WHEN ("
	idx := strings.Index(def, whenMarker)
	if idx < 0 {
		return ""
	}
	// The WHEN expression is wrapped in parentheses immediately after " WHEN ".
	start := idx + len(" WHEN (") - 1 // position of the opening '('
	depth := 0
	for i := start; i < len(def); i++ {
		switch def[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				// Return the content inside the outermost parens.
				return def[start+1 : i]
			}
		}
	}
	return ""
}

// extractTriggerArgs parses the argument list from pg_get_triggerdef.
// The relevant portion looks like: "EXECUTE FUNCTION funcname(arg1, arg2)"
// or "EXECUTE PROCEDURE funcname(arg1, arg2)".
func extractTriggerArgs(def string) []string {
	// Find the EXECUTE keyword and then the opening paren of the arg list.
	upper := strings.ToUpper(def)
	execIdx := strings.LastIndex(upper, "EXECUTE ")
	if execIdx < 0 {
		return nil
	}
	rest := def[execIdx:]
	openParen := strings.Index(rest, "(")
	if openParen < 0 {
		return nil
	}
	closeParen := strings.Index(rest[openParen:], ")")
	if closeParen < 0 {
		return nil
	}
	argStr := strings.TrimSpace(rest[openParen+1 : openParen+closeParen])
	if argStr == "" {
		return nil
	}
	parts := strings.Split(argStr, ",")
	args := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			// Strip surrounding single quotes added by quote_literal if present.
			trimmed = strings.Trim(trimmed, "'")
			args = append(args, trimmed)
		}
	}
	return args
}
