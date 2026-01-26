package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/yaroher/ratel/internal/parser"
)

var (
	generateInputFile string
	generateOutputDir string
	generatePackage   string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go models from SQL schema file",
	Long: `Generate Go model files from a PostgreSQL SQL schema file.

Example:
  ratel generate -i schema.sql -o ./models -p models

This will create one Go file per table in the output directory.`,
	Run: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&generateInputFile, "input", "i", "", "Input SQL schema file (required)")
	generateCmd.Flags().StringVarP(&generateOutputDir, "output", "o", "./models", "Output directory for generated models")
	generateCmd.Flags().StringVarP(&generatePackage, "package", "p", "models", "Package name for generated files")

	generateCmd.MarkFlagRequired("input")
}

func runGenerate(cmd *cobra.Command, args []string) {
	// Read input SQL file
	sqlContent, err := os.ReadFile(generateInputFile)
	if err != nil {
		exitWithError("failed to read input file: %v", err)
	}

	// Parse SQL
	schema, err := parser.ParseSQL(string(sqlContent))
	if err != nil {
		exitWithError("failed to parse SQL: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(generateOutputDir, 0755); err != nil {
		exitWithError("failed to create output directory: %v", err)
	}

	// Generate model files
	for _, table := range schema.Tables {
		if err := generateModelFile(table, generateOutputDir, generatePackage); err != nil {
			exitWithError("failed to generate model for table %s: %v", table.Name, err)
		}
		fmt.Printf("Generated: %s/%s.go\n", generateOutputDir, table.Name)
	}

	// Generate relations file
	if err := generateRelationsFile(schema.Tables, generateOutputDir, generatePackage); err != nil {
		exitWithError("failed to generate relations file: %v", err)
	}
	fmt.Printf("Generated: %s/relations.go\n", generateOutputDir)

	fmt.Printf("\nSuccessfully generated %d model(s) in %s\n", len(schema.Tables), generateOutputDir)
}

func generateModelFile(table parser.Table, outputDir, pkg string) error {
	var b strings.Builder

	tableName := table.Name
	structName := toPascalCase(tableName)
	aliasType := structName + "Alias"
	columnAliasType := structName + "ColumnAlias"
	scannerType := structName + "Scanner"
	tableType := structName + "Table"

	// Check if we need time import
	needsTime := false
	for _, col := range table.Columns {
		if col.GoType() == "time.Time" {
			needsTime = true
			break
		}
	}

	// Write package and imports
	b.WriteString(fmt.Sprintf("package %s\n\n", pkg))
	b.WriteString("import (\n")
	if needsTime {
		b.WriteString("\t\"time\"\n\n")
	}
	b.WriteString("\t\"github.com/yaroher/ratel/pkg/ddl\"\n")
	b.WriteString("\t\"github.com/yaroher/ratel/pkg/dml/set\"\n")
	b.WriteString("\t\"github.com/yaroher/ratel/pkg/schema\"\n")
	b.WriteString(")\n\n")

	// Write alias type
	b.WriteString(fmt.Sprintf("// %s is the table alias type for the %s table\n", aliasType, tableName))
	b.WriteString(fmt.Sprintf("type %s string\n\n", aliasType))
	b.WriteString(fmt.Sprintf("func (a %s) String() string { return string(a) }\n\n", aliasType))
	b.WriteString(fmt.Sprintf("const %sAliasName %s = \"%s\"\n\n", structName, aliasType, tableName))

	// Write column alias type
	b.WriteString(fmt.Sprintf("// %s represents column names for the %s table\n", columnAliasType, tableName))
	b.WriteString(fmt.Sprintf("type %s string\n\n", columnAliasType))
	b.WriteString(fmt.Sprintf("func (c %s) String() string { return string(c) }\n\n", columnAliasType))

	// Write column constants
	b.WriteString("const (\n")
	for _, col := range table.Columns {
		constName := structName + "Column" + toPascalCase(col.Name)
		b.WriteString(fmt.Sprintf("\t%s %s = \"%s\"\n", constName, columnAliasType, col.Name))
	}
	b.WriteString(")\n\n")

	// Write index constants if any
	if len(table.Indexes) > 0 {
		b.WriteString("// Index names\n")
		b.WriteString("const (\n")
		for _, idx := range table.Indexes {
			constName := structName + "Index" + toPascalCase(strings.TrimPrefix(idx.Name, "ix_"+tableName+"_"))
			b.WriteString(fmt.Sprintf("\t%s %s = \"%s\"\n", constName, aliasType, idx.Name))
		}
		b.WriteString(")\n\n")
	}

	// Write scanner struct
	b.WriteString(fmt.Sprintf("// %s is the scanner struct for %s rows\n", scannerType, tableName))
	b.WriteString(fmt.Sprintf("type %s struct {\n", scannerType))
	for _, col := range table.Columns {
		fieldName := toPascalCase(col.Name)
		goType := col.GoType()
		// Make nullable fields pointers if they're not NOT NULL and not primary key
		if !col.IsNotNull && !col.IsPrimaryKey && !strings.Contains(strings.ToUpper(col.RawType), "SERIAL") {
			if goType != "string" && goType != "[]byte" {
				goType = "*" + goType
			}
		}
		b.WriteString(fmt.Sprintf("\t%s %s\n", fieldName, goType))
	}
	b.WriteString("}\n\n")

	// Write GetTarget method
	b.WriteString(fmt.Sprintf("func (s *%s) GetTarget(col string) func() any {\n", scannerType))
	b.WriteString(fmt.Sprintf("\tswitch %s(col) {\n", columnAliasType))
	for _, col := range table.Columns {
		constName := structName + "Column" + toPascalCase(col.Name)
		fieldName := toPascalCase(col.Name)
		b.WriteString(fmt.Sprintf("\tcase %s:\n", constName))
		b.WriteString(fmt.Sprintf("\t\treturn func() any { return &s.%s }\n", fieldName))
	}
	b.WriteString("\tdefault:\n")
	b.WriteString("\t\tpanic(\"unknown column: \" + col)\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")

	// Write GetSetter method
	b.WriteString(fmt.Sprintf("func (s *%s) GetSetter(f %s) func() set.ValueSetter[%s] {\n", scannerType, columnAliasType, columnAliasType))
	b.WriteString("\tswitch f {\n")
	for _, col := range table.Columns {
		constName := structName + "Column" + toPascalCase(col.Name)
		fieldName := toPascalCase(col.Name)
		b.WriteString(fmt.Sprintf("\tcase %s:\n", constName))
		b.WriteString(fmt.Sprintf("\t\treturn func() set.ValueSetter[%s] { return set.NewSetter(f, &s.%s) }\n", columnAliasType, fieldName))
	}
	b.WriteString("\tdefault:\n")
	b.WriteString(fmt.Sprintf("\t\tpanic(\"unknown column: \" + string(f))\n"))
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")

	// Write GetValue method
	b.WriteString(fmt.Sprintf("func (s *%s) GetValue(f %s) func() any {\n", scannerType, columnAliasType))
	b.WriteString("\tswitch f {\n")
	for _, col := range table.Columns {
		constName := structName + "Column" + toPascalCase(col.Name)
		fieldName := toPascalCase(col.Name)
		b.WriteString(fmt.Sprintf("\tcase %s:\n", constName))
		b.WriteString(fmt.Sprintf("\t\treturn func() any { return s.%s }\n", fieldName))
	}
	b.WriteString("\tdefault:\n")
	b.WriteString(fmt.Sprintf("\t\tpanic(\"unknown column: \" + string(f))\n"))
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")

	// Write table struct
	b.WriteString(fmt.Sprintf("// %s represents the %s table with its columns\n", tableType, tableName))
	b.WriteString(fmt.Sprintf("type %s struct {\n", tableType))
	b.WriteString(fmt.Sprintf("\t*schema.Table[%s, %s, *%s]\n", aliasType, columnAliasType, scannerType))
	for _, col := range table.Columns {
		fieldName := toPascalCase(col.Name)
		colInterface := getColumnInterface(col)
		b.WriteString(fmt.Sprintf("\t%s %s[%s]\n", fieldName, colInterface, columnAliasType))
	}
	b.WriteString("}\n\n")

	// Write table variable
	b.WriteString(fmt.Sprintf("// %s is the global %s table instance\n", structName, tableName))
	b.WriteString(fmt.Sprintf("var %s = func() %s {\n", structName, tableType))

	// Create column variables
	for _, col := range table.Columns {
		varName := toCamelCase(col.Name) + "Col"
		b.WriteString(fmt.Sprintf("\t%s := schema.%s(%s%s", varName, col.DDLType(), structName+"Column"+toPascalCase(col.Name), getColumnArgs(col)))

		// Add options
		options := getColumnOptions(col, structName, columnAliasType)
		if len(options) > 0 {
			b.WriteString(",\n")
			for i, opt := range options {
				b.WriteString(fmt.Sprintf("\t\t%s", opt))
				if i < len(options)-1 {
					b.WriteString(",\n")
				}
			}
		}
		b.WriteString(")\n")
	}

	// Return table struct
	b.WriteString(fmt.Sprintf("\n\treturn %s{\n", tableType))
	b.WriteString(fmt.Sprintf("\t\tTable: schema.NewTable[%s, %s, *%s](\n", aliasType, columnAliasType, scannerType))
	b.WriteString(fmt.Sprintf("\t\t\t%sAliasName,\n", structName))
	b.WriteString(fmt.Sprintf("\t\t\tfunc() *%s { return &%s{} },\n", scannerType, scannerType))
	b.WriteString(fmt.Sprintf("\t\t\t[]*ddl.ColumnDDL[%s]{\n", columnAliasType))
	for _, col := range table.Columns {
		varName := toCamelCase(col.Name) + "Col"
		b.WriteString(fmt.Sprintf("\t\t\t\t%s.DDL(),\n", varName))
	}
	b.WriteString("\t\t\t},\n")

	// Add table options
	tableOptions := getTableOptions(table, structName, aliasType, columnAliasType)
	for _, opt := range tableOptions {
		b.WriteString(fmt.Sprintf("\t\t\t%s,\n", opt))
	}

	b.WriteString("\t\t),\n")

	// Assign columns
	for _, col := range table.Columns {
		fieldName := toPascalCase(col.Name)
		varName := toCamelCase(col.Name) + "Col"
		b.WriteString(fmt.Sprintf("\t\t%s: %s,\n", fieldName, varName))
	}
	b.WriteString("\t}\n")
	b.WriteString("}()\n\n")

	// Write ref variable
	b.WriteString(fmt.Sprintf("// %sRef is a reference to the %s table for relations\n", structName, tableName))
	b.WriteString(fmt.Sprintf("var %sRef schema.RelationTableAlias[%s] = %s.Table\n", structName, aliasType, structName))

	// Write to file
	filename := filepath.Join(outputDir, tableName+".go")
	return os.WriteFile(filename, []byte(b.String()), 0644)
}

func generateRelationsFile(tables []parser.Table, outputDir, pkg string) error {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", pkg))
	b.WriteString("import (\n")
	b.WriteString("\t\"github.com/yaroher/ratel/pkg/schema\"\n")
	b.WriteString(")\n\n")

	b.WriteString("// Relations file - define your table relations here\n")
	b.WriteString("// Example:\n")
	b.WriteString("//\n")
	b.WriteString("// var UsersOrders = schema.HasMany[UsersAlias, UsersColumnAlias, *UsersScanner, OrdersAlias, OrdersColumnAlias, *OrdersScanner](\n")
	b.WriteString("//     UsersAliasName,\n")
	b.WriteString("//     OrdersRef,\n")
	b.WriteString("//     OrdersColumnUserID,\n")
	b.WriteString("//     UsersColumnUserID,\n")
	b.WriteString("// )\n\n")

	// Write compile-time interface checks
	b.WriteString("// Compile-time interface checks\n")
	b.WriteString("var (\n")
	for _, table := range tables {
		structName := toPascalCase(table.Name)
		aliasType := structName + "Alias"
		b.WriteString(fmt.Sprintf("\t_ schema.RelationTableAlias[%s] = %s.Table\n", aliasType, structName))
	}
	b.WriteString(")\n")

	filename := filepath.Join(outputDir, "relations.go")
	return os.WriteFile(filename, []byte(b.String()), 0644)
}

func getColumnInterface(col parser.Column) string {
	upper := strings.ToUpper(col.RawType)
	baseType := upper
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}

	switch baseType {
	case "BIGSERIAL":
		return "schema.BigSerialColumnI"
	case "SERIAL":
		return "schema.SerialColumnI"
	case "SMALLSERIAL":
		return "schema.SmallSerialColumnI"
	case "BIGINT", "INT8":
		return "schema.BigIntColumnI"
	case "INTEGER", "INT", "INT4":
		return "schema.IntegerColumnI"
	case "SMALLINT", "INT2":
		return "schema.SmallIntColumnI"
	case "BOOLEAN", "BOOL":
		return "schema.BooleanColumnI"
	case "TEXT":
		return "schema.TextColumnI"
	case "VARCHAR", "CHARACTER":
		return "schema.VarcharColumnI"
	case "CHAR":
		return "schema.CharColumnI"
	case "NUMERIC", "DECIMAL":
		return "schema.NumericColumnI"
	case "REAL", "FLOAT4":
		return "schema.RealColumnI"
	case "DOUBLE", "FLOAT8", "FLOAT":
		return "schema.DoubleColumnI"
	case "TIMESTAMPTZ":
		return "schema.TimestamptzColumnI"
	case "TIMESTAMP":
		return "schema.TimestampColumnI"
	case "DATE":
		return "schema.DateColumnI"
	case "TIME":
		return "schema.TimeColumnI"
	case "TIMETZ":
		return "schema.TimetzColumnI"
	case "UUID":
		return "schema.UUIDColumnI"
	case "BYTEA":
		return "schema.ByteaColumnI"
	case "JSON":
		return "schema.JSONColumnI"
	case "JSONB":
		return "schema.JSONBColumnI"
	default:
		return "schema.TextColumnI"
	}
}

