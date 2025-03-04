package projecttest

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/project"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) CreateProject(ctx context.Context, name string, description string) (*domain.Project, error) {
	args := m.Called(ctx, name, description)
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *ServiceMock) GetVapiPackages(ctx context.Context, projectId uint) ([]domain.VapiPackage, error) {
	args := m.Called(ctx, projectId)
	return args.Get(0).([]domain.VapiPackage), args.Error(1)
}

func (m *ServiceMock) GetProjects(ctx context.Context, input project.GetProjectsInput) ([]domain.Project, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *ServiceMock) GetProject(ctx context.Context, id uint) (*domain.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *ServiceMock) DeleteProject(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

var _ project.Service = (*ServiceMock)(nil)

func NewTestService() *ServiceMock {
	return &ServiceMock{}
}
