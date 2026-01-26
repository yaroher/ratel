package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"ariga.io/atlas/sql/migrate"
	pbMigrate "ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
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
  ratel diff -s schema.sql -d migrations -n 001_initial

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

	pool, err := pgxpool.New(ctx, devURL)
	if err != nil {
		return fmt.Errorf("failed to create pgx pool: %w", err)
	}
	defer pool.Close()

	dev, err := pbMigrate.Open(stdlib.OpenDBFromPool(pool))
	if err != nil {
		return fmt.Errorf("failed to open dev database: %w", err)
	}

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

	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationDir, 0755); err != nil {
			return fmt.Errorf("failed to create migration directory: %w", err)
		}
	}

	// Открываем директорию миграций
	dir, err := migrate.NewLocalDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to open migration directory: %w", err)
	}

	if _, err := dir.Checksum(); errors.Is(err, migrate.ErrChecksumNotFound) {
		// Создаем пустой файл контрольных сумм для новой директории
		if err := migrate.WriteSumFile(dir, migrate.HashFile{}); err != nil {
			return fmt.Errorf("failed to create checksum file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to read checksum file: %w", err)
	}

	// Применяем существующие миграции к dev базе
	executor, err := migrate.NewExecutor(dev, dir, migrate.NopRevisionReadWriter{})
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Получаем текущее состояние после применения существующих миграций
	currentState := migrate.RealmConn(dev, &schema.InspectRealmOption{})
	currentRealm, err := executor.Replay(ctx, currentState)
	if err != nil && !errors.Is(err, migrate.ErrNoPendingFiles) {
		return fmt.Errorf("failed to replay existing migrations: %w", err)
	}

	// Применяем SQL схему к базе данных
	if _, err := dev.ExecContext(ctx, string(sqlContent)); err != nil {
		return fmt.Errorf("failed to apply SQL schema: %w", err)
	}

	// Инспектируем состояние базы данных после применения схемы
	desiredState, err := dev.InspectRealm(ctx, &schema.InspectRealmOption{})
	if err != nil {
		return fmt.Errorf("failed to inspect database state: %w", err)
	}

	// Создаем новый файл миграции с изменениями
	diffOpts := []schema.DiffOption{schema.DiffNormalized()}
	changes, err := dev.RealmDiff(currentRealm, desiredState, diffOpts...)
	if err != nil {
		return fmt.Errorf("failed to calculate diff: %w", err)
	}

	if len(changes) == 0 {
		fmt.Println("No schema changes detected")
		return nil
	}

	// Используем PlanChanges для генерации плана миграции
	plan, err := dev.PlanChanges(ctx, migrationName, changes)
	if err != nil {
		return fmt.Errorf("failed to plan changes: %w", err)
	}

	// Используем Planner для записи плана
	plannerOpts := []migrate.PlannerOption{
		migrate.PlanWithDiffOptions(diffOpts...),
	}

	planner := migrate.NewPlanner(dev, dir, plannerOpts...)

	// Записываем план
	if err := planner.WritePlan(plan); err != nil {
		return fmt.Errorf("failed to write migration plan: %w", err)
	}

	fmt.Printf("Migration '%s' created successfully\n", migrationName)
	return nil
}
