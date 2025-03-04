package apidepot

import (
	"github.com/habiliai/apidepot/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Cli struct {
}

func (c *Cli) getServerConfig(flags *pflag.FlagSet) (*config.ApiDepotServerConfig, error) {
	v := viper.New()
	if err := v.BindPFlags(flags); err != nil {
		return nil, errors.WithStack(err)
	}

	var cfg config.ApiDepotServerConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.WithStack(err)
	}

	logger.Debug("parse config", "config", cfg)

	return &cfg, nil
}

func NewCli() *Cli {
	return &Cli{}
}
