package apidepot

import (
	"context"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/spf13/cobra"
)

var (
	logger = tclog.GetLogger()
)

func (c *Cli) newRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "apidepot",
		Short: "Api Depot",
	}

	cmd.AddCommand(
		c.newServeCmd(),
		c.newScanCmd(),
	)

	return &cmd
}

func Execute(ctx context.Context) error {
	cli := NewCli()

	return cli.newRootCmd().ExecuteContext(ctx)
}
