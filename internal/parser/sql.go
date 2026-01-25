package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Column represents a parsed column definition
type Column struct {
	Name         string
	Type         string
	RawType      string
	IsNotNull    bool
	IsPrimaryKey bool
	IsUnique     bool
	Default      string
	Check        string
	References   *Reference
	OnDelete     string
}

// Reference represents a foreign key reference
type Reference struct {
	Table  string
	Column string
}

// Index represents a parsed index definition
type Index struct {
	Name    string
	Table   string
	Columns []string
	Where   string
	Unique  bool
}

// Table represents a parsed table definition
type Table struct {
	Name              string
	Columns           []Column
	PrimaryKey        []string // composite primary key
	UniqueConstraints [][]string
	Indexes           []Index
}

// Schema represents the entire parsed schema
type Schema struct {
	Tables  []Table
	Indexes []Index
}

// ParseSQL parses a SQL schema string and returns the parsed schema
func ParseSQL(sql string) (*Schema, error) {
	schema := &Schema{}

	// Remove comments
	sql = removeComments(sql)

	// Split into statements
	statements := splitStatements(sql)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		upperStmt := strings.ToUpper(stmt)

		if strings.HasPrefix(upperStmt, "CREATE TABLE") {
			table, err := parseCreateTable(stmt)
			if err != nil {
				return nil, err
			}
			schema.Tables = append(schema.Tables, *table)
		} else if strings.HasPrefix(upperStmt, "CREATE INDEX") || strings.HasPrefix(upperStmt, "CREATE UNIQUE INDEX") {
			index, err := parseCreateIndex(stmt)
			if err != nil {
				return nil, err
			}
			schema.Indexes = append(schema.Indexes, *index)
		}
	}

	// Associate indexes with tables
	for i := range schema.Indexes {
		for j := range schema.Tables {
			if schema.Tables[j].Name == schema.Indexes[i].Table {
				schema.Tables[j].Indexes = append(schema.Tables[j].Indexes, schema.Indexes[i])
			}
		}
	}

	return schema, nil
}

func removeComments(sql string) string {
	// Remove single-line comments
	re := regexp.MustCompile(`--[^\n]*`)
	sql = re.ReplaceAllString(sql, "")

	// Remove multi-line comments
	re = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	sql = re.ReplaceAllString(sql, "")

	return sql
}

func splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	depth := 0

	for _, ch := range sql {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
		}

		if ch == ';' && depth == 0 {
			statements = append(statements, current.String())
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

func parseCreateTable(stmt string) (*Table, error) {
	table := &Table{}

	// Extract table name
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s*\(`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) < 2 {
		return nil, nil
	}
	table.Name = matches[1]

	// Extract content between outer parentheses
	start := strings.Index(stmt, "(")
	end := strings.LastIndex(stmt, ")")
	if start == -1 || end == -1 {
		return nil, nil
	}
	content := stmt[start+1 : end]

	// Split by commas (but not inside parentheses)
	parts := splitByComma(content)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		upperPart := strings.ToUpper(part)

		if strings.HasPrefix(upperPart, "PRIMARY KEY") {
			// Composite primary key
			table.PrimaryKey = extractColumnList(part)
		} else if strings.HasPrefix(upperPart, "UNIQUE") && !strings.Contains(upperPart, " KEY ") {
			// Table-level unique constraint
			cols := extractColumnList(part)
			if len(cols) > 0 {
				table.UniqueConstraints = append(table.UniqueConstraints, cols)
			}
		} else if strings.HasPrefix(upperPart, "FOREIGN KEY") ||
			strings.HasPrefix(upperPart, "CONSTRAINT") ||
			strings.HasPrefix(upperPart, "CHECK") {
			// Skip these for now
			continue
		} else {
			// Column definition
			col := parseColumnDefinition(part)
			if col != nil {
				table.Columns = append(table.Columns, *col)
			}
		}
	}

	return table, nil
}

func splitByComma(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
		}

		if ch == ',' && depth == 0 {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func extractColumnList(s string) []string {
	// Find content in parentheses
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start == -1 || end == -1 {
		return nil
	}
	content := s[start+1 : end]

	// Split by comma and trim
	parts := strings.Split(content, ",")
	var cols []string
	for _, p := range parts {
		col := strings.TrimSpace(p)
		if col != "" {
			cols = append(cols, col)
		}
	}
	return cols
}

func parseColumnDefinition(def string) *Column {
	def = strings.TrimSpace(def)
	if def == "" {
		return nil
	}

	// Split into tokens
	tokens := tokenize(def)
	if len(tokens) < 2 {
		return nil
	}

	col := &Column{
		Name:    tokens[0],
		RawType: tokens[1],
	}

	// Parse type with precision
	col.Type = normalizeType(col.RawType, tokens[1:])

	// Parse constraints
	upper := strings.ToUpper(def)

	col.IsNotNull = strings.Contains(upper, "NOT NULL")
	col.IsPrimaryKey = strings.Contains(upper, "PRIMARY KEY")
	col.IsUnique = strings.Contains(upper, " UNIQUE")

	// Extract DEFAULT
	if idx := strings.Index(upper, "DEFAULT "); idx != -1 {
		rest := def[idx+8:]
		col.Default = extractDefaultValue(rest)
	}

	// Extract CHECK
	if idx := strings.Index(upper, "CHECK "); idx != -1 {
		rest := def[idx+6:]
		col.Check = extractParenContent(rest)
	}

	// Extract REFERENCES
	if idx := strings.Index(upper, "REFERENCES "); idx != -1 {
		rest := def[idx+11:]
		col.References = parseReference(rest)
	}

	// Extract ON DELETE
	if idx := strings.Index(upper, "ON DELETE "); idx != -1 {
		rest := upper[idx+10:]
		parts := strings.Fields(rest)
		if len(parts) > 0 {
			col.OnDelete = parts[0]
		}
	}

	return col
}

func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inParen := 0

	for _, ch := range s {
		if ch == '(' {
			inParen++
			current.WriteRune(ch)
		} else if ch == ')' {
			inParen--
			current.WriteRune(ch)
		} else if (ch == ' ' || ch == '\t' || ch == '\n') && inParen == 0 {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func normalizeType(rawType string, tokens []string) string {
	upper := strings.ToUpper(rawType)

	// Handle types with precision like NUMERIC(12, 2) or CHAR(3)
	if strings.Contains(rawType, "(") {
		return rawType
	}

	// Handle multi-word types
	if len(tokens) > 1 {
		nextUpper := strings.ToUpper(tokens[1])
		if nextUpper == "PRECISION" || nextUpper == "VARYING" || nextUpper == "WITHOUT" || nextUpper == "WITH" {
			return rawType + " " + tokens[1]
		}
	}

	return upper
}

func extractDefaultValue(s string) string {
	s = strings.TrimSpace(s)

	// Check for function call like now()
	if strings.HasPrefix(strings.ToLower(s), "now()") {
		return "now()"
	}

	// Check for boolean
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "true") {
		return "true"
	}
	if strings.HasPrefix(lower, "false") {
		return "false"
	}

	// Check for quoted string
	if strings.HasPrefix(s, "'") {
		end := strings.Index(s[1:], "'")
		if end != -1 {
			return s[:end+2]
		}
	}

	// Extract until next keyword or end
	tokens := strings.Fields(s)
	if len(tokens) > 0 {
		return tokens[0]
	}

	return s
}

func extractParenContent(s string) string {
	start := strings.Index(s, "(")
	if start == -1 {
		return ""
	}

	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '(' {
			depth++
		} else if s[i] == ')' {
			depth--
			if depth == 0 {
				return s[start+1 : i]
			}
		}
	}
	return ""
}

func parseReference(s string) *Reference {
	s = strings.TrimSpace(s)

	// Extract table name
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}

	tablePart := parts[0]

	// Check if column is in parentheses
	if idx := strings.Index(tablePart, "("); idx != -1 {
		table := tablePart[:idx]
		col := extractParenContent(tablePart)
		return &Reference{Table: table, Column: col}
	}

	// Column might be in next token
	table := tablePart
	var col string
	if len(parts) > 1 && strings.HasPrefix(parts[1], "(") {
		col = extractParenContent(parts[1])
	}

	return &Reference{Table: table, Column: col}
}

func parseCreateIndex(stmt string) (*Index, error) {
	index := &Index{}

	upper := strings.ToUpper(stmt)
	index.Unique = strings.Contains(upper, "UNIQUE INDEX")

	// Extract index name
	re := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s+ON\s+(\w+)`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) < 3 {
		return nil, nil
	}
	index.Name = matches[1]
	index.Table = matches[2]

	// Extract columns
	start := strings.Index(stmt, "(")
	if start != -1 {
		// Find the matching closing parenthesis
		depth := 0
		for i := start; i < len(stmt); i++ {
			if stmt[i] == '(' {
				depth++
			} else if stmt[i] == ')' {
				depth--
				if depth == 0 {
					content := stmt[start+1 : i]
					parts := strings.Split(content, ",")
					for _, p := range parts {
						col := strings.TrimSpace(p)
						if col != "" {
							index.Columns = append(index.Columns, col)
						}
					}
					// Check for WHERE clause
					rest := stmt[i+1:]
					if idx := strings.Index(strings.ToUpper(rest), "WHERE"); idx != -1 {
						index.Where = strings.TrimSpace(rest[idx+5:])
					}
					break
				}
			}
		}
	}

	return index, nil
}

// GoType returns the Go type for a SQL type
func (c *Column) GoType() string {
	upper := strings.ToUpper(c.RawType)

	// Remove precision/scale for matching
	baseType := upper
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}

	switch baseType {
	case "BIGSERIAL", "BIGINT", "INT8":
		return "int64"
	case "SERIAL", "INTEGER", "INT", "INT4":
		return "int32"
	case "SMALLSERIAL", "SMALLINT", "INT2":
		return "int16"
	case "BOOLEAN", "BOOL":
		return "bool"
	case "TEXT", "VARCHAR", "CHAR", "CHARACTER":
		return "string"
	case "NUMERIC", "DECIMAL", "REAL", "DOUBLE", "FLOAT", "FLOAT4", "FLOAT8":
		return "float64"
	case "TIMESTAMPTZ", "TIMESTAMP":
		return "time.Time"
	case "DATE":
		return "time.Time"
	case "TIME", "TIMETZ":
		return "time.Time"
	case "UUID":
		return "string"
	case "BYTEA":
		return "[]byte"
	case "JSON", "JSONB":
		return "json.RawMessage"
	default:
		return "any"
	}
}