func getColumnArgs(col parser.Column) string {
	upper := strings.ToUpper(col.RawType)
	baseType := upper
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}

	p1, p2 := col.TypePrecision()

	switch baseType {
	case "CHAR", "VARCHAR", "CHARACTER":
		if p1 > 0 {
			return fmt.Sprintf(", %d", p1)
		}
	case "NUMERIC", "DECIMAL":
		if p1 > 0 {
			return fmt.Sprintf(", %d, %d", p1, p2)
		}
	}

	return ""
}

func getColumnOptions(col parser.Column, structName, columnAliasType string) []string {
	var options []string

	if col.IsPrimaryKey {
		options = append(options, fmt.Sprintf("ddl.WithPrimaryKey[%s]()", columnAliasType))
	}
	if col.IsNotNull && !col.IsPrimaryKey && !strings.Contains(strings.ToUpper(col.RawType), "SERIAL") {
		options = append(options, fmt.Sprintf("ddl.WithNotNull[%s]()", columnAliasType))
	}
	if col.IsUnique {
		options = append(options, fmt.Sprintf("ddl.WithUnique[%s]()", columnAliasType))
	}
	if col.Default != "" {
		options = append(options, fmt.Sprintf("ddl.WithDefault[%s](\"%s\")", columnAliasType, col.Default))
	}
	if col.Check != "" {
		options = append(options, fmt.Sprintf("ddl.WithCheck[%s](\"%s\")", columnAliasType, col.Check))
	}
	if col.References != nil {
		options = append(options, fmt.Sprintf("ddl.WithReferences[%s](\"%s\", \"%s\")", columnAliasType, col.References.Table, col.References.Column))
	}
	if col.OnDelete != "" {
		options = append(options, fmt.Sprintf("ddl.WithOnDelete[%s](\"%s\")", columnAliasType, col.OnDelete))
	}

	return options
}

