package instancetest

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/instance"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (s *ServiceMock) CreateInstance(ctx context.Context, input instance.CreateInstanceInput) (*domain.Instance, error) {
	args := s.Called(ctx, input)
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (s *ServiceMock) DeployStack(ctx context.Context, instanceId uint, input instance.DeployStackInput) error {
	args := s.Called(ctx, instanceId, input)
	return args.Error(0)
}

func (s *ServiceMock) LaunchInstance(ctx context.Context, instanceId uint) error {
	args := s.Called(ctx, instanceId)
	return args.Error(0)
}

func (s *ServiceMock) StopInstance(ctx context.Context, instanceId uint) error {
	args := s.Called(ctx, instanceId)
	return args.Error(0)
}

func (s *ServiceMock) DeleteInstance(ctx context.Context, instanceId uint, force bool) error {
	args := s.Called(ctx, instanceId, force)
	return args.Error(0)
}

func (s *ServiceMock) RestartInstance(ctx context.Context, instanceId uint) error {
	args := s.Called(ctx, instanceId)
	return args.Error(0)
}

func (s *ServiceMock) EditInstance(ctx context.Context, instanceId uint, input instance.EditInstanceInput) error {
	args := s.Called(ctx, instanceId, input)
	return args.Error(0)
}

func (s *ServiceMock) GetInstance(ctx context.Context, instanceId uint) (*domain.Instance, error) {
	args := s.Called(ctx, instanceId)
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (s *ServiceMock) GetInstancesInStack(ctx context.Context, stackId uint) ([]domain.Instance, error) {
	args := s.Called(ctx, stackId)
	return args.Get(0).([]domain.Instance), args.Error(1)
}

var (
	_ instance.Service = (*ServiceMock)(nil)
)
