package main

import (
	"github.com/yaroher/protoc-gen-go-plain/generator"
	"google.golang.org/protobuf/compiler/protogen"
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

	// Create generator with ratel-specific options
	g, err := generator.NewGenerator(p, settings,
		generator.WithPlainSuffix("Scanner"),
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
