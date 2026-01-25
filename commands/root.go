package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ratel",
	Short: "Ratel - PostgreSQL ORM code generator for Go",
	Long: `
██████╗  █████╗ ████████╗███████╗██╗
██╔══██╗██╔══██╗╚══██╔══╝██╔════╝██║
██████╔╝███████║   ██║   █████╗  ██║
██╔══██╗██╔══██║   ██║   ██╔══╝  ██║
██║  ██║██║  ██║   ██║   ███████╗███████╗
╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚══════╝╚══════╝

Ratel is a type-safe PostgreSQL ORM for Go.

Commands:
  generate  - Generate Go models from SQL schema file
  schema    - Generate SQL schema from Go models directory
`,
	Version: "0.1.0",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// exitWithError prints error and exits
func exitWithError(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+msg+"\n", args...)
	os.Exit(1)
}
