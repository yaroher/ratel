package postgres

// Based on Atlas (https://github.com/ariga/atlas) — Apache 2.0 License
// Copyright 2021-present The Atlas Authors

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yaroher/ratel/pkg/migrate"
)

// Differ compares two SchemaState values and returns the changes needed
// to transition from current to desired.
type Differ struct{}

// Diff returns the changes needed to transition from current to desired.
func (d *Differ) Diff(current, desired *migrate.SchemaState) ([]migrate.Change, error) {
	var changes []migrate.Change

	currentSchemas := schemaMap(current.Schemas)
	desiredSchemas := schemaMap(desired.Schemas)

	// Added schemas — collect all new schemas first, then emit CREATE SCHEMA
	// before any tables so that cross-schema FK references resolve correctly.
	var addedSchemaNames []string
	addedSchemaMap := make(map[string]migrate.Schema)
	var allAddedTables []migrate.Table
	for name, s := range desiredSchemas {
		if _, ok := currentSchemas[name]; !ok {
			addedSchemaNames = append(addedSchemaNames, name)
			addedSchemaMap[name] = s
			allAddedTables = append(allAddedTables, s.Tables...)
		}
	}
	sort.Strings(addedSchemaNames)

	// Emit all CREATE SCHEMA statements first.
	for _, name := range addedSchemaNames {
		sc := addedSchemaMap[name]
		changes = append(changes, migrate.AddSchema{S: &sc})
	}

	// Emit extensions and functions for new schemas.
	for _, name := range addedSchemaNames {
		sc := addedSchemaMap[name]
		changes = append(changes, d.diffExtensions(nil, sc.Extensions)...)
		changes = append(changes, d.diffFunctions(nil, sc.Functions)...)
	}

	// Sort all added tables across schemas by FK dependencies, then emit.
	sort.Slice(allAddedTables, func(i, j int) bool {
		return qualifiedTableKey(allAddedTables[i]) < qualifiedTableKey(allAddedTables[j])
	})
	sortTablesByFK(allAddedTables)
	for i := range allAddedTables {
		tc := allAddedTables[i]
		changes = append(changes, migrate.AddTable{T: &tc})
	}

	// Dropped schemas — emit content drops first, then DropSchema
	var droppedSchemaNames []string
	for name := range currentSchemas {
		if _, ok := desiredSchemas[name]; !ok {
			droppedSchemaNames = append(droppedSchemaNames, name)
		}
	}
	sort.Strings(droppedSchemaNames)
	for _, name := range droppedSchemaNames {
		sc := currentSchemas[name]
		// Drop contents before dropping the schema itself
		changes = append(changes, d.diffFunctions(sc.Functions, nil)...)
		changes = append(changes, d.diffExtensions(sc.Extensions, nil)...)
		changes = append(changes, d.diffTables(sc.Tables, nil)...)
		changes = append(changes, migrate.DropSchema{S: &sc})
	}

	// Modified schemas — diff tables, extensions, functions within
	var modifiedSchemaNames []string
	for name := range currentSchemas {
		if _, ok := desiredSchemas[name]; ok {
			modifiedSchemaNames = append(modifiedSchemaNames, name)
		}
	}
	sort.Strings(modifiedSchemaNames)
	for _, name := range modifiedSchemaNames {
		curSchema := currentSchemas[name]
		desSchema := desiredSchemas[name]
		tableChanges := d.diffTables(curSchema.Tables, desSchema.Tables)
		changes = append(changes, tableChanges...)
		extChanges := d.diffExtensions(curSchema.Extensions, desSchema.Extensions)
		changes = append(changes, extChanges...)
		funcChanges := d.diffFunctions(curSchema.Functions, desSchema.Functions)
		changes = append(changes, funcChanges...)
	}

	return changes, nil
}

// ---- schema helpers ----

func schemaMap(schemas []migrate.Schema) map[string]migrate.Schema {
	m := make(map[string]migrate.Schema, len(schemas))
	for _, s := range schemas {
		m[s.Name] = s
	}
	return m
}

// ---- table diffing ----

