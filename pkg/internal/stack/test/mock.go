package stacktest

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (s *ServiceMock) EnableCustomVapi(ctx context.Context, stackId uint, input stack.CreateCustomVapiInput) (*domain.CustomVapi, error) {
	args := s.Called(ctx, stackId, input)
	return args.Get(0).(*domain.CustomVapi), args.Error(1)
}

func (s *ServiceMock) UpdateCustomVapi(ctx context.Context, stackId uint, name string, input stack.UpdateCustomVapiInput) error {
	args := s.Called(ctx, stackId, name, input)
	return args.Error(0)
}

func (s *ServiceMock) DisableCustomVapi(ctx context.Context, stackId uint, name string) error {
	args := s.Called(ctx, stackId, name)
	return args.Error(0)
}

func (s *ServiceMock) GetCustomVapiByName(ctx context.Context, stackId uint, name string) (*domain.CustomVapi, error) {
	args := s.Called(ctx, stackId, name)
	return args.Get(0).(*domain.CustomVapi), args.Error(1)
}

func (s *ServiceMock) GetCustomVapis(ctx context.Context, stackId uint) ([]domain.CustomVapi, error) {
	args := s.Called(ctx, stackId)
	return args.Get(0).([]domain.CustomVapi), args.Error(1)
}

func (s *ServiceMock) SetVapiEnv(ctx context.Context, stackId uint, env map[string]string) error {
	args := s.Called(ctx, stackId, env)
	return args.Error(0)
}

func (s *ServiceMock) UnsetVapiEnv(ctx context.Context, stackId uint, names []string) error {
	args := s.Called(ctx, stackId, names)
	return args.Error(0)
}

func (s *ServiceMock) GetStorageUsage(ctx context.Context, stackId uint) (int64, error) {
	args := s.Called(ctx, stackId)
	return args.Get(0).(int64), args.Error(1)
}

func (s *ServiceMock) GetMyTotalStorageUsage(ctx context.Context) (int64, error) {
	args := s.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (s *ServiceMock) GetStacks(ctx context.Context, projectId uint, name *string, cursor uint, limit int) ([]domain.Stack, error) {
	args := s.Called(ctx, projectId, name, cursor, limit)

	return args.Get(0).([]domain.Stack), args.Error(1)
}

func (s *ServiceMock) CreateStack(ctx context.Context, input stack.CreateStackInput) (*domain.Stack, error) {
	args := s.Called(ctx, input)

	return args.Get(0).(*domain.Stack), args.Error(1)
}

func (s *ServiceMock) PatchStack(ctx context.Context, id uint, input stack.PatchStackInput) error {
	args := s.Called(ctx, id, input)

	return args.Error(0)
}

func (s *ServiceMock) DeleteStack(ctx context.Context, u uint) error {
	args := s.Called(ctx, u)

	return args.Error(0)
}

func (s *ServiceMock) EnableVapi(ctx context.Context, stackId uint, input stack.EnableVapiInput) (*domain.StackVapi, error) {
	args := s.Called(ctx, stackId, input)

	return args.Get(0).(*domain.StackVapi), args.Error(1)
}

func (s *ServiceMock) UpdateVapi(ctx context.Context, stackId uint, vapidId uint, input stack.UpdateVapiInput) (*domain.StackVapi, error) {
	args := s.Called(ctx, stackId, vapidId, input)
	return args.Get(0).(*domain.StackVapi), args.Error(1)
}

func (s *ServiceMock) DisableVapi(ctx context.Context, stackId uint, vapiId uint) error {
	args := s.Called(ctx, stackId, vapiId)

	return args.Error(0)
}

func (s *ServiceMock) GetStack(ctx context.Context, id uint) (*domain.Stack, error) {
	args := s.Called(ctx, id)

	return args.Get(0).(*domain.Stack), args.Error(1)
}

func (s *ServiceMock) MigrateDatabase(ctx context.Context, id uint, input stack.MigrateDatabaseInput) error {
	args := s.Called(ctx, id, input)

	return args.Error(0)
}

func (s *ServiceMock) WaitForAvailable(ctx context.Context, id uint, types []stack.ShapleServiceType) error {
	args := s.Called(ctx, id, types)

	return args.Error(0)
}

func (s *ServiceMock) GetStackStatus(ctx context.Context, id uint) (stack.GetStatusOutput, error) {
	args := s.Called(ctx, id)

	return args.Get(0).(stack.GetStatusOutput), args.Error(1)
}

func (s *ServiceMock) EnableOrUpdateAuth(ctx context.Context, id uint, input stack.EnableOrUpdateAuthInput, b bool) error {
	args := s.Called(ctx, id, input, b)

	return args.Error(0)
}

func (s *ServiceMock) DisableAuth(ctx context.Context, u uint) error {
	args := s.Called(ctx, u)

	return args.Error(0)
}

func (s *ServiceMock) EnableOrUpdateStorage(ctx context.Context, u uint, input stack.EnableOrUpdateStorageInput, b bool) error {
	args := s.Called(ctx, u, input, b)

	return args.Error(0)
}

func (s *ServiceMock) DisableStorage(ctx context.Context, u uint) error {
	args := s.Called(ctx, u)

	return args.Error(0)
}

func (s *ServiceMock) EnableOrUpdatePostgrest(ctx context.Context, u uint, input stack.EnableOrUpdatePostgrestInput, b bool) error {
	args := s.Called(ctx, u, input, b)

	return args.Error(0)
}

func (s *ServiceMock) DisablePostgrest(ctx context.Context, u uint) error {
	args := s.Called(ctx, u)

	return args.Error(0)
}

var _ stack.Service = (*ServiceMock)(nil)

func NewTestService() *ServiceMock {
	return &ServiceMock{}
}
