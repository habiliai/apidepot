package servicestest

import (
	"context"
	"github.com/google/go-github/v60/github"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/stretchr/testify/mock"
)

var _ services.GithubClient = (*MockGithubClient)(nil)

type MockGithubClient struct {
	mock.Mock
}

func (m *MockGithubClient) CreateRepository(ctx context.Context, accessToken string, org string, repoName string, description string) error {
	args := m.Called(ctx, accessToken, org, repoName, description)
	return args.Error(0)
}

func (m *MockGithubClient) ForkRepo(ctx context.Context, accessToken, srcGitRepo, name string, organization *string) (*github.Repository, error) {
	args := m.Called(ctx, accessToken, srcGitRepo, name, organization)
	return args.Get(0).(*github.Repository), args.Error(1)
}

func (m *MockGithubClient) DeleteRepo(ctx context.Context, accessToken string, gitRepo string) error {
	args := m.Called(ctx, accessToken, gitRepo)
	return args.Error(0)
}

func (m *MockGithubClient) GenerateInstallationAccessToken(ctx context.Context, installationId int64) (string, error) {
	args := m.Called(ctx, installationId)
	return args.String(0), args.Error(1)
}

func (m *MockGithubClient) GetExistingInstallationId(ctx context.Context, accessToken string) (int64, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGithubClient) FetchUserAccessToken(ctx context.Context, authorizationCode string) (string, error) {
	args := m.Called(ctx, authorizationCode)
	return args.String(0), args.Error(1)
}

func (m *MockGithubClient) VerifyInstallation(ctx context.Context, accessToken string, installationId int64) (bool, error) {
	args := m.Called(ctx, accessToken, installationId)
	return args.Bool(0), args.Error(1)
}

func (m *MockGithubClient) CreateSSHKey(ctx context.Context, token string, title string, publicKey string) (*github.Key, error) {
	args := m.Called(ctx, token, title, publicKey)
	return args.Get(0).(*github.Key), args.Error(1)
}

func (m *MockGithubClient) DeleteSSHKey(ctx context.Context, token string, keyId int64) error {
	args := m.Called(ctx, token, keyId)
	return args.Error(0)
}

func (m *MockGithubClient) GetUser(ctx context.Context, token string) (*github.User, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*github.User), args.Error(1)
}

func NewTestGithubClient() *MockGithubClient {
	return &MockGithubClient{}
}
