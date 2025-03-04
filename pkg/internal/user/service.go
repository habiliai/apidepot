package user

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/histories"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/pkg/errors"
	"github.com/supabase-community/gotrue-go"
)

type Service interface {
	GetUser(
		ctx context.Context,
	) (*domain.User, error)
	GetStorageUsagesLatest(
		ctx context.Context,
	) (*StorageUsages, error)
	GetUserByAuthUserId(
		ctx context.Context,
		ownerID string,
	) (*domain.User, error)
	UpdateGithubAccessToken(
		ctx context.Context,
		accessToken string,
	) error
	UpdateGithubInstallationId(
		ctx context.Context,
		installationId int64,
	) error
}

type service struct {
	gotrueClient     gotrue.Client
	githubClient     services.GithubClient
	storageClient    *storage.Client
	historiesService histories.Service
}

const (
	ServiceKey = "user"
)

var logger = tclog.GetLogger()

func init() {
	digo.ProvideService(ServiceKey, func(container *digo.Container) (any, error) {
		githubClient, err := digo.Get[services.GithubClient](container, services.ServiceKeyGithubClient)
		if err != nil {
			return nil, err
		}

		gotrueClient, err := digo.Get[gotrue.Client](container, services.ServiceKeyGoTrueClient)
		if err != nil {
			return nil, err
		}

		storageClient, err := digo.Get[*storage.Client](container, services.ServiceKeyStorageClient)
		if err != nil {
			return nil, err
		}

		historiesService, err := digo.Get[histories.Service](container, histories.ServiceKey)
		if err != nil {
			return nil, err
		}

		logger.Debug("new user service")

		switch container.Env {
		case digo.EnvTest:
			fallthrough
		case digo.EnvProd:
			return &service{
				gotrueClient:     gotrueClient,
				githubClient:     githubClient,
				storageClient:    storageClient,
				historiesService: historiesService,
			}, nil
		}

		return nil, errors.New("unknown env")
	})
}
