package stack

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"slices"
	"time"
)

func (ss *service) MigrateDatabase(
	ctx context.Context,
	stackId uint,
	input MigrateDatabaseInput,
) error {
	stack, err := ss.GetStack(ctx, stackId)
	if err != nil {
		return err
	}

	regionalDbConfig := ss.dbConfig.GetRegionalConfig(stack.DefaultRegion)
	conn, err := pgx.Connect(
		ctx,
		stack.DB.Data().PostgresURI(regionalDbConfig.Host, regionalDbConfig.Port),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to database")
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		return errors.Wrapf(err, "failed to ping database")
	}

	slices.SortStableFunc(input.Migrations, func(i, j Migration) int {
		return int(i.Version.Sub(j.Version))
	})

	for _, migration := range input.Migrations {
		skip := false
		if err := pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
			migrationTableNameCandidates := []string{
				"stack.schema_migrations",
				"stack.migrations", // for backward compatibility
			}

			for i, tableName := range migrationTableNameCandidates {
				isLast := len(migrationTableNameCandidates) == i+1

				if _, err := tx.Exec(
					ctx,
					fmt.Sprintf(
						"INSERT INTO %s (version) VALUES ($1)",
						tableName,
					),
					migration.Version,
				); err != nil {
					var pgErr *pgconn.PgError
					if errors.As(err, &pgErr) {
						if pgErr.Code == "23505" { // checking duplicated key violates primary key
							logger.Info("migration already exists", "version", migration.Version)
							skip = true
						} else if pgErr.Code == "42P01" && !isLast {
							continue
						}
					}

					return errors.Wrapf(err, "failed to insert migration. version=%v", migration.Version)
				}

				break
			}

			if _, err := tx.Exec(ctx, migration.Query); err != nil {
				return errors.Wrapf(err, "failed to execute migration")
			}
			return nil
		}); err != nil && !skip {
			return err
		}
	}

	if stack.PostgrestEnabled {
		if _, err := conn.Exec(ctx, "NOTIFY pgrst, 'reload schema';"); err != nil {
			return errors.Wrapf(err, "failed to notify pgrst")
		}

		if err := ss.waitPostgrestForReady(ctx, stack, 5*time.Second); err != nil {
			return errors.Wrapf(err, "failed to wait postgrest for ready")
		}
	}

	return nil
}
