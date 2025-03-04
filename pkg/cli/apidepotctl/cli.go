package apidepotctl

import (
	"context"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var (
	logger = tclog.GetLogger()
)

type (
	Cli struct {
		v    *viper.Viper
		args CliArgs

		conn *grpc.ClientConn

		// testServerUrl is used for testing gotrue client
		testServerUrl string
	}

	CliInitOptions struct {
		verifyCli       bool
		connectApiDepot bool
		readConfig      bool
	}
)

func NewCli(testServerUrl string) *Cli {
	v := viper.New()
	v.SetConfigName("apidepot")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/apidepot")
	v.AddConfigPath("/etc/apidepot")
	if err := v.ReadInConfig(); err != nil {
		logger.Warn("failed to read in config", tclog.Err(err))
	}

	return &Cli{
		v:             v,
		testServerUrl: testServerUrl,
	}
}

func (c *Cli) connectApiDepot() error {
	ssl := !strings.Contains(c.args.Server, "local.shaple.io") && !strings.Contains(c.args.Server, "localhost") && !strings.Contains(c.args.Server, "127.0.0.1")
	conn, err := proto.NewClient(c.args.Server, ssl, c.args.Timeout)
	if err != nil {
		return errors.WithStack(err)
	}

	c.conn = conn
	return nil
}

func (c *Cli) close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *Cli) verifyCli(ctx context.Context) (context.Context, error) {
	conf := c.args.Session
	if conf.AppId == "" || conf.AppSecret == "" {
		return ctx, errors.New("app_id and app_secret are required")
	}

	tcc := proto.NewApiDepotClient(c.conn)
	verifyCliAppResp, err := tcc.VerifyCliApp(ctx, &proto.VerifyCliAppRequest{
		AppId:     conf.AppId,
		AppSecret: conf.AppSecret,
	})
	if err != nil {
		return ctx, errors.WithStack(err)
	}

	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+verifyCliAppResp.AccessToken), nil
}

func (c *Cli) writeConfig() error {
	if !c.args.Save {
		return nil
	}

	bin, err := yaml.Marshal(c.args.CliConfig)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := os.WriteFile(c.args.ConfigFile, bin, 0644); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *Cli) readConfig(flags *pflag.FlagSet) error {
	if err := c.v.BindPFlags(flags); err != nil {
		return errors.WithStack(err)
	}

	if c.args.ConfigFile != "" {
		c.v.SetConfigFile(c.args.ConfigFile)
		if err := c.v.MergeInConfig(); err != nil {
			logger.Warn("not existed config file", tclog.Err(err))
		} else {
			logger.Info("read config", "filepath", c.args.ConfigFile)
		}
	}

	if err := c.v.Unmarshal(&c.args.CliConfig); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *Cli) init(ctx context.Context, flags *pflag.FlagSet, options ...CliInitOptions) (context.Context, error) {
	if len(options) >= 2 {
		return ctx, errors.New("too many options")
	}

	var err error
	option := options[0]
	if option.readConfig {
		if err := c.readConfig(flags); err != nil {
			return ctx, err
		}
	}
	if option.connectApiDepot {
		if err := c.connectApiDepot(); err != nil {
			return ctx, err
		}
	}
	if option.verifyCli {
		if ctx, err = c.verifyCli(ctx); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}
