package main

import (
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

// RelationInfo contains parsed relation information
type RelationInfo struct {
	// Relation variable name (e.g., "UserOrders", "OrderUser")
	VarName string
	// Relation type: "HasMany", "HasOne", "BelongsTo", "ManyToMany"
	RelationType string
	// Current message name (e.g., "User")
	CurrentMsg string
	// Related message name (e.g., "Order")
	RelatedMsg string
	// Field name in Scanner struct for the relation (e.g., "Orders", "User")
	FieldName string
	// Is it a slice (repeated field)?
	IsSlice bool
	// Foreign key column name
	ForeignKey string
	// Local key column name (primary key)
	LocalKey string
	// Owner key for BelongsTo
	OwnerKey string
	// For ManyToMany: pivot table info
	PivotTable string
}

// parseRelations extracts relation info from a table
func parseRelations(table *RatelTable) []*RelationInfo {
	relations := make([]*RelationInfo, 0, len(table.Relations))

	msgName := table.Message.GoIdent.GoName
	tableName := getTableName(table)

	for _, rel := range table.Relations {
		info := &RelationInfo{
			CurrentMsg: msgName,
			FieldName:  rel.Field.GoName,
			IsSlice:    rel.Field.Desc.IsList(),
		}

		// Get related message name
		if rel.Field.Message != nil {
			info.RelatedMsg = rel.Field.Message.GoIdent.GoName
		}

		// Determine relation type and extract options
		switch {
		case rel.Options.GetOneToMany() != nil:
			info.RelationType = "HasMany"
			otm := rel.Options.GetOneToMany()
			info.ForeignKey = otm.RefName
			if info.ForeignKey == "" {
				info.ForeignKey = tableName + "_id"
			}
			info.LocalKey = findPrimaryKey(table)
			info.VarName = msgName + info.RelatedMsg + "s"

		case rel.Options.GetHasOne() != nil:
			info.RelationType = "HasOne"
			ho := rel.Options.GetHasOne()
			info.ForeignKey = ho.RefName
			if info.ForeignKey == "" {
				info.ForeignKey = tableName + "_id"
			}
			info.LocalKey = findPrimaryKey(table)
			info.VarName = msgName + info.RelatedMsg

		case rel.Options.GetBelongsTo() != nil:
			info.RelationType = "BelongsTo"
			bt := rel.Options.GetBelongsTo()
			info.ForeignKey = bt.ForeignKey
			if info.ForeignKey == "" {
				info.ForeignKey = strcase.ToSnake(info.RelatedMsg) + "_id"
			}
			info.OwnerKey = bt.OwnerKey
			if info.OwnerKey == "" {
				info.OwnerKey = "id"
			}
			info.VarName = msgName + info.RelatedMsg

		case rel.Options.GetManyToMany() != nil:
			info.RelationType = "ManyToMany"
			// ManyToMany requires more complex handling - skip for now
			// Would need pivot table info
			continue
		}

		relations = append(relations, info)
	}

	return relations
}

// findPrimaryKey finds the primary key column name for a table
func findPrimaryKey(table *RatelTable) string {
	// Check composite primary key first
	if table.Options != nil && table.Options.PrimaryKey != nil && len(table.Options.PrimaryKey.Columns) > 0 {
		return table.Options.PrimaryKey.Columns[0]
	}

	// Look for column with primary_key constraint
	for _, col := range table.Columns {
		if col.Options != nil && col.Options.Constraints != nil && col.Options.Constraints.PrimaryKey {
			return col.SQLName
		}
	}

	// Default to "id"
	return "id"
}

// generateRelationVariables generates relation variable declarations
func generateRelationVariables(gf *protogen.GeneratedFile, table *RatelTable, relations []*RelationInfo) {
	if len(relations) == 0 {
		return
	}

	msgName := table.Message.GoIdent.GoName
	aliasTypeName := msgName + "Alias"
	colAliasTypeName := msgName + "ColumnAlias"
	scannerTypeName := msgName + "Scanner"

	gf.P("// ============================================================================")
	gf.P("// ", msgName, " Relations")
	gf.P("// ============================================================================")
	gf.P()

	for _, rel := range relations {
		relatedAliasTypeName := rel.RelatedMsg + "Alias"
		relatedColAliasTypeName := rel.RelatedMsg + "ColumnAlias"
		relatedScannerTypeName := rel.RelatedMsg + "Scanner"
		relatedTableVar := rel.RelatedMsg + "s"

		switch rel.RelationType {
		case "HasMany":
			// var UserOrders = schema.HasMany[...]
			fkColConst := rel.RelatedMsg + "Column" + strcase.ToCamel(rel.ForeignKey)
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			gf.P("// ", rel.VarName, " defines the one-to-many relationship: ", msgName, " has many ", rel.RelatedMsg)
			gf.P("var ", rel.VarName, " = schema.HasMany[")
			gf.P("\t", aliasTypeName, ", ", colAliasTypeName, ", *", scannerTypeName, ",")
			gf.P("\t", relatedAliasTypeName, ", ", relatedColAliasTypeName, ", *", relatedScannerTypeName, ",")
			gf.P("](")
			gf.P("\t", aliasTypeName, "Name,")
			gf.P("\t", relatedTableVar, "Ref,")
			gf.P("\t", fkColConst, ",")
			gf.P("\t", localColConst, ",")
			gf.P(")")
			gf.P()

		case "HasOne":
			// var UserProfile = schema.HasOne[...]
			fkColConst := rel.RelatedMsg + "Column" + strcase.ToCamel(rel.ForeignKey)
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			gf.P("// ", rel.VarName, " defines the one-to-one relationship: ", msgName, " has one ", rel.RelatedMsg)
			gf.P("var ", rel.VarName, " = schema.HasOne[")
			gf.P("\t", aliasTypeName, ", ", colAliasTypeName, ", *", scannerTypeName, ",")
			gf.P("\t", relatedAliasTypeName, ", ", relatedColAliasTypeName, ", *", relatedScannerTypeName, ",")
			gf.P("](")
			gf.P("\t", aliasTypeName, "Name,")
			gf.P("\t", relatedTableVar, "Ref,")
			gf.P("\t", fkColConst, ",")
			gf.P("\t", localColConst, ",")
			gf.P(")")
			gf.P()

		case "BelongsTo":
			// var OrderUser = schema.BelongsTo[...]
			fkColConst := msgName + "Column" + strcase.ToCamel(rel.ForeignKey)
			ownerColConst := rel.RelatedMsg + "Column" + strcase.ToCamel(rel.OwnerKey)

			gf.P("// ", rel.VarName, " defines the belongs-to relationship: ", msgName, " belongs to ", rel.RelatedMsg)
			gf.P("var ", rel.VarName, " = schema.BelongsTo[")
			gf.P("\t", aliasTypeName, ", ", colAliasTypeName, ", *", scannerTypeName, ",")
			gf.P("\t", relatedAliasTypeName, ", ", relatedColAliasTypeName, ", *", relatedScannerTypeName, ",")
			gf.P("](")
			gf.P("\t", aliasTypeName, "Name,")
			gf.P("\t", relatedTableVar, "Ref,")
			gf.P("\t", fkColConst, ",")
			gf.P("\t", ownerColConst, ",")
			gf.P(")")
			gf.P()
		}
	}
}

// generateRelationsMethod generates the Relations() method with loaders
func generateRelationsMethod(gf *protogen.GeneratedFile, table *RatelTable, relations []*RelationInfo) {
	msgName := table.Message.GoIdent.GoName
	colAliasTypeName := msgName + "ColumnAlias"
	scannerTypeName := msgName + "Scanner"

	gf.P("// Relations returns the relation loaders for the ", getTableName(table), " table")
	gf.P("func (s *", scannerTypeName, ") Relations() []exec.RelationLoader[*", scannerTypeName, "] {")

	if len(relations) == 0 {
		gf.P("\treturn nil")
		gf.P("}")
		gf.P()
		return
	}

	gf.P("\treturn []exec.RelationLoader[*", scannerTypeName, "]{")

	for _, rel := range relations {
		relatedScannerTypeName := rel.RelatedMsg + "Scanner"
		relatedTableVar := rel.RelatedMsg + "s"

		switch rel.RelationType {
		case "HasMany":
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			// HasMany uses slice of values (not pointers) in Scanner struct
			// but schema.HasManyLoad returns slice of pointers, so we need to convert
			gf.P("\t\tschema.HasManyLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", localColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related []*", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = make([]", relatedScannerTypeName, ", len(related))")
			gf.P("\t\t\t\tfor i, r := range related {")
			gf.P("\t\t\t\t\tif r != nil {")
			gf.P("\t\t\t\t\t\tbase.", rel.FieldName, "[i] = *r")
			gf.P("\t\t\t\t\t}")
			gf.P("\t\t\t\t}")
			gf.P("\t\t\t},")
			gf.P("\t\t),")

		case "HasOne":
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			gf.P("\t\tschema.HasOneLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", localColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related *", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = related")
			gf.P("\t\t\t},")
			gf.P("\t\t),")

		case "BelongsTo":
			fkColConst := msgName + "Column" + strcase.ToCamel(rel.ForeignKey)

			gf.P("\t\tschema.BelongsToLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", fkColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related *", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = related")
			gf.P("\t\t\t},")
			gf.P("\t\t),")
		}
	}

	gf.P("\t}")
	gf.P("}")
	gf.P()

	// Generate type check
	_ = colAliasTypeName // Used in constants above
}

// generateRelationOptions generates typed QueryOption functions for each relation
func generateRelationOptions(gf *protogen.GeneratedFile, table *RatelTable, relations []*RelationInfo) {
	if len(relations) == 0 {
		return
	}

	msgName := table.Message.GoIdent.GoName
	colAliasTypeName := msgName + "ColumnAlias"
	scannerTypeName := msgName + "Scanner"
	tableName := msgName + "s" // e.g., "Users"

	gf.P("// ============================================================================")
	gf.P("// ", msgName, " Relation Query Options")
	gf.P("// ============================================================================")
	gf.P()

	// Generate individual relation options
	for _, rel := range relations {
		relatedScannerTypeName := rel.RelatedMsg + "Scanner"
		relatedTableVar := rel.RelatedMsg + "s"
		optionFuncName := tableName + "With" + rel.FieldName // e.g., "UsersWithOrders"

		switch rel.RelationType {
		case "HasMany":
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			gf.P("// ", optionFuncName, " returns a QueryOption to load ", rel.FieldName, " relation")
			gf.P("func ", optionFuncName, "() exec.QueryOption[", colAliasTypeName, ", *", scannerTypeName, "] {")
			gf.P("\treturn exec.WithRelationLoaders[", colAliasTypeName, ", *", scannerTypeName, "](")
			gf.P("\t\tschema.HasManyLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", localColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related []*", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = make([]", relatedScannerTypeName, ", len(related))")
			gf.P("\t\t\t\tfor i, r := range related {")
			gf.P("\t\t\t\t\tif r != nil {")
			gf.P("\t\t\t\t\t\tbase.", rel.FieldName, "[i] = *r")
			gf.P("\t\t\t\t\t}")
			gf.P("\t\t\t\t}")
			gf.P("\t\t\t},")
			gf.P("\t\t),")
			gf.P("\t)")
			gf.P("}")
			gf.P()

		case "HasOne":
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			gf.P("// ", optionFuncName, " returns a QueryOption to load ", rel.FieldName, " relation")
			gf.P("func ", optionFuncName, "() exec.QueryOption[", colAliasTypeName, ", *", scannerTypeName, "] {")
			gf.P("\treturn exec.WithRelationLoaders[", colAliasTypeName, ", *", scannerTypeName, "](")
			gf.P("\t\tschema.HasOneLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", localColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related *", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = related")
			gf.P("\t\t\t},")
			gf.P("\t\t),")
			gf.P("\t)")
			gf.P("}")
			gf.P()

		case "BelongsTo":
			fkColConst := msgName + "Column" + strcase.ToCamel(rel.ForeignKey)

			gf.P("// ", optionFuncName, " returns a QueryOption to load ", rel.FieldName, " relation")
			gf.P("func ", optionFuncName, "() exec.QueryOption[", colAliasTypeName, ", *", scannerTypeName, "] {")
			gf.P("\treturn exec.WithRelationLoaders[", colAliasTypeName, ", *", scannerTypeName, "](")
			gf.P("\t\tschema.BelongsToLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", fkColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related *", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = related")
			gf.P("\t\t\t},")
			gf.P("\t\t),")
			gf.P("\t)")
			gf.P("}")
			gf.P()
		}
	}

	// Generate WithAllRelations option
	allRelationsFuncName := tableName + "WithAllRelations"
	gf.P("// ", allRelationsFuncName, " returns a QueryOption to load all relations")
	gf.P("func ", allRelationsFuncName, "() exec.QueryOption[", colAliasTypeName, ", *", scannerTypeName, "] {")
	gf.P("\treturn exec.WithRelationLoaders[", colAliasTypeName, ", *", scannerTypeName, "](")

	for _, rel := range relations {
		relatedScannerTypeName := rel.RelatedMsg + "Scanner"
		relatedTableVar := rel.RelatedMsg + "s"

		switch rel.RelationType {
		case "HasMany":
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			gf.P("\t\tschema.HasManyLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", localColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related []*", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = make([]", relatedScannerTypeName, ", len(related))")
			gf.P("\t\t\t\tfor i, r := range related {")
			gf.P("\t\t\t\t\tif r != nil {")
			gf.P("\t\t\t\t\t\tbase.", rel.FieldName, "[i] = *r")
			gf.P("\t\t\t\t\t}")
			gf.P("\t\t\t\t}")
			gf.P("\t\t\t},")
			gf.P("\t\t),")

		case "HasOne":
			localColConst := msgName + "Column" + strcase.ToCamel(rel.LocalKey)

			gf.P("\t\tschema.HasOneLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", localColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related *", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = related")
			gf.P("\t\t\t},")
			gf.P("\t\t),")

		case "BelongsTo":
			fkColConst := msgName + "Column" + strcase.ToCamel(rel.ForeignKey)

			gf.P("\t\tschema.BelongsToLoad(")
			gf.P("\t\t\t", rel.VarName, ",")
			gf.P("\t\t\t", relatedTableVar, ".Table,")
			gf.P("\t\t\t", fkColConst, ",")
			gf.P("\t\t\tfunc(base *", scannerTypeName, ", related *", relatedScannerTypeName, ") {")
			gf.P("\t\t\t\tbase.", rel.FieldName, " = related")
			gf.P("\t\t\t},")
			gf.P("\t\t),")
		}
	}

	gf.P("\t)")
	gf.P("}")
	gf.P()
}
