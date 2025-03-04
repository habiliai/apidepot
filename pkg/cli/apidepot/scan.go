package apidepot

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/histories"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

func (c *Cli) newScanCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "scan",
		Short: "Scan for running instances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := c.getServerConfig(cmd.Flags())
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeoutCause(cmd.Context(), cfg.Scan.Timeout, tclerrors.ErrTimeout)
			defer cancel()

			container := digo.NewContainer(ctx, digo.EnvProd, cfg)

			historyService, err := digo.Get[histories.Service](container, histories.ServiceKey)
			if err != nil {
				return err
			}

			db, err := digo.Get[*gorm.DB](container, services.ServiceKeyDB)
			if err != nil {
				return err
			}

			return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				ctx = helpers.WithTx(ctx, tx)

				var eg errgroup.Group
				eg.Go(func() error {
					return historyService.WriteInstanceHistoriesAt(ctx)
				})
				eg.Go(func() error {
					return historyService.WriteStackHistoriesAt(ctx)
				})

				return eg.Wait()
			})
		},
	}

	f := cmd.Flags()
	f.Duration("scan.timeout", 5*time.Minute, "Scan timeout")
	f.String("db.host", "localhost", "Database host")
	f.Int("db.port", 6543, "Database port")
	f.String("db.user", "postgres", "Database user")
	f.String("db.password", "postgres", "Database password")
	f.String("db.name", "postgres", "Database name")
	f.String("db.pingTimeout", "5s", "Database ping timeout")
	f.Bool("db.autoMigration", true, "Auto migration")
	f.Int("db.maxIdleConns", 10, "Max idle connections")
	f.Int("db.maxOpenConns", 100, "Max open connections")
	f.String("db.connMaxLifetime", "1h", "Connection max lifetime")

	return &cmd
}
