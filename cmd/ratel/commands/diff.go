package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	diffPackages     []string
	diffTables       []string
	diffDiscover     bool
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Generate SQL migration diff",
	Long: `Generate a SQL migration diff between the current database schema and the schema defined in Go models.

Examples:
  # From SQL file:
  ratel diff -s schema.sql -d migrations -n add_users

  # From Go models package:
  ratel diff -p github.com/myproject/models -d migrations -n add_users

  # From Go models with auto-discovery:
  ratel diff -p github.com/myproject/models --discover -d migrations -n add_users

  # From multiple packages:
  ratel diff -p github.com/myproject/auth,github.com/myproject/store --discover -d migrations -n init

This will create a SQL migration file that can be applied to the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiff(cmd, args)
	},
}

const defaultPostgresVersion = 18

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&diffSqlFile, "sql", "s", "", "SQL schema file to compare against")
	diffCmd.Flags().StringSliceVarP(&diffPackages, "package", "p", nil, "Go package path(s) containing models (can be repeated)")
	diffCmd.Flags().StringSliceVarP(&diffTables, "tables", "t", nil, "Table variable names (e.g., Users,Products)")
	diffCmd.Flags().BoolVar(&diffDiscover, "discover", false, "Auto-discover tables from source files")
	diffCmd.Flags().StringVarP(&diffMigrationDir, "dir", "d", "./migrations", "Migration directory for output")
	diffCmd.Flags().StringVarP(&diffName, "name", "n", "migration", "Migration name")
	diffCmd.Flags().Int16P("pg_version", "v", defaultPostgresVersion, "PostgreSQL version")
	diffCmd.MarkFlagRequired("dir")
}

func runDiff(cmd *cobra.Command, _ []string) error {
	// Validate flags - need either sql file or package
	if diffSqlFile == "" && len(diffPackages) == 0 {
		return errors.New("either --sql (-s) or --package (-p) is required")
	}
	if diffSqlFile != "" && len(diffPackages) > 0 {
		return errors.New("cannot use both --sql and --package, choose one")
	}
	if diffMigrationDir == "" {
		return errors.New("migration directory is required (use --dir or -d flag)")
	}
	if diffName == "" {
		return errors.New("migration name is required (use --name or -n flag)")
	}
	diffPgVersion, _ := cmd.Flags().GetInt16("pg_version")

	var sqlFilePath string
	var cleanupSQL func()

	if len(diffPackages) > 0 {
		// Generate SQL from Go models package(s)
		tmpFile, err := generateSchemaFromPackages(diffPackages, diffTables)
		if err != nil {
			return fmt.Errorf("failed to generate schema from package: %w", err)
		}
		sqlFilePath = tmpFile
		cleanupSQL = func() { os.Remove(tmpFile) }
	} else {
		// Use provided SQL file
		if _, err := os.Stat(diffSqlFile); os.IsNotExist(err) {
			return fmt.Errorf("sql file does not exist: %s", diffSqlFile)
		}
		sqlFilePath = diffSqlFile
		cleanupSQL = func() {} // no cleanup needed
	}
	defer cleanupSQL()

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
	return migrateDiff(cmd.Context(), connStr, sqlFilePath, diffMigrationDir, diffName)
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

// packageTables holds discovered tables for a specific package
type packageTables struct {
	pkg    string
	tables []string
}

// generateSchemaFromPackages generates SQL schema from one or more Go model packages
// and returns path to temporary SQL file
func generateSchemaFromPackages(packages []string, tables []string) (string, error) {
	workspaceRoot := mustGetWorkspaceRoot()

	var allPkgTables []packageTables

	if len(tables) > 0 && len(packages) == 1 {
		// Explicit tables with single package — backward compatible
		allPkgTables = append(allPkgTables, packageTables{pkg: packages[0], tables: tables})
	} else {
		// Discover tables from each package
		for _, pkg := range packages {
			discovered, err := discoverTables(pkg, workspaceRoot)
			if err != nil {
				return "", fmt.Errorf("failed to discover tables in %s: %w", pkg, err)
			}
			if len(discovered) == 0 {
				return "", fmt.Errorf("no tables discovered in package %s", pkg)
			}
			fmt.Printf("Discovered tables in %s: %v\n", pkg, discovered)
			allPkgTables = append(allPkgTables, packageTables{pkg: pkg, tables: discovered})
		}
	}

	if len(allPkgTables) == 0 {
		return "", errors.New("no tables specified")
	}

	// Create temporary Go file in workspace
	tmpGoFile, err := os.CreateTemp(workspaceRoot, "ratel_schema_gen_*.go")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpGoFileName := tmpGoFile.Name()
	defer os.Remove(tmpGoFileName)

	// Generate the temporary Go program
	program := generateMultiPkgSchemaProgram(allPkgTables)

	if _, err := tmpGoFile.WriteString(program); err != nil {
		tmpGoFile.Close()
		return "", fmt.Errorf("failed to write temp program: %w", err)
	}
	tmpGoFile.Close()

	// Run the program from workspace root
	runCmd := execCommand("go", "run", tmpGoFileName)
	runCmd.Dir = workspaceRoot
	runCmd.Stderr = os.Stderr
	output, err := runCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run schema generator: %w", err)
	}

	// Write output to temporary SQL file
	tmpSqlFile, err := os.CreateTemp("", "ratel_schema_*.sql")
	if err != nil {
		return "", fmt.Errorf("failed to create temp SQL file: %w", err)
	}

	if _, err := tmpSqlFile.Write(output); err != nil {
		tmpSqlFile.Close()
		os.Remove(tmpSqlFile.Name())
		return "", fmt.Errorf("failed to write temp SQL file: %w", err)
	}
	tmpSqlFile.Close()

	return tmpSqlFile.Name(), nil
}

// execCommand is a wrapper for exec.Command (allows testing)
var execCommand = func(name string, arg ...string) *execCmd {
	return &execCmd{exec.Command(name, arg...)}
}

type execCmd struct {
	*exec.Cmd
}
