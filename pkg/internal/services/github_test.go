package services_test

import (
	"context"
	"fmt"
	"github.com/google/go-github/v60/github"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"os"
	"testing"
)

type GitHubClientTestSuite struct {
	suite.Suite
	context.Context

	gh  services.GithubClient
	git services.GitService
}

func TestGithubClient(t *testing.T) {
	suite.Run(t, new(GitHubClientTestSuite))
}

func (s *GitHubClientTestSuite) SetupTest() {
	ctx := context.TODO()
	container := digo.NewContainer(ctx, digo.EnvTest, nil)
	s.gh = digo.MustGet[services.GithubClient](container, services.ServiceKeyGithubClient)
	s.git = digo.MustGet[services.GitService](container, services.ServiceKeyGitService)
	s.Context = ctx
}

func (s *GitHubClientTestSuite) TestGithubClient_GetUser() {
	githubAccessToken := os.Getenv("YOUR_GITHUB_TOKEN")
	if githubAccessToken == "" {
		s.T().Skip("github token is missing")
	}
	ctx := context.TODO()

	user, err := s.gh.GetUser(ctx, githubAccessToken)

	s.Require().NoError(err)
	s.Require().NotNil(user)
}

func (s *GitHubClientTestSuite) TestHardfork() {
	githubAccessToken := os.Getenv("YOUR_GITHUB_TOKEN")
	if githubAccessToken == "" {
		s.T().Skip("github token is missing")
	}

	user, err := s.gh.GetUser(s, githubAccessToken)
	s.Require().NoError(err)

	githubClient := github.NewClient(nil).WithAuthToken(githubAccessToken)

	githubClient.Repositories.Delete(s, *user.Login, "hardfork_test")

	err = s.gh.CreateRepository(s.Context, githubAccessToken, *user.Login, "hardfork_test", "")
	s.Require().NoError(err)
	defer githubClient.Repositories.Delete(s, *user.Login, "hardfork_test")

	srcGitUrl := "https://github.com/habiliai/service-template-example"
	dstGitUrl := fmt.Sprintf("https://%s@github.com/jcooky/hardfork_test", githubAccessToken)

	s.Require().NoError(s.git.CopyRepo(s, srcGitUrl, dstGitUrl))

	resp, err := http.Get("https://raw.githubusercontent.com/jcooky/hardfork_test/refs/heads/main/README.md")
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	contents, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	s.Greater(len(contents), 0)
	s.Contains(string(contents), "service-template-example")
	s.T().Logf("contents: %s", contents)
}
