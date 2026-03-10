package commands

import (
	"errors"
	"fmt"

	"ariga.io/atlas/sql/migrate"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migration management commands",
}

var (
	hashDir    string
	hashEngine string
)

var migrateHashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Recalculate atlas.sum checksum for migrations directory",
	Long: `Recalculate the atlas.sum checksum file for the given migrations directory.
Use this after manually adding or editing .sql migration files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrateHash(hashDir, hashEngine)
	},
}

func init() {
	migrateHashCmd.Flags().StringVarP(&hashDir, "dir", "d", "./migrations", "Migration directory")
	migrateHashCmd.Flags().StringVarP(&hashEngine, "engine", "e", "atlas", "Migration engine: atlas or ratel")

	migrateCmd.AddCommand(migrateHashCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runMigrateHash(dir, engine string) error {
	switch engine {
	case "atlas", "ratel":
		// Both engines use atlas.sum format.
	default:
		return fmt.Errorf("unsupported engine: %s (use atlas or ratel)", engine)
	}

	localDir, err := migrate.NewLocalDir(dir)
	if err != nil {
		return fmt.Errorf("opening migrations directory %q: %w", dir, err)
	}

	sum, err := localDir.Checksum()
	if err != nil && !errors.Is(err, migrate.ErrChecksumNotFound) {
		return fmt.Errorf("computing checksum: %w", err)
	}
	if sum == nil {
		sum = migrate.HashFile{}
	}

	if err := migrate.WriteSumFile(localDir, sum); err != nil {
		return fmt.Errorf("writing atlas.sum: %w", err)
	}

	fmt.Printf("atlas.sum updated in %s (engine: %s)\n", dir, engine)
	return nil
}
