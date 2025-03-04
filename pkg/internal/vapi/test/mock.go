package vapitest

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (s *ServiceMock) GetDBMigrations(ctx context.Context, vapiRel domain.VapiRelease) ([]vapi.Migration, error) {
	args := s.Called(ctx, vapiRel)
	return args.Get(0).([]vapi.Migration), args.Error(1)
}

func (s *ServiceMock) GetAllDependenciesOfVapiReleases(ctx context.Context, vapiReleases []domain.VapiRelease) ([]domain.VapiRelease, error) {
	args := s.Called(ctx, vapiReleases)
	return args.Get(0).([]domain.VapiRelease), args.Error(1)
}

func (s *ServiceMock) IsGrantedVapi(ctx context.Context, stackId uint, rel *domain.VapiRelease) error {
	args := s.Called(ctx, stackId, rel)
	return args.Error(0)
}

func (s *ServiceMock) Register(ctx context.Context, gitRepo string, gitBranch string, name string, description string, domains []string, vapiPoolId string, homepage string) (*domain.VapiRelease, error) {
	args := s.Called(ctx, gitRepo, gitBranch, name, description, domains, vapiPoolId, homepage)
	return args.Get(0).(*domain.VapiRelease), args.Error(1)
}

func (s *ServiceMock) GetPackagesByOwnerId(ctx context.Context, ownerId uint) ([]domain.VapiPackage, error) {
	args := s.Called(ctx, ownerId)
	return args.Get(0).([]domain.VapiPackage), args.Error(1)
}

func (s *ServiceMock) GetPackages(ctx context.Context, input vapi.GetPackagesInput) ([]domain.VapiPackage, error) {
	args := s.Called(ctx, input)
	return args.Get(0).([]domain.VapiPackage), args.Error(1)
}

func (s *ServiceMock) GetReleaseByVersionInPackage(ctx context.Context, packageId uint, version string) (*domain.VapiRelease, error) {
	args := s.Called(ctx, packageId, version)
	return args.Get(0).(*domain.VapiRelease), args.Error(1)
}

func (s *ServiceMock) DeleteAllReleases(ctx context.Context) error {
	args := s.Called(ctx)
	return args.Error(0)
}

func (s *ServiceMock) DeleteReleasesByPackageId(ctx context.Context, packageId uint) error {
	args := s.Called(ctx, packageId)
	return args.Error(0)
}

func (s *ServiceMock) DeleteAllPackages(ctx context.Context, projectId uint) error {
	args := s.Called(ctx, projectId)
	return args.Error(0)
}

func (s *ServiceMock) GetRelease(ctx context.Context, id uint) (*domain.VapiRelease, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(*domain.VapiRelease), args.Error(1)
}

func (s *ServiceMock) DeleteRelease(ctx context.Context, id uint) error {
	args := s.Called(ctx, id)
	return args.Error(0)
}

func (s *ServiceMock) GetPackage(ctx context.Context, id uint) (*domain.VapiPackage, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(*domain.VapiPackage), args.Error(1)
}

func (s *ServiceMock) DeletePackage(ctx context.Context, id uint) error {
	args := s.Called(ctx, id)
	return args.Error(0)
}

func (s *ServiceMock) SearchVapis(ctx context.Context, input vapi.SearchVapisInput) (vapi.SearchVapisOutput, error) {
	args := s.Called(ctx, input)
	return args.Get(0).(vapi.SearchVapisOutput), args.Error(1)
}

func NewService() *ServiceMock {
	return &ServiceMock{}
}

var _ vapi.Service = (*ServiceMock)(nil)
