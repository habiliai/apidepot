package cliapp_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/cliapp"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/user"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/supabase-community/gotrue-go"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
	"gorm.io/gorm"
	"testing"
)

type ServiceTestSuite struct {
	suite.Suite
	context.Context

	users   *usertest.Service
	cliapps cliapp.Service
	gotrue  gotrue.Client
	db      *gorm.DB
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	s.Context = context.TODO()

	container := digo.NewContainer(s.Context, digo.EnvTest, nil)
	s.users = usertest.NewService()
	digo.Set(container, user.ServiceKey, s.users)
	s.cliapps = digo.MustGet[cliapp.Service](container, cliapp.ServiceKey)
	s.gotrue = digo.MustGet[gotrue.Client](container, services.ServiceKeyGoTrueClient)
	s.db = digo.MustGet[*gorm.DB](container, services.ServiceKeyDB)

	s.Context = helpers.WithTx(s.Context, s.db)
}

func (s *ServiceTestSuite) TearDownTest() {
	listUsersResp, err := s.gotrue.AdminListUsers()
	s.Require().NoError(err)
	for _, user := range listUsersResp.Users {
		s.Require().NoError(s.gotrue.AdminDeleteUser(gotruetypes.AdminDeleteUserRequest{
			UserID: user.ID,
		}))
	}
}

func (s *ServiceTestSuite) TestService_GetCliAppByAppId() {
	// Test for 1 user access via cli app by adding a cli app and getting it by app id

	host := "test.com"

	session, err := s.gotrue.Signup(gotruetypes.SignupRequest{
		Email:    "test@test.com",
		Password: "test123",
	})
	s.Require().NoError(err)

	user := domain.User{
		AuthUserId: session.ID.String(),
	}
	s.Require().NoError(user.Save(s.db))
	s.users.On("GetUser", mock.Anything).Return(&user, nil).Twice()
	defer s.users.AssertExpectations(s.T())

	cliAppResp, err := s.cliapps.RegisterCliApp(s.Context, host, session.RefreshToken)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(s.cliapps.DeleteCliApp(s, cliAppResp.AppId))
		_, err = s.cliapps.VerifyCliApp(s.Context, cliAppResp.AppId, cliAppResp.AppSecret)
		s.Require().Error(err)
	}()

	getCliAppResp, err := s.cliapps.VerifyCliApp(s.Context, cliAppResp.AppId, cliAppResp.AppSecret)
	s.Require().NoError(err)

	authUser, err := s.gotrue.WithToken(getCliAppResp.AccessToken).GetUser()
	s.Require().NoError(err)
	s.Require().Equal(user.AuthUserId, authUser.ID.String())
}