// DDLType returns the DDL column type function name
func (c *Column) DDLType() string {
	upper := strings.ToUpper(c.RawType)
	baseType := upper
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}

	switch baseType {
	case "BIGSERIAL":
		return "BigSerialColumn"
	case "SERIAL":
		return "SerialColumn"
	case "SMALLSERIAL":
		return "SmallSerialColumn"
	case "BIGINT", "INT8":
		return "BigIntColumn"
	case "INTEGER", "INT", "INT4":
		return "IntegerColumn"
	case "SMALLINT", "INT2":
		return "SmallIntColumn"
	case "BOOLEAN", "BOOL":
		return "BooleanColumn"
	case "TEXT":
		return "TextColumn"
	case "VARCHAR", "CHARACTER":
		return "VarcharColumn"
	case "CHAR":
		return "CharColumn"
	case "NUMERIC", "DECIMAL":
		return "NumericColumn"
	case "REAL", "FLOAT4":
		return "RealColumn"
	case "DOUBLE", "FLOAT8", "FLOAT":
		return "DoubleColumn"
	case "TIMESTAMPTZ":
		return "TimestamptzColumn"
	case "TIMESTAMP":
		return "TimestampColumn"
	case "DATE":
		return "DateColumn"
	case "TIME":
		return "TimeColumn"
	case "TIMETZ":
		return "TimetzColumn"
	case "UUID":
		return "UUIDColumn"
	case "BYTEA":
		return "ByteaColumn"
	case "JSON":
		return "JSONColumn"
	case "JSONB":
		return "JSONBColumn"
	default:
		return "TextColumn"
	}
}

// TypePrecision returns precision parameters for types like CHAR(3) or NUMERIC(12,2)
func (c *Column) TypePrecision() (int, int) {
	raw := c.RawType
	start := strings.Index(raw, "(")
	end := strings.Index(raw, ")")
	if start == -1 || end == -1 {
		return 0, 0
	}

	content := raw[start+1 : end]
	parts := strings.Split(content, ",")

	var p1, p2 int
	if len(parts) >= 1 {
		fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &p1)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &p2)
	}

	return p1, p2
}
