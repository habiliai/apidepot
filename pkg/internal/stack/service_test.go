package stack_test

import (
	"context"
	"github.com/google/uuid"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/habiliai/apidepot/pkg/internal/user"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/martinlindhe/base36"
	"github.com/mokiat/gog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"strings"
	"testing"
)

type StackServiceTestSuite struct {
	suite.Suite
	context.Context

	stack   *domain.Stack
	project domain.Project

	db             *gorm.DB
	stackService   stack.Service
	vapis          vapi.Service
	users          *usertest.Service
	user           domain.User
	storageClient  *storage.Client
	k8syamlService *k8syaml.Service
}

func TestStackService(t *testing.T) {
	suite.Run(t, new(StackServiceTestSuite))
}

func (s *StackServiceTestSuite) SetupTest() {
	ctx := context.TODO()
	{
		container := digo.NewContainer(ctx, digo.EnvTest, nil)

		s.users = usertest.NewService()
		digo.Set(container, user.ServiceKey, s.users)
		s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Maybe()

		s.db = digo.MustGet[*gorm.DB](container, services.ServiceKeyDB)
		ctx = helpers.WithTx(ctx, s.db)
		s.stackService = digo.MustGet[stack.Service](container, stack.ServiceKey)
		s.vapis = digo.MustGet[vapi.Service](container, vapi.ServiceKey)
		s.storageClient = digo.MustGet[*storage.Client](container, services.ServiceKeyStorageClient)
		s.k8syamlService = digo.MustGet[*k8syaml.Service](container, k8syaml.ServiceKey)

		s.Context = ctx
	}

	{
		s.T().Log("-- add user")
		s.user = domain.User{
			Name:                 "1dennispark",
			GithubInstallationId: 1,
			Role:                 domain.UserRoleUser,
		}
		s.Require().NoError(s.user.Save(s.db))
	}
	{
		s.T().Log("-- add project")
		name, err := uuid.NewRandom()
		s.Require().NoError(err)
		s.project = domain.Project{
			Name:    name.String(),
			Owner:   s.user,
			OwnerID: s.user.ID,
		}
		s.Require().NoError(s.project.Save(s.db))
	}

	{
		s.T().Log("-- add stack by project 'test123'")
		name, err := uuid.NewRandom()
		s.Require().NoError(err)
		s.stack, err = s.stackService.CreateStack(s, stack.CreateStackInput{
			ProjectID:     s.project.ID,
			Name:          name.String(),
			SiteURL:       "http://localhost:8080",
			DefaultRegion: tcltypes.InstanceZoneDefault,
		})
		s.Require().NoError(err)
		s.Require().NotNil(s.stack)
		s.Require().NotEmpty(s.stack.Hash)
		s.Require().NotEmpty(s.stack.Domain)
	}
}

func (s *StackServiceTestSuite) TearDownTest() {
	defer s.users.AssertExpectations(s.T())
	if s.stack != nil {
		s.T().Log("delete stack if it is not nil")
		err := s.stackService.DeleteStack(s, s.stack.ID)
		s.Require().NoError(err)
	}
}

func TestGetRandomNames(t *testing.T) {
	times := 1_000_000
	for i := 0; i < times; i++ {
		uid, err := uuid.NewRandom()
		assert.NoError(t, err)
		name := strings.ToLower(base36.EncodeBytes(uid[:]))
		if len(name) != 25 {
			var uid2 = uuid.UUID(base36.DecodeToBytes(strings.ToUpper(name)))
			require.Equal(t, uid.String(), uid2.String())
		}
	}
}

func (s *StackServiceTestSuite) TestGetStacks() {
	ctx := s.Context

	project := domain.Project{
		Name:    "test-get-stacks",
		Owner:   s.user,
		OwnerID: s.user.ID,
	}
	s.Require().NoError(project.Save(s.db))

	// add stack
	name, err := uuid.NewRandom()
	s.Require().NoError(err)
	_, err = s.stackService.CreateStack(ctx, stack.CreateStackInput{
		ProjectID:     project.ID,
		Name:          name.String(),
		SiteURL:       "http://localhost:8080",
		DefaultRegion: tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)

	// search stacks
	{
		res, err := s.stackService.GetStacks(ctx,
			project.ID,
			gog.PtrOf(name.String()),
			0,
			0,
		)
		s.Require().NoError(err)
		s.Require().Len(res, 1)
	}
}

func (s *StackServiceTestSuite) TestAddStack1() {
	ctx := s.Context

	s.Run("when add stack named 'test123', should be ok", func() {
		// when
		stack, err := s.stackService.CreateStack(ctx, stack.CreateStackInput{
			ProjectID: s.project.ID,
			Name:      "test123",
			SiteURL:   "http://localhost:8080",
		})
		defer s.stackService.DeleteStack(ctx, stack.ID)

		// then
		s.Require().NoError(err)
		s.Require().NotNil(stack)
		s.Require().NotEmpty(stack.Hash)
		s.Require().NotEmpty(stack.Domain)
	})

}

func (s *StackServiceTestSuite) TestAddStack2() {
	ctx := s.Context

	s.Run("when add stack named 51 length characters, should be error bad request", func() {
		//when
		stack, err := s.stackService.CreateStack(ctx, stack.CreateStackInput{
			ProjectID: s.project.ID,
			Name:      "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890",
			SiteURL:   "http://localhost:8080",
		})
		defer func() {
			if stack == nil {
				return
			}
			s.stackService.DeleteStack(ctx, stack.ID)
		}()

		// then
		s.Require().Error(err)
		s.Require().ErrorAs(err, &tclerrors.ErrBadRequest)
	})

}

func (s *StackServiceTestSuite) TestAddStack3() {
	ctx := s.Context

	s.Run("when add stack named non alphanumeric, should be error bad request", func() {
		// when
		stack, err := s.stackService.CreateStack(ctx, stack.CreateStackInput{
			ProjectID: s.project.ID,
			Name:      "한글 이름 입니다.",
			SiteURL:   "http://localhost:8080",
		})
		defer func() {
			if stack == nil {
				return
			}
			s.stackService.DeleteStack(ctx, stack.ID)
		}()

		// then
		s.Require().Error(err)
		s.Require().ErrorAs(err, &tclerrors.ErrBadRequest)
	})
}

func (s *StackServiceTestSuite) TestPatchStack() {
	s.Run("given enable stack, when patch stack, should be ok", func() {
		// given
		st, err := s.stackService.CreateStack(s.Context, stack.CreateStackInput{
			ProjectID:     s.project.ID,
			Name:          "test-patch-stack",
			SiteURL:       "http://localhost:8080",
			DefaultRegion: tcltypes.InstanceZoneDefault,
		})
		s.Require().NoError(err)
		defer s.stackService.DeleteStack(s.Context, st.ID)

		// when
		err = s.stackService.PatchStack(s.Context, st.ID, stack.PatchStackInput{
			SiteURL: gog.PtrOf("http://localhost:8081"),
		})

		// then
		s.Require().NoError(err)
	})
}
