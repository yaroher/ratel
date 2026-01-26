package ddl

import (
	"fmt"
	"strings"
)

type SchemaSqler interface {
	SchemaSql() []string
}

// DependencySqler is an optional interface for tables that have foreign key dependencies
type DependencySqler interface {
	SchemaSqler
	// TableName returns the name of this table
	TableName() string
	// Dependencies returns names of tables this table depends on (via foreign keys)
	Dependencies() []string
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

// SchemaSortedStatements returns SQL statements with tables sorted by dependencies.
// Tables without foreign keys come first, then tables that depend on them, etc.
func SchemaSortedStatements(sqlers ...SchemaSqler) ([]string, error) {
	// Separate tables with dependency info from those without
	var depSqlers []DependencySqler
	var otherSqlers []SchemaSqler

	for _, sqler := range sqlers {
		if sqler == nil {
			continue
		}
		if dep, ok := sqler.(DependencySqler); ok {
			depSqlers = append(depSqlers, dep)
		} else {
			otherSqlers = append(otherSqlers, sqler)
		}
	}

	// Sort tables with dependencies topologically
	sorted, err := topologicalSort(depSqlers)
	if err != nil {
		return nil, err
	}

	// Collect statements: first other sqlers, then sorted tables
	statements := make([]string, 0, len(sqlers)*2)
	for _, sqler := range otherSqlers {
		statements = append(statements, sqler.SchemaSql()...)
	}
	for _, sqler := range sorted {
		statements = append(statements, sqler.SchemaSql()...)
	}

	return statements, nil
}

// topologicalSort sorts tables by their foreign key dependencies using Kahn's algorithm
func topologicalSort(tables []DependencySqler) ([]DependencySqler, error) {
	if len(tables) == 0 {
		return nil, nil
	}

	// Build adjacency list and in-degree map
	tableMap := make(map[string]DependencySqler)
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // table -> tables that depend on it

	for _, t := range tables {
		name := t.TableName()
		tableMap[name] = t
		inDegree[name] = 0
	}

	// Calculate in-degrees based on dependencies
	for _, t := range tables {
		name := t.TableName()
		for _, dep := range t.Dependencies() {
			// Only count dependency if the referenced table is in our set
			// (external references don't affect sort order)
			if _, exists := tableMap[dep]; exists {
				inDegree[name]++
				dependents[dep] = append(dependents[dep], name)
			}
		}
	}

	// Start with tables that have no dependencies (in-degree = 0)
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []DependencySqler
	for len(queue) > 0 {
		// Pop from queue
		name := queue[0]
		queue = queue[1:]

		result = append(result, tableMap[name])

		// Reduce in-degree for dependent tables
		for _, dependent := range dependents[name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(result) != len(tables) {
		var remaining []string
		for name, degree := range inDegree {
			if degree > 0 {
				remaining = append(remaining, name)
			}
		}
		return nil, fmt.Errorf("circular dependency detected among tables: %v", remaining)
	}

	return result, nil
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

// SchemaSortedSQL returns SQL with tables sorted by dependencies
func SchemaSortedSQL(sqlers ...SchemaSqler) (string, error) {
	statements, err := SchemaSortedStatements(sqlers...)
	if err != nil {
		return "", err
	}
	if len(statements) == 0 {
		return "", nil
	}
	content := strings.Join(statements, ";\n")
	if !strings.HasSuffix(content, ";") {
		content += ";"
	}
	return content + "\n", nil
}
