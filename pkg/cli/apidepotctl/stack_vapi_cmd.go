package apidepotctl

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cli) newStackVapiCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "vapi",
		Short: "Manage stack vapi",
	}

	f := cmd.PersistentFlags()
	f.StringVarP(&c.args.StackVapi.Name, "vapi.name", "m", "", "Specify env var for vapi")

	cmd.AddCommand(
		c.newInstallStackVapiCmd(),
		c.newUninstallStackVapiCmd(),
	)

	return &cmd
}

func (c *Cli) newInstallStackVapiCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "install",
		Short: "Install stack vapi",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer c.close()
			ctx := cmd.Context()

			ctx, err = c.verifyCli(ctx)
			if err != nil {
				return
			}

			if err = c.readConfig(cmd.Flags()); err != nil {
				return
			}

			if err = c.connectApiDepot(); err != nil {
				return
			}

			tcc := proto.NewApiDepotClient(c.conn)
			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}

			vapiRelease, err := c.getVapiRelease(ctx, c.args.StackVapi.Name, c.args.StackVapi.Version)
			if err != nil {
				return err
			}

			if _, err := tcc.InstallVapi(ctx, &proto.InstallVapiRequest{
				StackId: st.Id,
				VapiId:  vapiRelease.Id,
			}); err != nil {
				return errors.WithStack(err)
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&c.args.StackVapi.Version, "vapi.version", "v", "latest", "Specify version for vapi")

	return &cmd
}

func (c *Cli) newUninstallStackVapiCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "install",
		Short: "Install stack vapi",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer c.close()
			ctx := cmd.Context()

			if ctx, err = c.verifyCli(ctx); err != nil {
				return
			} else if err = c.readConfig(cmd.Flags()); err != nil {
				return
			} else if err = c.connectApiDepot(); err != nil {
				return
			} else {
				logger.Debug("succeeded cli init")
			}

			tcc := proto.NewApiDepotClient(c.conn)
			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}

			vapiRelease, err := c.getVapiRelease(ctx, c.args.StackVapi.Name, c.args.StackVapi.Version)
			if err != nil {
				return err
			}

			if _, err := tcc.UninstallVapi(ctx, &proto.UninstallVapiRequest{
				StackId: st.Id,
				VapiId:  vapiRelease.Id,
			}); err != nil {
				return errors.WithStack(err)
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&c.args.StackVapi.Version, "vapi.version", "v", "latest", "Specify version for vapi")

	return &cmd
}

func (c *Cli) getVapiRelease(
	ctx context.Context,
	name, version string,
) (*proto.VapiRelease, error) {
	tcc := proto.NewApiDepotClient(c.conn)
	var vapiPackage *proto.VapiPackage
	if resp, err := tcc.GetVapiPackages(ctx, &proto.GetVapiPackagesRequest{
		Name: &name,
	}); err != nil {
		return nil, errors.WithStack(err)
	} else if len(resp.Packages) == 0 {
		return nil, errors.Errorf("vapi package %s not found", name)
	} else {
		vapiPackage = resp.Packages[0]
	}

	vapiRelease, err := tcc.GetVapiReleaseByVersionInPackage(ctx, &proto.GetVapiReleaseByVersionInPackageRequest{
		PackageId: vapiPackage.Id,
		Version:   version,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return vapiRelease, nil
}
