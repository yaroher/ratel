package ddl

// RawSQL wraps a raw SQL statement and implements SchemaSqler.
// Use this to pass arbitrary SQL (extensions, functions, grants)
// to SchemaSQL alongside table definitions.
type RawSQL string

func (r RawSQL) SchemaSql() []string { return []string{string(r)} }
