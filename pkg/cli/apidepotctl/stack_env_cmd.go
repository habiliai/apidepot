package apidepotctl

import (
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cli) newStackEnvCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "env",
		Short: "Manage stack vapi's environment variables",
	}

	cmd.AddCommand(
		c.newSetStackEnvCmd(),
		c.newUnsetStackEnvCmd(),
	)

	return &cmd
}

func (c *Cli) newSetStackEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set KEY=VALUE [...KEY=VALUE]",
		Short: "Set stack vapi's environment variable",
		Long: `Set stack vapi's environment variable

key has prefix for VAPI name. e.g. "user-management.MAX_USERS=1000"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer c.close()
			ctx := cmd.Context()

			if err = c.readConfig(cmd.Flags()); err != nil {
				return
			} else if err = c.connectApiDepot(); err != nil {
				return
			} else if ctx, err = c.verifyCli(ctx); err != nil {
				return
			} else {
				logger.Debug("set stack env", "args", args)
			}

			var userEnvVars []*proto.SetStackEnvRequest_EnvVar
			for _, arg := range args {
				key, value := util.SplitStringToPair(arg, "=")
				if key == "" || value == "" {
					return errors.Errorf("invalid key-value pair: %s", arg)
				}
				userEnvVars = append(userEnvVars, &proto.SetStackEnvRequest_EnvVar{
					Name:  key,
					Value: value,
				})
			}

			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}

			tcc := proto.NewApiDepotClient(c.conn)
			if _, err := tcc.SetStackEnv(ctx, &proto.SetStackEnvRequest{
				StackId: st.Id,
				EnvVars: userEnvVars,
			}); err != nil {
				return errors.WithStack(err)
			}

			return nil
		},
	}
}

func (c *Cli) newUnsetStackEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset KEY [...KEY]",
		Short: "Unset stack vapi's environment variable",
		Long: `Unset stack vapi's environment variable

key has prefix for VAPI name. e.g. "user-management.MAX_USERS"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer c.close()
			ctx := cmd.Context()

			if err = c.readConfig(cmd.Flags()); err != nil {
				return
			} else if err = c.connectApiDepot(); err != nil {
				return
			} else if ctx, err = c.verifyCli(ctx); err != nil {
				return
			} else {
				logger.Debug("unset stack env", "args", args)
			}

			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}

			tcc := proto.NewApiDepotClient(c.conn)
			if _, err := tcc.UnsetStackEnv(ctx, &proto.UnsetStackEnvRequest{
				StackId:     st.Id,
				EnvVarNames: args,
			}); err != nil {
				return errors.WithStack(err)
			}

			return nil
		},
	}
}
