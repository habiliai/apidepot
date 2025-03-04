package apidepotctl

import (
	"context"
	"github.com/spf13/cobra"
	"time"
)

func (c *Cli) NewRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "apidepotctl",
		Short: "ApiDepot Control CLI",
	}

	f := cmd.PersistentFlags()
	f.StringP("server", "s", "apidepot.local.shaple.io", "api-depot server")
	f.StringP("apiKey", "k", "", "API key")
	f.DurationP("timeout", "t", 30*time.Second, "Timeout for request")
	f.StringVarP(&c.args.ConfigFile, "config-file", "f", "./apidepot.yml", "Path to config file")
	f.BoolVarP(&c.args.Save, "save", "S", true, "Save config file")

	cmd.AddCommand(
		c.newLoginCmd(),
		c.newLogoutCmd(),
		c.newStackCmd(),
	)

	return &cmd
}

func Execute(ctx context.Context) error {
	cli := NewCli("")

	return cli.NewRootCmd().ExecuteContext(ctx)
}
