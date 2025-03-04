package cliapp

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/supabase-community/gotrue-go"
)

type Service interface {
	RegisterCliApp(
		ctx context.Context,
		host string,
		refreshToken string,
	) (*AddCliAppOutput, error)
	VerifyCliApp(
		ctx context.Context,
		appId string,
		appSecret string,
	) (*VerifyCliAppResponse, error)
	DeleteCliApp(
		ctx context.Context,
		appId string,
	) error
}

type service struct {
	users        user.Service
	gotrueClient gotrue.Client
}

var (
	_          Service        = (*service)(nil)
	logger                    = tclog.GetLogger()
	ServiceKey digo.ObjectKey = "cliapp.Service"
)

func init() {
	digo.ProvideService(ServiceKey, func(ctx *digo.Container) (interface{}, error) {
		users, err := digo.Get[user.Service](ctx, user.ServiceKey)
		if err != nil {
			return nil, err
		}

		gotrueClient, err := digo.Get[gotrue.Client](ctx, services.ServiceKeyGoTrueClient)
		if err != nil {
			return nil, err
		}

		logger.Debug("new cliapp service")

		return &service{
			users:        users,
			gotrueClient: gotrueClient,
		}, nil
	})
}
