package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/yaroher/ratel/pkg/migrate"
)

// inspectFunctions returns all ordinary functions (prokind = 'f') defined in
// the given schema, populated with arguments, return type, language, body,
// volatility, security, and optional comment.
func (ins *Inspector) inspectFunctions(ctx context.Context, schemaName string) ([]migrate.Function, error) {
	const q = `
SELECT
    p.proname,
    n.nspname,
    pg_get_function_arguments(p.oid) AS args,
    pg_get_function_result(p.oid)    AS return_type,
    l.lanname,
    p.prosrc,
    p.provolatile::text,
    p.prosecdef,
    d.description
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
JOIN pg_language  l ON p.prolang      = l.oid
LEFT JOIN pg_description d
       ON d.objoid   = p.oid
      AND d.classoid = 'pg_proc'::regclass
WHERE n.nspname = $1
  AND p.prokind::text = 'f'
ORDER BY p.proname
`
	rows, err := ins.pool.Query(ctx, q, schemaName)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (migrate.Function, error) {
		var (
			name        string
			schema      string
			argsStr     string
			returnType  string
			language    string
			body        string
			provolatile string
			prosecdef   bool
			comment     *string
		)
		if err := row.Scan(
			&name, &schema, &argsStr, &returnType,
			&language, &body, &provolatile, &prosecdef,
			&comment,
		); err != nil {
			return migrate.Function{}, err
		}

		fn := migrate.Function{
			Name:       name,
			Schema:     schema,
			Args:       parseArgString(argsStr),
			ReturnType: returnType,
			Language:   language,
			Body:       body,
			Volatility: mapVolatility(provolatile),
			Security:   mapSecurity(prosecdef),
		}
		if comment != nil {
			fn.Comment = *comment
		}
		return fn, nil
	})
}

// mapVolatility converts the single-character provolatile value stored in
// pg_proc to the corresponding SQL keyword.
func mapVolatility(v string) string {
	switch v {
	case "s":
		return "STABLE"
	case "i":
		return "IMMUTABLE"
	default: // 'v'
		return "VOLATILE"
	}
}

// mapSecurity converts the prosecdef boolean to the SQL SECURITY clause value.
func mapSecurity(secdef bool) string {
	if secdef {
		return "DEFINER"
	}
	return "INVOKER"
}

// parseArgString splits the string produced by pg_get_function_arguments into
// individual FunctionArg values.
//
// The format produced by PostgreSQL is:
//
//	[mode] [name] type [DEFAULT expr], ...
//
// where mode and name are optional. Examples:
//
//	"a integer, b text DEFAULT 'foo'"
//	"IN p1 integer, OUT p2 text"
//	"VARIADIC vals integer[]"
//	""  (no arguments)
func parseArgString(s string) []migrate.FunctionArg {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Split on top-level commas (commas inside parentheses belong to type
	// expressions such as numeric(10,2) and must not be treated as separators).
	parts := splitTopLevel(s)

	args := make([]migrate.FunctionArg, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		args = append(args, parseOneArg(part))
	}
	return args
}

// splitTopLevel splits src on commas that are not enclosed in parentheses.
func splitTopLevel(src string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(src); i++ {
		switch src[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, src[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, src[start:])
	return parts
}

// knownModes is the set of PostgreSQL argument mode keywords.
var knownModes = map[string]bool{
	"IN":       true,
	"OUT":      true,
	"INOUT":    true,
	"VARIADIC": true,
}

// parseOneArg parses a single argument fragment such as:
//
//	"a integer"
//	"IN a integer"
//	"OUT b text DEFAULT NULL"
//	"integer"   (anonymous, no name)
func parseOneArg(s string) migrate.FunctionArg {
	var arg migrate.FunctionArg

	// Separate DEFAULT clause first (case-insensitive, only top-level).
	if idx := findDefaultKeyword(s); idx >= 0 {
		arg.Default = strings.TrimSpace(s[idx+len("DEFAULT"):])
		s = strings.TrimSpace(s[:idx])
	}

	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return arg
	}

	// Check for a leading mode keyword.
	if knownModes[strings.ToUpper(tokens[0])] {
		arg.Mode = strings.ToUpper(tokens[0])
		tokens = tokens[1:]
	}

	switch len(tokens) {
	case 0:
		// nothing left
	case 1:
		// Either just a type (anonymous arg) or edge case.
		arg.Type = tokens[0]
	default:
		// First token is the argument name, the rest form the type.
		arg.Name = tokens[0]
		arg.Type = strings.Join(tokens[1:], " ")
	}

	return arg
}

// findDefaultKeyword returns the byte index of the word DEFAULT in s,
// respecting top-level token boundaries (not inside parentheses).
// Returns -1 if not found.
func findDefaultKeyword(s string) int {
	upper := strings.ToUpper(s)
	search := "DEFAULT"
	offset := 0
	for {
		idx := strings.Index(upper[offset:], search)
		if idx < 0 {
			return -1
		}
		abs := offset + idx
		// Verify it is a word boundary (not part of a larger identifier).
		before := abs == 0 || !isIdentChar(rune(upper[abs-1]))
		after := abs+len(search) >= len(upper) || !isIdentChar(rune(upper[abs+len(search)]))
		if before && after {
			return abs
		}
		offset = abs + 1
	}
}

func isIdentChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') || r == '_'
}
