package servicestest

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/stretchr/testify/mock"
)

type MockGitService struct {
	mock.Mock
}

func (m *MockGitService) Clone(ctx context.Context, gitUrl string, gitBranch string, optionsFn ...func(*services.CloneOptions)) (*git.Repository, error) {
	args := m.Called(ctx, gitUrl, gitBranch, optionsFn)
	return args.Get(0).(*git.Repository), args.Error(1)
}

func (m *MockGitService) ReadFile(repo *git.Repository, path string) ([]byte, error) {
	args := m.Called(repo, path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockGitService) CommitFile(repo *git.Repository, path string, body []byte, commitMessage string) error {
	args := m.Called(repo, path, body, commitMessage)
	return args.Error(0)
}

func (m *MockGitService) OpenDir(repo *git.Repository, path string) (*object.Tree, error) {
	args := m.Called(repo, path)
	return args.Get(0).(*object.Tree), args.Error(1)
}

func (m *MockGitService) CopyRepo(ctx context.Context, srcGitUrl, dstGitUrl string) error {
	args := m.Called(ctx, srcGitUrl, dstGitUrl)
	return args.Error(0)
}

var (
	_ services.GitService = (*MockGitService)(nil)
)

func NewTestGitService() *MockGitService {
	return new(MockGitService)
}
