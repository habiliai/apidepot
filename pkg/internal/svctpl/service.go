package svctpl

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/habiliai/apidepot/pkg/internal/user"
)

type Service interface {
	SearchServiceTemplates(
		ctx context.Context,
		cursor uint,
		limit uint,
		searchQuery string,
	) (*SearchServiceTemplatesOutput, error)
	GetServiceTemplateByID(
		ctx context.Context,
		id uint,
	) (*domain.ServiceTemplate, error)
	CreateStackFromServiceTemplate(
		ctx context.Context,
		serviceTemplateId uint,
		input stack.CreateStackInput,
	) (*domain.Stack, error)
}

type serviceImpl struct {
	stacks       stack.Service
	users        user.Service
	githubClient services.GithubClient
	gitService   services.GitService
}

var (
	logger                    = tclog.GetLogger()
	_          Service        = (*serviceImpl)(nil)
	ServiceKey digo.ObjectKey = "svctpl.Service"
)

func init() {
	digo.ProvideService(ServiceKey, func(ctx *digo.Container) (any, error) {
		stacks, err := digo.Get[stack.Service](ctx, stack.ServiceKey)
		if err != nil {
			return nil, err
		}

		users, err := digo.Get[user.Service](ctx, user.ServiceKey)
		if err != nil {
			return nil, err
		}

		git, err := digo.Get[services.GitService](ctx, services.ServiceKeyGitService)
		if err != nil {
			return nil, err
		}

		githubClient, err := digo.Get[services.GithubClient](ctx, services.ServiceKeyGithubClient)
		if err != nil {
			return nil, err
		}

		return &serviceImpl{
			stacks:       stacks,
			users:        users,
			gitService:   git,
			githubClient: githubClient,
		}, nil
	})
}
