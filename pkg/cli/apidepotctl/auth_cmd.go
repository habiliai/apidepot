package apidepotctl

import (
	"bufio"
	"fmt"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/supabase-community/gotrue-go"
	"golang.org/x/term"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"os"
)

func (c *Cli) newLoginCmd() *cobra.Command {
	loginCmd := cobra.Command{
		Use:   "login",
		Short: "Login to APIDepot",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer c.close()
			ctx := cmd.Context()
			if err = c.readConfig(cmd.Flags()); err != nil {
				return err
			}

			var email string
			if c.args.Login.Email == "" {
				fmt.Print("Enter email: ")
				email, err = bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil {
					return errors.Wrapf(err, "failed to read email from stdin")
				}
			} else {
				email = c.args.Login.Email
			}

			var password string
			if c.args.Login.Password != "" {
				password = c.args.Login.Password
			} else {
				fmt.Print("Enter password: ")
				passwordBin, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return errors.Wrapf(err, "failed to read password from terminal")
				}
				password = string(passwordBin)
			}

			logger.Debug("print", "url", c.getServerUrl(), "server", c.args.Server, "server-arg", c.v.GetString("server"))

			gotrue := gotrue.New("", c.args.ApiKey).WithCustomGoTrueURL(fmt.Sprintf("%s/auth/v1", c.getServerUrl()))
			session, err := gotrue.SignInWithEmailPassword(email, password)
			if err != nil {
				return errors.Wrapf(err, "failed to sign in with email and password. email=%s", email)
			}
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+session.AccessToken)

			if err = c.connectApiDepot(); err != nil {
				return err
			}

			tcc := proto.NewApiDepotClient(c.conn)
			user, err := tcc.GetUser(ctx, &emptypb.Empty{})
			if err != nil {
				return errors.WithStack(err)
			}
			username := user.Profile.Name

			fmt.Printf("👋Welcome, %s!\n", username)

			hostname, err := os.Hostname()
			if err != nil {
				hostname = ""
			}

			if resp, err := tcc.RegisterCliApp(ctx, &proto.RegisterCliAppRequest{
				Host:         hostname,
				RefreshToken: session.RefreshToken,
			}); err != nil {
				return errors.WithStack(err)
			} else {
				c.args.Session.AppId = resp.AppId
				c.args.Session.AppSecret = resp.AppSecret
			}

			return c.writeConfig()
		},
	}

	f := loginCmd.Flags()
	f.StringVarP(&c.args.Login.Email, "email", "e", "", "Email address")
	f.StringVarP(&c.args.Login.Password, "password", "p", "", "Password")

	return &loginCmd
}

func (c *Cli) newLogoutCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "logout",
		Short: "Logout from APIDepot",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			defer c.close()

			if err = c.readConfig(cmd.Flags()); err != nil {
				return errors.WithStack(err)
			}
			if err = c.connectApiDepot(); err != nil {
				return
			}

			tcc := proto.NewApiDepotClient(c.conn)
			ctx, err = c.verifyCli(ctx)
			if err != nil {
				return errors.WithStack(err)
			}

			if _, err := tcc.DeleteCliApp(ctx, &proto.DeleteCliAppRequest{
				AppId: c.args.Session.AppId,
			}); err != nil {
				return errors.WithStack(err)
			}

			c.args.Session.AppId = ""
			c.args.Session.AppSecret = ""
			return c.writeConfig()
		},
	}

	f := cmd.Flags()
	f.String("session.appId", "", "App ID")
	f.String("session.appSecret", "", "App secret")

	return &cmd
}
