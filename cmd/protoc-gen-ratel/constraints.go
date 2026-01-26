package main

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

// ConstraintType represents the type of database constraint
type ConstraintType string

const (
	ConstraintTypeUnique     ConstraintType = "unique"
	ConstraintTypePrimaryKey ConstraintType = "primary_key"
	ConstraintTypeForeignKey ConstraintType = "foreign_key"
	ConstraintTypeCheck      ConstraintType = "check"
)

// ConstraintInfo describes a constraint for code generation
type ConstraintInfo struct {
	Name        string         // SQL constraint name: "users_email_key"
	GoConstName string         // Go constant name: "UserConstraintEmailKey"
	GoErrorName string         // Go error var name: "ErrUserEmailUnique"
	GoCheckFunc string         // Go check function name: "IsUserEmailUniqueError"
	Type        ConstraintType // unique, pk, fk, check
	Table       string         // SQL table name
	Columns     []string       // columns involved
	Message     string         // Go message name
}

// collectConstraints gathers all constraints from a table for code generation
func collectConstraints(table *RatelTable) []*ConstraintInfo {
	var constraints []*ConstraintInfo

	msgName := table.Message.GoIdent.GoName
	tableName := getTableName(table)

	// 1. Collect column-level constraints
	for _, col := range table.Columns {
		if col.Options == nil || col.Options.Constraints == nil {
			continue
		}

		c := col.Options.Constraints

		// UNIQUE constraint on column
		if c.Unique {
			name := fmt.Sprintf("%s_%s_key", tableName, col.SQLName)
			constraints = append(constraints, &ConstraintInfo{
				Name:        name,
				GoConstName: msgName + "Constraint" + strcase.ToCamel(col.SQLName) + "Key",
				GoErrorName: "Err" + msgName + strcase.ToCamel(col.SQLName) + "Unique",
				GoCheckFunc: "Is" + msgName + strcase.ToCamel(col.SQLName) + "UniqueError",
				Type:        ConstraintTypeUnique,
				Table:       tableName,
				Columns:     []string{col.SQLName},
				Message:     msgName,
			})
		}

		// PRIMARY KEY constraint on column
		if c.PrimaryKey {
			name := fmt.Sprintf("%s_pkey", tableName)
			constraints = append(constraints, &ConstraintInfo{
				Name:        name,
				GoConstName: msgName + "ConstraintPkey",
				GoErrorName: "Err" + msgName + "PrimaryKey",
				GoCheckFunc: "Is" + msgName + "PrimaryKeyError",
				Type:        ConstraintTypePrimaryKey,
				Table:       tableName,
				Columns:     []string{col.SQLName},
				Message:     msgName,
			})
		}
	}

	// 2. Collect table-level unique constraints
	if table.Options != nil && len(table.Options.Unique) > 0 {
		for _, uq := range table.Options.Unique {
			colNames := append([]string{}, uq.Columns...)

			// Generate constraint name from columns
			name := uq.Name
			if name == "" {
				name = fmt.Sprintf("%s_%s_key", tableName, strings.Join(colNames, "_"))
			}

			// Generate Go names
			colsCamel := strcase.ToCamel(strings.Join(colNames, "_"))
			constraints = append(constraints, &ConstraintInfo{
				Name:        name,
				GoConstName: msgName + "Constraint" + colsCamel + "Key",
				GoErrorName: "Err" + msgName + colsCamel + "Unique",
				GoCheckFunc: "Is" + msgName + colsCamel + "UniqueError",
				Type:        ConstraintTypeUnique,
				Table:       tableName,
				Columns:     colNames,
				Message:     msgName,
			})
		}
	}

	// 3. Collect composite primary key
	if table.Options != nil && table.Options.PrimaryKey != nil && len(table.Options.PrimaryKey.Columns) > 0 {
		name := fmt.Sprintf("%s_pkey", tableName)
		constraints = append(constraints, &ConstraintInfo{
			Name:        name,
			GoConstName: msgName + "ConstraintPkey",
			GoErrorName: "Err" + msgName + "PrimaryKey",
			GoCheckFunc: "Is" + msgName + "PrimaryKeyError",
			Type:        ConstraintTypePrimaryKey,
			Table:       tableName,
			Columns:     table.Options.PrimaryKey.Columns,
			Message:     msgName,
		})
	}

	// 4. Collect table-level CHECK constraints
	if table.Options != nil && len(table.Options.Constraints) > 0 {
		for i, checkExpr := range table.Options.Constraints {
			// Extract constraint name from CHECK expression if possible
			// Default naming: tablename_check or tablename_check1, etc.
			name := fmt.Sprintf("%s_check", tableName)
			if i > 0 {
				name = fmt.Sprintf("%s_check%d", tableName, i)
			}

			// Try to extract meaningful name from expression
			goSuffix := "Check"
			if i > 0 {
				goSuffix = fmt.Sprintf("Check%d", i)
			}

			constraints = append(constraints, &ConstraintInfo{
				Name:        name,
				GoConstName: msgName + "Constraint" + goSuffix,
				GoErrorName: "Err" + msgName + goSuffix,
				GoCheckFunc: "Is" + msgName + goSuffix + "Error",
				Type:        ConstraintTypeCheck,
				Table:       tableName,
				Columns:     nil, // CHECK can involve multiple columns
				Message:     msgName,
			})

			// Use checkExpr to suppress unused variable warning
			_ = checkExpr
		}
	}

	// 5. Collect unique indexes (they also create constraints)
	if table.Options != nil && len(table.Options.Indexes) > 0 {
		for _, idx := range table.Options.Indexes {
			if !idx.Unique {
				continue
			}

			// Generate index/constraint name
			name := idx.Name
			if name == "" {
				name = fmt.Sprintf("idx_%s_%s", tableName, strings.Join(idx.Columns, "_"))
			}

			// Generate Go names
			colsCamel := strcase.ToCamel(strings.Join(idx.Columns, "_"))
			constraints = append(constraints, &ConstraintInfo{
				Name:        name,
				GoConstName: msgName + "Constraint" + colsCamel + "Idx",
				GoErrorName: "Err" + msgName + colsCamel + "UniqueIdx",
				GoCheckFunc: "Is" + msgName + colsCamel + "UniqueIdxError",
				Type:        ConstraintTypeUnique,
				Table:       tableName,
				Columns:     idx.Columns,
				Message:     msgName,
			})
		}
	}

	// Deduplicate constraints by name (column-level PK might duplicate table-level PK)
	return deduplicateConstraints(constraints)
}

// deduplicateConstraints removes duplicate constraints by name
func deduplicateConstraints(constraints []*ConstraintInfo) []*ConstraintInfo {
	seen := make(map[string]bool)
	var result []*ConstraintInfo

	for _, c := range constraints {
		if !seen[c.Name] {
			seen[c.Name] = true
			result = append(result, c)
		}
	}

	return result
}
