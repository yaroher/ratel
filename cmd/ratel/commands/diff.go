package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlclient"
	"github.com/spf13/cobra"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	diffSqlFile      string
	diffMigrationDir string
	diffName         string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Generate SQL migration diff",
	Long: `Generate a SQL migration diff between the current database schema and the schema defined in Go models.

Example:
  ratel diff -p github.com/myproject/models -o migrations/001_initial.sql

This will create a SQL migration file that can be applied to the database to bring it up to date with the schema defined in your Go models.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiff(cmd, args)
	},
}

const defaultPostgresVersion = 18

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&diffSqlFile, "sql", "s", "", "SQL schema file to compare against (required)")
	diffCmd.Flags().StringVarP(&diffMigrationDir, "dir", "d", "./migrations", "Migration directory for output")
	diffCmd.Flags().StringVarP(&diffName, "name", "n", "migration", "Migration name")
	diffCmd.Flags().Int16P("pg_version", "v", defaultPostgresVersion, "PostgreSQL version")
	diffCmd.MarkFlagRequired("sql")
	diffCmd.MarkFlagRequired("dir")
}

func runDiff(cmd *cobra.Command, _ []string) error {
	// Validate flags
	if diffSqlFile == "" {
		return errors.New("sql file is required (use --sql or -s flag)")
	}
	if diffMigrationDir == "" {
		return errors.New("migration directory is required (use --dir or -d flag)")
	}
	if diffName == "" {
		return errors.New("migration name is required (use --name or -n flag)")
	}
	diffPgVersion, _ := cmd.Flags().GetInt16("pg_version")

	// Check if SQL file exists
	if _, err := os.Stat(diffSqlFile); os.IsNotExist(err) {
		return fmt.Errorf("sql file does not exist: %s", diffSqlFile)
	}

	// Create migration directory if it doesn't exist
	if err := os.MkdirAll(diffMigrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	pgContainer, err := postgres.Run(cmd.Context(),
		"postgres:"+fmt.Sprintf("%d-alpine", diffPgVersion),
		postgres.WithDatabase("ratel"),
		postgres.WithUsername("ratel"),
		postgres.WithPassword("ratel"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}
	defer func() {
		if err := pgContainer.Terminate(cmd.Context()); err != nil {
			panic(err)
		}
	}()
	connStr, err := pgContainer.ConnectionString(cmd.Context(), "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get connection string: %v", err)
	}
	return migrateDiff(cmd.Context(), connStr, diffSqlFile, diffMigrationDir, diffName)
}

func migrateDiff(ctx context.Context, devURL, sqlFilePath, migrationDir, migrationName string) error {
	if devURL == "" {
		return errors.New("dev-url is required")
	}
	if sqlFilePath == "" {
		return errors.New("sql file path is required")
	}
	if migrationDir == "" {
		return errors.New("migration directory is required")
	}

	// Открываем соединение с dev-базой
	dev, err := sqlclient.Open(ctx, devURL)
	if err != nil {
		return fmt.Errorf("failed to open dev database: %w", err)
	}
	defer dev.Close()

	// Получаем блокировку на 10 секунд
	unlock, err := dev.Lock(ctx, "atlas_migrate_diff", 10*time.Second)
	if err != nil {
		return fmt.Errorf("acquiring database lock: %w", err)
	}
	defer func() {
		if err := unlock(); err != nil {
			fmt.Printf("Warning: failed to unlock database: %v\n", err)
		}
	}()

	// Читаем SQL файл
	sqlContent, err := os.ReadFile(sqlFilePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// Создаем временную директорию для SQL файла
	sqlFile := migrate.NewLocalFile(filepath.Base(sqlFilePath), sqlContent)
	sqlDir := migrate.MemDir{}
	if err := sqlDir.WriteFile(sqlFile.Name(), sqlFile.Bytes()); err != nil {
		return fmt.Errorf("failed to create in-memory SQL directory: %w", err)
	}

	// Открываем директорию миграций
	dir, err := migrate.NewLocalDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to open migration directory: %w", err)
	}

	executor, err := migrate.NewExecutor(dev.Driver, &sqlDir, migrate.NopRevisionReadWriter{})
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	var stateReader migrate.StateReader
	if dev.URL.Schema != "" {
		stateReader = migrate.SchemaConn(dev, "", nil)
	} else {
		stateReader = migrate.RealmConn(dev, &schema.InspectRealmOption{})
	}

	desiredState, err := executor.Replay(ctx, stateReader)
	if err != nil && !errors.Is(err, migrate.ErrNoPendingFiles) {
		return fmt.Errorf("failed to replay SQL file: %w", err)
	}

	diffOpts := []schema.DiffOption{schema.DiffNormalized()}

	plannerOpts := []migrate.PlannerOption{
		migrate.PlanWithDiffOptions(diffOpts...),
	}

	planner := migrate.NewPlanner(dev.Driver, dir, plannerOpts...)
	plan, err := func() (*migrate.Plan, error) {
		if dev.URL.Schema != "" {
			return planner.PlanSchema(ctx, migrationName, migrate.Realm(desiredState))
		}
		return planner.Plan(ctx, migrationName, migrate.Realm(desiredState))
	}()

	var cerr *migrate.NotCleanError
	switch {
	case errors.As(err, &cerr):
		return fmt.Errorf("dev database is not clean (%s)", cerr.Reason)
	case errors.Is(err, migrate.ErrNoPlan):
		// Нет изменений - это не ошибка
		return nil
	case err != nil:
		return fmt.Errorf("failed to plan migration: %w", err)
	default:
		if err := planner.WritePlan(plan); err != nil {
			return fmt.Errorf("failed to write migration plan: %w", err)
		}
		return nil
	}
}
