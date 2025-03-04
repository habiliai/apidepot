package vapi

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/pkg/errors"
)

type Service interface {
	Register(
		ctx context.Context,
		gitRepo string,
		gitBranch string,
		name string,
		description string,
		domains []string,
		vapiPoolId string,
		homepage string,
	) (*domain.VapiRelease, error)
	GetDBMigrations(
		ctx context.Context,
		vapiRel domain.VapiRelease,
	) ([]Migration, error)
	SearchVapis(
		ctx context.Context,
		input SearchVapisInput,
	) (SearchVapisOutput, error)
	GetPackage(
		ctx context.Context,
		id uint,
	) (*domain.VapiPackage, error)
	DeletePackage(
		ctx context.Context,
		id uint,
	) error
	GetRelease(
		ctx context.Context,
		id uint,
	) (*domain.VapiRelease, error)
	GetReleaseByVersionInPackage(
		ctx context.Context,
		packageId uint,
		version string,
	) (*domain.VapiRelease, error)
	DeleteRelease(
		ctx context.Context,
		id uint,
	) error
	DeleteAllReleases(
		ctx context.Context,
	) error
	DeleteAllPackages(
		ctx context.Context,
		projectId uint,
	) error
	DeleteReleasesByPackageId(
		ctx context.Context,
		packageId uint,
	) error
	GetPackages(ctx context.Context, input GetPackagesInput) ([]domain.VapiPackage, error)
	GetPackagesByOwnerId(ctx context.Context, ownerId uint) ([]domain.VapiPackage, error)
	GetAllDependenciesOfVapiReleases(
		ctx context.Context,
		vapiReleases []domain.VapiRelease,
	) ([]domain.VapiRelease, error)
}

type service struct {
	storage      *storage.Client
	users        user.Service
	git          services.GitService
	githubClient services.GithubClient
}

const ServiceKey = "apidepot.vapis"

var logger = tclog.GetLogger()

func NewService(
	storageClient *storage.Client,
	userService user.Service,
	gitService services.GitService,
	client services.GithubClient,
) (Service, error) {
	return &service{
		storage:      storageClient,
		users:        userService,
		git:          gitService,
		githubClient: client,
	}, nil
}

func init() {
	digo.ProvideService(ServiceKey, func(serviceContainer *digo.Container) (interface{}, error) {
		logger.Info("new Service")

		storageClient, err := digo.Get[*storage.Client](serviceContainer, services.ServiceKeyStorageClient)
		if err != nil {
			return nil, err
		}

		userService, err := digo.Get[user.Service](serviceContainer, user.ServiceKey)
		if err != nil {
			return nil, err
		}

		gitService, err := digo.Get[services.GitService](serviceContainer, services.ServiceKeyGitService)
		if err != nil {
			return nil, err
		}

		githubClient, err := digo.Get[services.GithubClient](serviceContainer, services.ServiceKeyGithubClient)
		if err != nil {
			return nil, err
		}
		switch serviceContainer.Env {
		case digo.EnvProd, digo.EnvTest:
			return NewService(
				storageClient,
				userService,
				gitService,
				githubClient,
			)
		default:
			return nil, errors.New("unknown env")
		}
	})
}
