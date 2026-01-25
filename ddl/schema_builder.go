package ddl

import "strings"

type SchemaSqler interface {
	SchemaSql() []string
}

func SchemaStatements(sqlers ...SchemaSqler) []string {
	statements := make([]string, 0, len(sqlers)*2)
	for _, sqler := range sqlers {
		if sqler == nil {
			continue
		}
		statements = append(statements, sqler.SchemaSql()...)
	}
	return statements
}

func SchemaSQL(sqlers ...SchemaSqler) string {
	statements := SchemaStatements(sqlers...)
	if len(statements) == 0 {
		return ""
	}
	content := strings.Join(statements, ";\n")
	if !strings.HasSuffix(content, ";") {
		content += ";"
	}
	return content + "\n"
}
