package project

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Service interface {
	GetProjects(ctx context.Context, input GetProjectsInput) ([]domain.Project, error)
	CreateProject(ctx context.Context, name string, description string) (*domain.Project, error)
	GetProject(ctx context.Context, id uint) (*domain.Project, error)
	DeleteProject(ctx context.Context, id uint) error
}

type service struct {
	users user.Service
}

var (
	logger         = tclog.GetLogger()
	_      Service = (*service)(nil)
)

func NewService(
	users user.Service,
) Service {
	return &service{
		users: users,
	}
}

func (s *service) GetProjects(
	ctx context.Context,
	input GetProjectsInput,
) ([]domain.Project, error) {
	user, err := s.users.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	tx := helpers.GetTx(ctx)

	stmt := tx
	if input.Name != nil {
		stmt = stmt.Where("name = ?", *input.Name)
	}

	if input.Page > 0 && input.PerPage > 0 {
		stmt = stmt.Offset((input.Page - 1) * input.PerPage).Limit(input.PerPage)
	}

	var projects []domain.Project
	if err := stmt.Find(&projects, "owner_id = ?", user.ID).Error; err != nil {
		return nil, err
	}

	logger.Debug("projects", "count", len(projects))

	return projects, nil
}

func (s *service) CreateProject(ctx context.Context, name string, description string) (*domain.Project, error) {
	tx := helpers.GetTx(ctx)

	owner, err := s.users.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "name is required")
	} else if len(name) > 50 {
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "name is too long, max length is 50 characters")
	}

	project := domain.Project{
		Name:        name,
		Description: description,
		OwnerID:     owner.ID,
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		if err := project.Save(tx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *service) GetProject(ctx context.Context, id uint) (*domain.Project, error) {
	user, err := s.users.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	tx := helpers.GetTx(ctx)
	project, err := domain.GetProjectByID(tx.Preload("Stacks"), id)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != user.ID && !user.IsSuperuser() {
		return nil, errors.Wrapf(tclerrors.ErrForbidden, "you are not allowed to access this project")
	}

	return &project, nil
}

func (s *service) DeleteProject(ctx context.Context, id uint) error {
	if _, err := s.GetProject(ctx, id); err != nil {
		return err
	}

	tx := helpers.GetTx(ctx)
	if err := domain.DeleteProjectByID(tx, id); err != nil {
		return err
	}

	return nil
}

const ServiceKey digo.ObjectKey = "projectService"

func init() {
	digo.ProvideService(ServiceKey, func(ctx *digo.Container) (any, error) {
		users, err := digo.Get[user.Service](ctx, user.ServiceKey)
		if err != nil {
			return nil, err
		}

		return NewService(users), nil
	})
}
