package vapi_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type VapiTestSuite struct {
	suite.Suite

	db          *gorm.DB
	vapiService vapi.Service
	users       *usertest.Service
	storage     *storage.Client

	user domain.User

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *VapiTestSuite) Context() context.Context {
	return s.ctx
}

func (s *VapiTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	var err error

	container := digo.NewContainer(
		s.Context(),
		digo.EnvTest,
		nil,
	)

	s.db, err = digo.Get[*gorm.DB](container, services.ServiceKeyDB)
	s.Require().NoError(err)
	s.ctx = helpers.WithTx(s.Context(), s.db)
	s.users = usertest.NewService()
	s.storage, err = digo.Get[*storage.Client](container, services.ServiceKeyStorageClient)

	s.Require().NoError(err)
	s.vapiService, err = vapi.NewService(
		s.storage,
		s.users,
		digo.MustGet[services.GitService](container, services.ServiceKeyGitService),
		digo.MustGet[services.GithubClient](container, services.ServiceKeyGithubClient),
	)
	s.Require().NoError(err)

	s.Require().NoError(err)
	s.user = domain.User{GithubInstallationId: 1}
	s.Require().NoError(s.user.Save(s.db))
}

func (s *VapiTestSuite) TearDownTest() {
	defer s.cancel()
}

func TestVapiService(t *testing.T) {
	suite.Run(t, new(VapiTestSuite))
}

func (s *VapiTestSuite) deployTestVapis() {
	s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Twice()
	defer s.users.AssertExpectations(s.T())

	{
		output, err := s.vapiService.Register(s.Context(),
			"habiliai/vapi-helloworld-sns",
			"main",
			"sns",
			`# Test Vapi - Hello World SNS
This is a test vapi for the Hello World SNS
`,
			[]string{"Social"},
			"",
			"https://habili.ai",
		)
		s.Require().NoError(err)
		s.Require().NotNil(output)
	}
	{
		output, err := s.vapiService.Register(s.Context(),
			"habiliai/vapi-helloworld",
			"main",
			"helloworld",
			`# Test Vapi - Hello World
This is a test vapi for the Hello World
`,
			[]string{"Social"},
			"",
			"https://habili.ai",
		)
		s.Require().NoError(err)
		s.Require().NotNil(output)
	}
}