func (d *Differ) diffTables(current, desired []migrate.Table) []migrate.Change {
	var changes []migrate.Change

	curMap := tableMap(current)
	desMap := tableMap(desired)

	// Added tables — collect, sort alphabetically for determinism, then by FK deps.
	var added []migrate.Table
	for name, t := range desMap {
		if _, ok := curMap[name]; !ok {
			added = append(added, t)
		}
	}
	sort.Slice(added, func(i, j int) bool { return added[i].Name < added[j].Name })
	sortTablesByFK(added)
	for i := range added {
		tc := added[i]
		changes = append(changes, migrate.AddTable{T: &tc})
	}

	// Dropped tables (reverse dependency order — drop dependents first)
	var dropped []migrate.Table
	for name, t := range curMap {
		if _, ok := desMap[name]; !ok {
			dropped = append(dropped, t)
		}
	}
	sort.Slice(dropped, func(i, j int) bool { return dropped[i].Name < dropped[j].Name })
	sortTablesByFK(dropped)
	for i := len(dropped) - 1; i >= 0; i-- {
		tc := dropped[i]
		changes = append(changes, migrate.DropTable{T: &tc})
	}

	// Modified tables
	for name, curTable := range curMap {
		desTable, ok := desMap[name]
		if !ok {
			continue
		}
		subChanges := d.diffTableContents(curTable, desTable)
		if len(subChanges) > 0 {
			ct := curTable
			dt := desTable
			changes = append(changes, migrate.ModifyTable{
				From:    &ct,
				To:      &dt,
				Changes: subChanges,
			})
		}
	}

	return changes
}

