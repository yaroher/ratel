package migrate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/yaroher/ratel/internal/atlas"
	"github.com/yaroher/ratel/internal/mfs"
	"github.com/yaroher/ratel/pkg/pgx-ext/sqlexec"
)

// Migrate Run migration. usual uses in service startup.
func Migrate(pool *pgxpool.Pool, lg *zap.Logger, migrations ...fs.FS) error {
	revision, err := atlas.NewRevisionReaderWriter(sqlexec.NewTxExecutor(pool))
	if err != nil {
		return fmt.Errorf("failed to create revision reader writer in AtlasMigrate: %v", err)
	}
	executor := stdlib.OpenDBFromPool(pool)
	pg, err := postgres.Open(executor)
	if err != nil {
		return fmt.Errorf("failed to open postgres in AtlasMigrate: %v", err)
	}
	migrator, err := migrate.NewExecutor(pg,
		atlas.NewEmbedDir(mfs.MergeMultiple(migrations...)),
		revision,
		migrate.WithAllowDirty(true),
		migrate.WithLogger(atlas.NewMigrationLogger(lg)),
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator in AtlasMigrate: %v", err)
	}
	err = migrator.ExecuteN(context.Background(), 0)
	if err != nil {
		if errors.Is(err, migrate.ErrNoPendingFiles) {
			lg.Info("no pending migrations in AtlasMigrate to apply")
			return nil
		}
		lg.Error("failed to apply migrations in AtlasMigrate: %v", zap.Error(err))
		return err
	}
	return nil
}
