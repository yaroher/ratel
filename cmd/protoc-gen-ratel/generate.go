package main

import (
	"strings"

	"github.com/yaroher/protoc-gen-go-plain/generator"
	"github.com/yaroher/protoc-gen-go-plain/goplain"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	typepb "google.golang.org/protobuf/types/known/typepb"
)

// Generate is the main entry point for the protoc-gen-ratel plugin
func Generate(p *protogen.Plugin) error {
	settings, err := generator.NewPluginSettingsFromPlugin(p)
	if err != nil {
		return err
	}

	// Override settings for ratel
	settings.JSONJX = false
	settings.JXPB = false
	settings.GeneratePool = false

	// Inject virtual_columns from ratel.table as virtual_fields into goplain options
	// so go-plain adds them to the Scanner struct automatically
	injectVirtualFields(p)

	// Create generator with ratel-specific options
	// WithForceEnumAsString: ratel stores enums as TEXT in PostgreSQL,
	// so Scanner fields must be string for pgx to scan text values directly.
	g, err := generator.NewGenerator(p, settings,
		generator.WithPlainSuffix("Scanner"),
		generator.WithForceEnumAsString(),
		generator.WithTypeOverrides(getRatelTypeOverrides()),
		generator.WithExistingCasters(getRatelCasters()),
	)
	if err != nil {
		return err
	}

	// Run the standard plain generation (generates *Scanner structs)
	if err := g.Generate(); err != nil {
		return err
	}

	// Generate ratel-specific files
	return generateRatelFiles(p, g)
}

// generateRatelFiles generates ratel-specific code
func generateRatelFiles(p *protogen.Plugin, gen *generator.Generator) error {
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}

		// Collect ratel tables directly from protogen messages
		tables := collectRatelTables(f)
		if len(tables) == 0 {
			continue
		}

		// Generate ratel file
		if err := generateRatelFile(p, f, tables, gen); err != nil {
			return err
		}
	}

	return nil
}

// injectVirtualFields reads ratel.table.virtual_columns and injects them as
// goplain.message.virtual_fields so go-plain adds them to the Scanner struct.
func injectVirtualFields(p *protogen.Plugin) {
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}
		for _, msg := range f.Messages {
			tableOpts := getRatelTableOptions(msg)
			if tableOpts == nil || len(tableOpts.VirtualColumns) == 0 {
				continue
			}

			// Convert virtual_columns to typepb.Field entries
			var vfields []*typepb.Field
			for _, vc := range tableOpts.VirtualColumns {
				vfields = append(vfields, &typepb.Field{
					Name: vc.SqlName,
					Kind: sqlTypeToFieldKind(vc.SqlType),
				})
			}

			// Get existing goplain message options and append virtual fields
			msgOpts := msg.Desc.Options()
			if msgOpts == nil {
				continue
			}
			ext := proto.GetExtension(msgOpts, goplain.E_Message)
			if ext == nil {
				continue
			}
			goplainOpts, ok := ext.(*goplain.MessageOptions)
			if !ok || goplainOpts == nil {
				continue
			}
			goplainOpts.VirtualFields = append(goplainOpts.VirtualFields, vfields...)
		}
	}
}

// sqlTypeToFieldKind maps SQL type string to protobuf Field.Kind for virtual fields.
func sqlTypeToFieldKind(sqlType string) typepb.Field_Kind {
	upper := strings.ToUpper(sqlType)
	switch {
	case upper == "BIGINT" || upper == "INT8" || upper == "BIGSERIAL" || upper == "SERIAL8":
		return typepb.Field_TYPE_INT64
	case upper == "INTEGER" || upper == "INT" || upper == "INT4" || upper == "SERIAL" || upper == "SERIAL4":
		return typepb.Field_TYPE_INT32
	case upper == "SMALLINT" || upper == "INT2":
		return typepb.Field_TYPE_INT32
	case upper == "BOOLEAN" || upper == "BOOL":
		return typepb.Field_TYPE_BOOL
	case upper == "REAL" || upper == "FLOAT4":
		return typepb.Field_TYPE_FLOAT
	case upper == "DOUBLE PRECISION" || upper == "FLOAT8":
		return typepb.Field_TYPE_DOUBLE
	case upper == "BYTEA":
		return typepb.Field_TYPE_BYTES
	default:
		// TEXT, VARCHAR, TIMESTAMPTZ, JSONB, UUID, etc. → string in Scanner
		return typepb.Field_TYPE_STRING
	}
}
