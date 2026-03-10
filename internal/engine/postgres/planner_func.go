package postgres

import (
	"fmt"
	"strings"

	"github.com/yaroher/ratel/pkg/migrate"
)

func (p *Planner) planAddFunction(c migrate.AddFunction) ([]migrate.PlannedChange, error) {
	return planFunctionSQL(c.F), nil
}

func (p *Planner) planDropFunction(c migrate.DropFunction) ([]migrate.PlannedChange, error) {
	f := c.F
	argTypes := make([]string, len(f.Args))
	for i, a := range f.Args {
		argTypes[i] = a.Type
	}
	sql := fmt.Sprintf("DROP FUNCTION IF EXISTS %s(%s)",
		qualifiedName(f.Schema, f.Name),
		strings.Join(argTypes, ", "))
	return []migrate.PlannedChange{pc(sql, "drop function "+f.Name, "")}, nil
}

func (p *Planner) planModifyFunction(c migrate.ModifyFunction) ([]migrate.PlannedChange, error) {
	return planFunctionSQL(c.To), nil
}

// planFunctionSQL emits a CREATE OR REPLACE FUNCTION statement.
func planFunctionSQL(f *migrate.Function) []migrate.PlannedChange {
	// Build argument list
	args := make([]string, len(f.Args))
	for i, a := range f.Args {
		arg := ""
		if a.Mode != "" {
			arg += strings.ToUpper(a.Mode) + " "
		}
		if a.Name != "" {
			arg += q(a.Name) + " "
		}
		arg += a.Type
		if a.Default != "" {
			arg += " DEFAULT " + a.Default
		}
		args[i] = arg
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "CREATE OR REPLACE FUNCTION %s(%s)",
		qualifiedName(f.Schema, f.Name),
		strings.Join(args, ", "))

	fmt.Fprintf(&sb, "\nRETURNS %s", f.ReturnType)

	lang := f.Language
	if lang == "" {
		lang = "plpgsql"
	}
	fmt.Fprintf(&sb, "\nLANGUAGE %s", lang)

	if f.Volatility != "" {
		sb.WriteString("\n" + strings.ToUpper(f.Volatility))
	}

	security := strings.ToUpper(f.Security)
	if security == "DEFINER" {
		sb.WriteString("\nSECURITY DEFINER")
	} else {
		sb.WriteString("\nSECURITY INVOKER")
	}

	fmt.Fprintf(&sb, "\nAS $body$\n%s\n$body$", f.Body)

	rev := fmt.Sprintf("DROP FUNCTION IF EXISTS %s", qualifiedName(f.Schema, f.Name))
	return []migrate.PlannedChange{pc(sb.String(), "create or replace function "+f.Name, rev)}
}