// sortTablesByFK sorts tables in topological order based on foreign key
// dependencies (Kahn's algorithm). Tables that are referenced by others come
// first so that CREATE TABLE statements execute in a valid order.
// Uses schema-qualified names so cross-schema FK references are handled correctly.
func sortTablesByFK(tables []migrate.Table) {
	if len(tables) <= 1 {
		return
	}

	// Build qualified name → index mapping for cross-schema support.
	idx := make(map[string]int, len(tables))
	for i, t := range tables {
		idx[qualifiedTableKey(t)] = i
	}

	// Calculate in-degree: how many tables within this set reference each table.
	inDeg := make([]int, len(tables))
	deps := make(map[int][]int) // dependency index → dependent indexes
	for i, t := range tables {
		for _, fk := range t.ForeignKeys {
			refKey := fkRefKey(t.Schema, fk)
			if j, ok := idx[refKey]; ok && j != i {
				inDeg[i]++
				deps[j] = append(deps[j], i)
			}
		}
	}

	// Kahn's algorithm.
	var queue []int
	for i, d := range inDeg {
		if d == 0 {
			queue = append(queue, i)
		}
	}

	sorted := make([]migrate.Table, 0, len(tables))
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		sorted = append(sorted, tables[cur])
		for _, dep := range deps[cur] {
			inDeg[dep]--
			if inDeg[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// If there's a cycle, append remaining tables in original order.
	if len(sorted) < len(tables) {
		for i, t := range tables {
			if inDeg[i] > 0 {
				sorted = append(sorted, t)
			}
		}
	}

	copy(tables, sorted)
}

// qualifiedTableKey returns "schema.name" or just "name" if schema is empty.
func qualifiedTableKey(t migrate.Table) string {
	if t.Schema == "" {
		return t.Name
	}
	return t.Schema + "." + t.Name
}

// fkRefKey returns the qualified lookup key for a FK reference.
// If RefSchema is empty, it falls back to the owning table's schema.
func fkRefKey(tableSchema string, fk migrate.ForeignKey) string {
	schema := fk.RefSchema
	if schema == "" {
		schema = tableSchema
	}
	if schema == "" {
		return fk.RefTable
	}
	return schema + "." + fk.RefTable
}

func tableMap(tables []migrate.Table) map[string]migrate.Table {
	m := make(map[string]migrate.Table, len(tables))
	for _, t := range tables {
		m[t.Name] = t
	}
	return m
}

func (d *Differ) diffTableContents(cur, des migrate.Table) []migrate.Change {
	changes := make([]migrate.Change, 0, 8)

	changes = append(changes, diffColumns(cur.Columns, des.Columns)...)
	changes = append(changes, diffIndexes(cur.Indexes, des.Indexes)...)
	changes = append(changes, diffForeignKeys(cur.ForeignKeys, des.ForeignKeys)...)
	changes = append(changes, diffChecks(cur.Checks, des.Checks)...)
	changes = append(changes, diffRLS(cur, des)...)
	changes = append(changes, diffPolicies(cur.Name, cur.Policies, des.Policies)...)
	changes = append(changes, diffTriggers(cur.Name, cur.Triggers, des.Triggers)...)

	return changes
}

// ---- column diffing ----

func diffColumns(current, desired []migrate.Column) []migrate.Change {
	var changes []migrate.Change

	curMap := columnMap(current)
	desMap := columnMap(desired)

	for name, dc := range desMap {
		if _, ok := curMap[name]; !ok {
			c := dc
			changes = append(changes, migrate.AddColumn{C: &c})
		}
	}

	for name, cc := range curMap {
		if _, ok := desMap[name]; !ok {
			c := cc
			changes = append(changes, migrate.DropColumn{C: &c})
		}
	}

	for name, cc := range curMap {
		dc, ok := desMap[name]
		if !ok {
			continue
		}
		if columnModified(cc, dc) {
			from := cc
			to := dc
			changes = append(changes, migrate.ModifyColumn{From: &from, To: &to})
		}
	}

	return changes
}

func columnMap(cols []migrate.Column) map[string]migrate.Column {
	m := make(map[string]migrate.Column, len(cols))
	for _, c := range cols {
		m[c.Name] = c
	}
	return m
}

func columnModified(a, b migrate.Column) bool {
	return a.Type != b.Type || a.Nullable != b.Nullable || a.Default != b.Default
}

// ---- index diffing ----

func diffIndexes(current, desired []migrate.Index) []migrate.Change {
	var changes []migrate.Change

	curMap := indexMap(current)
	desMap := indexMap(desired)

	for name, di := range desMap {
		if _, ok := curMap[name]; !ok {
			i := di
			changes = append(changes, migrate.AddIndex{I: &i})
		}
	}

	for name, ci := range curMap {
		if _, ok := desMap[name]; !ok {
			i := ci
			changes = append(changes, migrate.DropIndex{I: &i})
		}
	}

	for name, ci := range curMap {
		di, ok := desMap[name]
		if !ok {
			continue
		}
		if indexModified(ci, di) {
			from := ci
			to := di
			changes = append(changes, migrate.ModifyIndex{From: &from, To: &to})
		}
	}

	return changes
}

func indexMap(indexes []migrate.Index) map[string]migrate.Index {
	m := make(map[string]migrate.Index, len(indexes))
	for _, i := range indexes {
		m[i.Name] = i
	}
	return m
}

func indexModified(a, b migrate.Index) bool {
	if a.Unique != b.Unique || a.Method != b.Method || a.Where != b.Where {
		return true
	}
	if !stringSliceEqual(a.Include, b.Include) {
		return true
	}
	if len(a.Columns) != len(b.Columns) {
		return true
	}
	for i := range a.Columns {
		if a.Columns[i].Name != b.Columns[i].Name || a.Columns[i].Desc != b.Columns[i].Desc {
			return true
		}
	}
	return false
}

// ---- foreign key diffing ----

func diffForeignKeys(current, desired []migrate.ForeignKey) []migrate.Change {
	var changes []migrate.Change

	curMap := fkMap(current)
	desMap := fkMap(desired)

	for name, dfk := range desMap {
		if _, ok := curMap[name]; !ok {
			fk := dfk
			changes = append(changes, migrate.AddForeignKey{FK: &fk})
		}
	}

	for name, cfk := range curMap {
		if _, ok := desMap[name]; !ok {
			fk := cfk
			changes = append(changes, migrate.DropForeignKey{FK: &fk})
		}
	}

	return changes
}

func fkMap(fks []migrate.ForeignKey) map[string]migrate.ForeignKey {
	m := make(map[string]migrate.ForeignKey, len(fks))
	for _, fk := range fks {
		m[fk.Name] = fk
	}
	return m
}

// ---- check constraint diffing ----

func diffChecks(current, desired []migrate.Check) []migrate.Change {
	var changes []migrate.Change

	curMap := checkMap(current)
	desMap := checkMap(desired)

	for name, dc := range desMap {
		if _, ok := curMap[name]; !ok {
			c := dc
			changes = append(changes, migrate.AddCheck{C: &c})
		}
	}

	for name, cc := range curMap {
		if _, ok := desMap[name]; !ok {
			c := cc
			changes = append(changes, migrate.DropCheck{C: &c})
		}
	}

	return changes
}

func checkMap(checks []migrate.Check) map[string]migrate.Check {
	m := make(map[string]migrate.Check, len(checks))
	for _, c := range checks {
		m[c.Name] = c
	}
	return m
}

// ---- RLS diffing ----

func diffRLS(cur, des migrate.Table) []migrate.Change {
	var changes []migrate.Change

	if !cur.RLSEnabled && des.RLSEnabled {
		changes = append(changes, migrate.EnableRLS{Table: des.Name})
	} else if cur.RLSEnabled && !des.RLSEnabled {
		changes = append(changes, migrate.DisableRLS{Table: cur.Name})
	}

	if !cur.RLSForced && des.RLSForced {
		changes = append(changes, migrate.ForceRLS{Table: des.Name})
	} else if cur.RLSForced && !des.RLSForced {
		changes = append(changes, migrate.UnforceRLS{Table: cur.Name})
	}

	return changes
}

// ---- policy diffing ----

func diffPolicies(tableName string, current, desired []migrate.Policy) []migrate.Change {
	var changes []migrate.Change

	curMap := policyMap(current)
	desMap := policyMap(desired)

	for name, dp := range desMap {
		if _, ok := curMap[name]; !ok {
			p := dp
			changes = append(changes, migrate.AddPolicy{Table: tableName, P: &p})
		}
	}

	for name, cp := range curMap {
		if _, ok := desMap[name]; !ok {
			p := cp
			changes = append(changes, migrate.DropPolicy{Table: tableName, P: &p})
		}
	}

	for name, cp := range curMap {
		dp, ok := desMap[name]
		if !ok {
			continue
		}
		if policyModified(cp, dp) {
			from := cp
			to := dp
			changes = append(changes, migrate.ModifyPolicy{Table: tableName, From: &from, To: &to})
		}
	}

	return changes
}

func policyMap(policies []migrate.Policy) map[string]migrate.Policy {
	m := make(map[string]migrate.Policy, len(policies))
	for _, p := range policies {
		m[p.Name] = p
	}
	return m
}

func policyModified(a, b migrate.Policy) bool {
	return a.Using != b.Using ||
		a.WithCheck != b.WithCheck ||
		a.Command != b.Command ||
		a.Permissive != b.Permissive ||
		!stringSliceEqual(a.Roles, b.Roles)
}

// ---- trigger diffing ----

func diffTriggers(tableName string, current, desired []migrate.Trigger) []migrate.Change {
	var changes []migrate.Change

	curMap := triggerMap(current)
	desMap := triggerMap(desired)

	for name, dt := range desMap {
		if _, ok := curMap[name]; !ok {
			t := dt
			changes = append(changes, migrate.AddTrigger{Table: tableName, T: &t})
		}
	}

	for name, ct := range curMap {
		if _, ok := desMap[name]; !ok {
			t := ct
			changes = append(changes, migrate.DropTrigger{Table: tableName, T: &t})
		}
	}

	for name, ct := range curMap {
		dt, ok := desMap[name]
		if !ok {
			continue
		}
		if triggerModified(ct, dt) {
			from := ct
			to := dt
			changes = append(changes, migrate.ModifyTrigger{Table: tableName, From: &from, To: &to})
		}
	}

	return changes
}

func triggerMap(triggers []migrate.Trigger) map[string]migrate.Trigger {
	m := make(map[string]migrate.Trigger, len(triggers))
	for _, t := range triggers {
		m[t.Name] = t
	}
	return m
}

func triggerModified(a, b migrate.Trigger) bool {
	return !stringSliceEqual(a.Events, b.Events) ||
		a.Timing != b.Timing ||
		a.Function != b.Function ||
		a.When != b.When
}

// ---- extension diffing ----

func (d *Differ) diffExtensions(current, desired []migrate.Extension) []migrate.Change {
	var changes []migrate.Change

	curMap := extensionMap(current)
	desMap := extensionMap(desired)

	for name, de := range desMap {
		if _, ok := curMap[name]; !ok {
			e := de
			changes = append(changes, migrate.AddExtension{E: &e})
		}
	}

	for name, ce := range curMap {
		if _, ok := desMap[name]; !ok {
			e := ce
			changes = append(changes, migrate.DropExtension{E: &e})
		}
	}

	return changes
}

func extensionMap(exts []migrate.Extension) map[string]migrate.Extension {
	m := make(map[string]migrate.Extension, len(exts))
	for _, e := range exts {
		m[e.Name] = e
	}
	return m
}

// ---- function diffing ----

func (d *Differ) diffFunctions(current, desired []migrate.Function) []migrate.Change {
	var changes []migrate.Change

	curMap := functionMap(current)
	desMap := functionMap(desired)

	for sig, df := range desMap {
		if _, ok := curMap[sig]; !ok {
			f := df
			changes = append(changes, migrate.AddFunction{F: &f})
		}
	}

	for sig, cf := range curMap {
		if _, ok := desMap[sig]; !ok {
			f := cf
			changes = append(changes, migrate.DropFunction{F: &f})
		}
	}

	for sig, cf := range curMap {
		df, ok := desMap[sig]
		if !ok {
			continue
		}
		if functionModified(cf, df) {
			from := cf
			to := df
			changes = append(changes, migrate.ModifyFunction{From: &from, To: &to})
		}
	}

	return changes
}

// functionSignature returns a stable key for a function: name(arg1type,arg2type,...).
func functionSignature(f migrate.Function) string {
	argTypes := make([]string, len(f.Args))
	for i, a := range f.Args {
		argTypes[i] = a.Type
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(argTypes, ","))
}

func functionMap(funcs []migrate.Function) map[string]migrate.Function {
	m := make(map[string]migrate.Function, len(funcs))
	for _, f := range funcs {
		m[functionSignature(f)] = f
	}
	return m
}

func functionModified(a, b migrate.Function) bool {
	return a.Body != b.Body ||
		a.ReturnType != b.ReturnType ||
		a.Language != b.Language ||
		a.Volatility != b.Volatility ||
		a.Security != b.Security
}

// ---- utility ----

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
