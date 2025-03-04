package usertest

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/stretchr/testify/mock"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
)

type Service struct {
	mock.Mock
}

func (s *Service) GetStorageUsagesLatest(ctx context.Context) (*user.StorageUsages, error) {
	args := s.Called(ctx)
	return args.Get(0).(*user.StorageUsages), args.Error(1)
}

func (s *Service) GetToken(refreshToken string) (*gotruetypes.Session, error) {
	args := s.Called(refreshToken)
	return args.Get(0).(*gotruetypes.Session), args.Error(1)
}

func (s *Service) UpdateGithubAccessToken(ctx context.Context, accessToken string) error {
	args := s.Called(ctx, accessToken)
	return args.Error(0)
}

func (s *Service) UpdateGithubInstallationId(ctx context.Context, installationId int64) error {
	args := s.Called(ctx, installationId)
	return args.Error(0)
}

func (s *Service) GetUserByAuthUserId(ctx context.Context, ownerID string) (*domain.User, error) {
	args := s.Called(ctx, ownerID)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (s *Service) GetUser(ctx context.Context) (*domain.User, error) {
	args := s.Called(ctx)
	return args.Get(0).(*domain.User), args.Error(1)
}

var _ user.Service = (*Service)(nil)

func NewService() *Service {
	return &Service{}
}
