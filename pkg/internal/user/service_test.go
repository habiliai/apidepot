package user_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/services"
	servicestest "github.com/habiliai/apidepot/pkg/internal/services/test"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/stretchr/testify/suite"
	"github.com/supabase-community/gotrue-go"
	. "github.com/supabase-community/gotrue-go/types"
	"gorm.io/gorm"
	"testing"
)

type ServiceTestSuite struct {
	suite.Suite

	gotrue  gotrue.Client
	github  *servicestest.MockGithubClient
	service user.Service
	storage *storage.Client
	db      *gorm.DB

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *ServiceTestSuite) Context() context.Context {
	return s.ctx
}

func (s *ServiceTestSuite) SetupTest() {
	var err error
	s.ctx, s.cancel = context.WithCancel(context.Background())
	container := digo.NewContainer(s.Context(), digo.EnvTest, nil)

	s.db = digo.MustGet[*gorm.DB](container, services.ServiceKeyDB)
	s.ctx = helpers.WithTx(s.Context(), s.db)
	s.gotrue = digo.MustGet[gotrue.Client](container, services.ServiceKeyGoTrueClient)
	s.github = servicestest.NewTestGithubClient()
	digo.Set(container, services.ServiceKeyGithubClient, s.github)
	s.storage = digo.MustGet[*storage.Client](container, services.ServiceKeyStorageClient)
	s.service, err = digo.Get[user.Service](container, user.ServiceKey)
	s.Require().NoError(err)
}

func (s *ServiceTestSuite) TearDownTest() {
	defer s.cancel()
	{
		usersResp, err := s.gotrue.AdminListUsers()
		s.Require().NoError(err)
		for _, u := range usersResp.Users {
			s.NoError(s.gotrue.AdminDeleteUser(AdminDeleteUserRequest{
				UserID: u.ID,
			}))
		}
	}
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
