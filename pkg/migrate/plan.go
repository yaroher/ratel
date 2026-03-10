package migrate

// Plan is a generated migration.
type Plan struct {
	Name    string
	Changes []PlannedChange
}

// PlannedChange is a single DDL statement in a migration.
type PlannedChange struct {
	SQL     string // DDL statement
	Comment string // Human-readable description
	Reverse string // Rollback statement (if possible)
}
