package digo

import (
	"context"
	"github.com/habiliai/apidepot/pkg/config"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
)

type (
	Env       string
	Container struct {
		context.Context
		Env    Env
		Config *config.ApiDepotServerConfig

		objects map[ObjectKey]any
	}
)

const (
	EnvProd = "prod"
	EnvTest = "test"
)

var (
	logger = tclog.GetLogger()
)

func NewContainer(
	ctx context.Context,
	env Env,
	cfg *config.ApiDepotServerConfig,
) *Container {
	return &Container{
		Context: ctx,
		Env:     env,
		Config:  cfg,
		objects: map[ObjectKey]any{},
	}
}