func getTableOptions(table parser.Table, structName, aliasType, columnAliasType string) []string {
	var options []string

	// Composite primary key
	if len(table.PrimaryKey) > 0 {
		cols := make([]string, len(table.PrimaryKey))
		for i, pk := range table.PrimaryKey {
			cols[i] = structName + "Column" + toPascalCase(pk)
		}
		options = append(options, fmt.Sprintf("ddl.WithPrimaryKeyColumns[%s, %s]([]%s{%s})",
			aliasType, columnAliasType, columnAliasType, strings.Join(cols, ", ")))
	}

	// Unique constraints
	for _, uc := range table.UniqueConstraints {
		cols := make([]string, len(uc))
		for i, c := range uc {
			cols[i] = structName + "Column" + toPascalCase(c)
		}
		options = append(options, fmt.Sprintf("ddl.WithUniqueColumns[%s, %s]([]%s{%s})",
			aliasType, columnAliasType, columnAliasType, strings.Join(cols, ", ")))
	}

	// Indexes
	if len(table.Indexes) > 0 {
		var indexDefs []string
		for _, idx := range table.Indexes {
			cols := make([]string, len(idx.Columns))
			for i, c := range idx.Columns {
				cols[i] = structName + "Column" + toPascalCase(strings.TrimSuffix(strings.TrimSuffix(c, " DESC"), " ASC"))
			}

			indexDef := fmt.Sprintf("ddl.NewIndex[%s, %s](\"%s\", %sAliasName).OnColumns(%s)",
				aliasType, columnAliasType, idx.Name, structName, strings.Join(cols, ", "))

			if idx.Where != "" {
				indexDef += fmt.Sprintf(".Where(\"%s\")", idx.Where)
			}
			indexDefs = append(indexDefs, indexDef)
		}
		options = append(options, fmt.Sprintf("ddl.WithIndexes[%s, %s](\n\t\t\t\t%s,\n\t\t\t)",
			aliasType, columnAliasType, strings.Join(indexDefs, ",\n\t\t\t\t")))
	}

	return options
}

func toPascalCase(s string) string {
	return strcase.ToCamel(s)
}

func toCamelCase(s string) string {
	return strcase.ToLowerCamel(s)
}
