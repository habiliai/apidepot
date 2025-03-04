package proto

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/cliapp"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/instance"
	"github.com/habiliai/apidepot/pkg/internal/organization"
	"github.com/habiliai/apidepot/pkg/internal/project"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/habiliai/apidepot/pkg/internal/svctpl"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type (
	apiDepotServer struct {
		UnsafeApiDepotServer

		db              *gorm.DB
		vapiService     vapi.Service
		orgService      organization.Service
		userService     user.Service
		githubClient    services.GithubClient
		includeDebug    bool
		stackService    stack.Service
		projectService  project.Service
		instanceService instance.Service
		cliappService   cliapp.Service
		svctplService   svctpl.Service
		gitService      services.GitService
		storageClient   *storage.Client
	}
)

const (
	ServiceKey = "proto.apiDepotServer"
)

var (
	_ ApiDepotServer = (*apiDepotServer)(nil)
)

func (s *apiDepotServer) ResetSchema(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if !s.includeDebug {
		return nil, errors.WithStack(tclerrors.ErrNotFound)
	}

	tx := s.db.WithContext(ctx)

	tx.Exec(`DROP SCHEMA IF EXISTS apidepot CASCADE`)
	if err := domain.AutoMigrate(tx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func init() {
	digo.ProvideService(ServiceKey, func(ctx *digo.Container) (any, error) {
		db, err := digo.Get[*gorm.DB](ctx, services.ServiceKeyDB)
		if err != nil {
			return nil, err
		}

		vapiService, err := digo.Get[vapi.Service](ctx, vapi.ServiceKey)
		if err != nil {
			return nil, err
		}

		orgService, err := digo.Get[organization.Service](ctx, organization.ServiceKey)
		if err != nil {
			return nil, err
		}

		userService, err := digo.Get[user.Service](ctx, user.ServiceKey)
		if err != nil {
			return nil, err
		}

		githubClient, err := digo.Get[services.GithubClient](ctx, services.ServiceKeyGithubClient)
		if err != nil {
			return nil, err
		}

		stackService, err := digo.Get[stack.Service](ctx, stack.ServiceKey)
		if err != nil {
			return nil, err
		}

		projectService, err := digo.Get[project.Service](ctx, project.ServiceKey)
		if err != nil {
			return nil, err
		}

		instanceService, err := digo.Get[instance.Service](ctx, instance.ServiceKey)
		if err != nil {
			return nil, err
		}

		cliappService, err := digo.Get[cliapp.Service](ctx, cliapp.ServiceKey)
		if err != nil {
			return nil, err
		}

		svctplService, err := digo.Get[svctpl.Service](ctx, svctpl.ServiceKey)
		if err != nil {
			return nil, err
		}

		gitService, err := digo.Get[services.GitService](ctx, services.ServiceKeyGitService)
		if err != nil {
			return nil, err
		}

		storageClient, err := digo.Get[*storage.Client](ctx, services.ServiceKeyStorageClient)
		if err != nil {
			return nil, err
		}

		switch ctx.Env {
		case digo.EnvProd:
			return &apiDepotServer{
				db:              db,
				vapiService:     vapiService,
				orgService:      orgService,
				userService:     userService,
				includeDebug:    ctx.Config.IncludeDebug,
				githubClient:    githubClient,
				stackService:    stackService,
				projectService:  projectService,
				instanceService: instanceService,
				cliappService:   cliappService,
				svctplService:   svctplService,
				gitService:      gitService,
				storageClient:   storageClient,
			}, nil
		case digo.EnvTest:
			return &apiDepotServer{
				db:              db,
				vapiService:     vapiService,
				includeDebug:    true,
				orgService:      orgService,
				userService:     userService,
				githubClient:    githubClient,
				stackService:    stackService,
				projectService:  projectService,
				instanceService: instanceService,
				cliappService:   cliappService,
				svctplService:   svctplService,
				gitService:      gitService,
				storageClient:   storageClient,
			}, nil
		default:
			return nil, errors.Errorf("unknown env: %s", ctx.Env)
		}
	})
}
