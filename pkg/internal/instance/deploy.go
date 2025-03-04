package instance

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/util/functx/v2"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"slices"
	"time"
)

func (s *service) DeployStack(
	ctx context.Context,
	instanceId uint,
	input DeployStackInput,
) error {
	if input.Timeout == nil || *input.Timeout == "" {
		input.Timeout = gog.PtrOf("30s")
	}

	timeout, err := time.ParseDuration(*input.Timeout)
	if err != nil {
		return errors.Wrapf(tclerrors.ErrBadRequest, "failed to parse timeout: %v", err)
	}

	tx := helpers.GetTx(ctx)
	ctx, fDone := functx.WithFuncTx(ctx)
	defer fDone(ctx, true)

	instance, err := s.GetInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	if instance.State == domain.InstanceStateNone {
		if err := tx.Transaction(func(tx *gorm.DB) error {
			if err := instance.TransitionToInitialize(tx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if err := tx.Transaction(func(tx *gorm.DB) (rErr error) {
		if err := s.applyK8s(ctx, instance, timeout); err != nil {
			return err
		}

		if err := s.migrationDatabase(ctx, instance); err != nil {
			return err
		}

		if err := instance.TransitionToRunning(tx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	fDone(ctx, false)
	return nil
}

func (s *service) migrationDatabase(
	ctx context.Context,
	instance *domain.Instance,
) error {
	tx := helpers.GetTx(ctx)
	dbData := instance.Stack.DB.Data()

	if instance.Stack.Vapis == nil {
		if err := tx.Model(&instance.Stack).Preload("Vapi").Association("Vapis").Find(&instance.Stack.Vapis); err != nil {
			return errors.Wrapf(err, "failed to find stack vapis")
		}
	}

	regionalDbConfig := s.dbConfig.GetRegionalConfig(instance.Stack.DefaultRegion)
	conn, err := pgx.Connect(ctx, dbData.PostgresURI(regionalDbConfig.Host, regionalDbConfig.Port))
	if err != nil {
		return errors.Wrapf(err, "failed to connect to postgres")
	}
	defer conn.Close(ctx)

	if err := pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		for _, stackVapi := range instance.Stack.Vapis {
			vapiRelease := stackVapi.Vapi
			if err := s.migrateVapiDatabase(ctx, tx, &vapiRelease); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if instance.Stack.PostgrestEnabled {
		// postgrest had to be notified schema reloading
		if _, err := conn.Exec(ctx, "NOTIFY pgrst, 'reload schema';"); err != nil {
			return errors.Wrapf(err, "failed to notify pgrst")
		}
	}

	return nil
}

func (s *service) migrateVapiDatabase(
	ctx context.Context,
	conn pgx.Tx,
	vapiRelease *domain.VapiRelease,
) error {
	tx := helpers.GetTx(ctx)

	if err := vapiRelease.DFS(tx, func(v domain.VapiRelease, _ *domain.VapiRelease) error {
		migrations, err := s.vapis.GetDBMigrations(ctx, v)
		if err != nil {
			return err
		}

		slices.SortStableFunc(migrations, func(lhs, rhs vapi.Migration) int {
			return lhs.Version.Compare(rhs.Version)
		})

		return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
			rows, err := tx.Query(
				ctx,
				`SELECT version FROM stack.vapi_schema_migrations WHERE vapi_package_id = $1 ORDER BY version FOR UPDATE;`,
				v.PackageID,
			)
			if err != nil {
				return errors.Wrapf(err, "failed to select migrations")
			}

			versions, err := pgx.AppendRows([]time.Time{}, rows, func(row pgx.CollectableRow) (time.Time, error) {
				var version time.Time
				if err := row.Scan(&version); err != nil {
					return time.Time{}, err
				}
				return version, nil
			})
			if err != nil {
				return err
			}

			for _, migration := range migrations {
				if slices.ContainsFunc(versions, func(version time.Time) bool {
					return version.Equal(migration.Version)
				}) {
					logger.Debug("migration already exists", "version", migration.Version)
					continue
				}
				logger.Debug("migrating", "version", migration.Version)

				if err := pgx.BeginFunc(ctx, tx, func(tx pgx.Tx) error {
					if _, err := tx.Exec(ctx, migration.Query); err != nil {
						return errors.Wrapf(err, "failed to execute migration. query=%s", migration.Query)
					}

					if _, err := tx.Exec(ctx,
						`INSERT INTO stack.vapi_schema_migrations (version, vapi_package_id) VALUES ($1, $2)`,
						migration.Version,
						v.PackageID,
					); err != nil {
						var pgErr *pgconn.PgError
						if errors.As(err, &pgErr) && pgErr.Code == "23505" { // checking duplicated key violates primary key
							logger.Info("migration already exists", "version", migration.Version)
							return nil
						}
						return errors.Wrapf(err, "failed to insert migration. version=%v", migration.Version)
					}

					return nil
				}); err != nil {
					return err
				}
			}

			return nil
		})
	}); err != nil {
		return err
	}

	return nil
}
