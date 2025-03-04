package apidepotctl

import (
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cli) newStackCustomVapiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-vapi",
		Short: "Manage custom vapi",
	}

	cmd.AddCommand(c.newStackCustomVapiListCmd())
	cmd.AddCommand(c.newStackCustomVapiCreateCmd())
	cmd.AddCommand(c.newStackCustomVapiDeleteCmd())

	return cmd
}

func (c *Cli) newStackCustomVapiListCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List custom vapi",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			if ctx, err = c.init(ctx, cmd.Flags(), CliInitOptions{
				verifyCli:       true,
				readConfig:      true,
				connectApiDepot: true,
			}); err != nil {
				return err
			} else {
				logger.Debug("list custom vapi")
			}

			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}
			tcc := proto.NewApiDepotClient(c.conn)

			if resp, err := tcc.GetCustomVapisOnStack(ctx, &proto.StackId{
				Id: st.Id,
			}); err != nil {
				return errors.WithStack(err)
			} else {
				for _, vapi := range resp.CustomVapis {
					println("custom vapi{ name:", vapi.Name, "}")
				}
			}

			return nil
		},
	}

	return &cmd
}

func (c *Cli) newStackCustomVapiCreateCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "install NAME",
		Short: "install custom vapi",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			if len(args) < 1 {
				return errors.Errorf("missing NAME")
			} else if ctx, err = c.init(ctx, cmd.Flags(), CliInitOptions{
				verifyCli:       true,
				readConfig:      true,
				connectApiDepot: true,
			}); err != nil {
				return
			} else {
				logger.Debug("create custom vapi", "args", args)
			}

			customVapiName := args[0]
			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}

			tcc := proto.NewApiDepotClient(c.conn)
			if _, err = tcc.InstallCustomVapi(ctx, &proto.InstallCustomVapiRequest{
				StackId: st.Id,
				Name:    customVapiName,
			}); err != nil {
				return err
			}
			return nil
		},
	}

	return &cmd
}

func (c *Cli) newStackCustomVapiDeleteCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete NAME",
		Short: "Delete custom vapi",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			if len(args) < 1 {
				return errors.Errorf("missing NAME")
			} else if ctx, err = c.init(ctx, cmd.Flags(), CliInitOptions{
				verifyCli:       true,
				readConfig:      true,
				connectApiDepot: true,
			}); err != nil {
				return err
			} else {
				logger.Debug("delete custom vapi", "args", args)
			}

			customVapiName := args[0]
			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}
			tcc := proto.NewApiDepotClient(c.conn)
			if _, err = tcc.UninstallCustomVapi(ctx, &proto.UninstallCustomVapiRequest{
				StackId: st.Id,
				Name:    customVapiName,
			}); err != nil {
				return err
			}

			return nil
		},
	}

	return &cmd
}
