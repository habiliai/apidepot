package services_test

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/spf13/afero/tarfs"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

func TestGitService(t *testing.T) {
	suite.Run(t, new(GitServiceTestSuite))
}

type GitServiceTestSuite struct {
	suite.Suite
	context.Context

	cancel context.CancelFunc
	git    services.GitService
}

func (s *GitServiceTestSuite) SetupTest() {
	s.Context, s.cancel = context.WithCancel(context.Background())
	s.git = services.NewGitService()
}

func (s *GitServiceTestSuite) TearDownTest() {
	s.cancel()
}

func (s *GitServiceTestSuite) TestCloneAndOpenFile() {
	s.Run("given without ssh key, when clone public repo and read README.md, then should be ok", func() {
		repo, err := s.git.Clone(
			context.Background(),
			"https://github.com/paust-team/pko-t5",
			"main",
		)
		s.Require().NoError(err)

		file, err := s.git.ReadFile(repo, "README.md")
		s.Require().NoError(err)

		s.NotNil(file)

		s.Contains(string(file), "pko-t5")
		s.T().Logf("README.md: %s", string(file))
	})
}

func (s *GitServiceTestSuite) TestClone_InvalidGitRepo() {
	_, err := s.git.Clone(
		context.Background(),
		"https://github.com/invalid-git-repo",
		"main",
	)
	s.Require().Error(err)
}

func (s *GitServiceTestSuite) TestCloneAndReadDir() {
	s.Run("given public repo, when clone and read root dir, then should be ok", func() {
		repo, err := s.git.Clone(
			context.Background(),
			"https://github.com/paust-team/pko-t5",
			"main",
		)
		s.Require().NoError(err)

		gitDir, err := s.git.OpenDir(repo, "pkot5")
		s.Require().NoError(err)

		var filenames []string
		s.Require().NoError(util.WalkGitTree(gitDir, func(
			basePath string,
			entry object.TreeEntry,
			fileContents []byte,
		) error {
			if fileContents != nil {
				filenames = append(filenames, filepath.Join(basePath, entry.Name))
			}
			return nil
		}))

		s.True(len(filenames) > 0)
		s.Contains(filenames, "args.py")
		s.Contains(filenames, "xl/__init__.py")
		s.Contains(filenames, "utils/__init__.py")
	})
}

func (s *GitServiceTestSuite) TestCloneAndReadDir_RootDir() {
	repo, err := s.git.Clone(
		s,
		"https://github.com/paust-team/pko-t5",
		"main",
	)

	s.Require().NoError(err)

	gitDir, err := s.git.OpenDir(repo, "")
	s.Require().NoError(err)

	readme, err := gitDir.File("README.md")
	readmeText, err := readme.Contents()
	s.Require().NoError(err)

	s.T().Logf("README.md: %s", readmeText)
	s.Contains(readmeText, "pko-t5")
}

func (s *GitServiceTestSuite) TestArchiveTarGitDir() {
	repo, err := s.git.Clone(
		context.Background(),
		"https://github.com/habiliai/vapi-user-management",
		"main",
	)
	s.Require().NoError(err)

	gitDir, err := s.git.OpenDir(repo, "")
	s.Require().NoError(err)
	s.NotNil(gitDir)

	var mem bytes.Buffer
	s.Require().NoError(util.ArchiveGitTree(gitDir, &mem))

	s.Greater(len(mem.Bytes()), 0)
	s.NoError(os.WriteFile("helloworld.tar", mem.Bytes(), 0644))
	defer os.Remove("helloworld.tar")

	tfs := tarfs.New(tar.NewReader(&mem))

	{
		info, err := tfs.Stat("_core")
		s.Require().NoError(err)

		s.True(info.IsDir())
	}
	{
		info, err := tfs.Stat("update-user")
		s.Require().NoError(err)

		s.True(info.IsDir())
	}
	{
		info, err := tfs.Stat("update-user/index.ts")
		s.Require().NoError(err)

		s.False(info.IsDir())
	}
	{
		info, err := tfs.Stat("get-user-info/index.ts")
		s.Require().NoError(err)

		s.False(info.IsDir())
	}
}

func (s *GitServiceTestSuite) TestCloneAndCommitFile() {
	githubAccessToken := os.Getenv("YOUR_GITHUB_TOKEN")
	if githubAccessToken == "" {
		s.T().Skip("github token is missing")
	}

	ctx := context.TODO()

	gitUrl := fmt.Sprintf("https://%s@github.com/jcooky/test-repo", githubAccessToken)

	gitService := services.NewGitService()

	gitRepo, err := gitService.Clone(ctx, gitUrl, "main",
		services.GitCloneOptionsAllowCommit(),
	)
	s.Require().NoError(err)

	wt, err := gitRepo.Worktree()
	s.Require().NoError(err)
	s.Require().NotNil(wt)

	s.Require().NoError(
		gitService.CommitFile(gitRepo, "test.txt", []byte("hello world"), "test commit"),
	)
}
