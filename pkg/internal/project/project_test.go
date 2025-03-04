package project_test

import (
	"context"
	"github.com/google/uuid"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/project"
	"github.com/habiliai/apidepot/pkg/internal/services"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type ProjectServiceTestSuite struct {
	suite.Suite
	context.Context

	db       *gorm.DB
	projects project.Service
	users    *usertest.Service
}

func TestProjectService(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}

func (s *ProjectServiceTestSuite) SetupTest() {
	ctx := context.TODO()
	container := digo.NewContainer(
		ctx,
		digo.EnvTest,
		nil,
	)

	s.db = digo.MustGet[*gorm.DB](container, services.ServiceKeyDB)
	s.users = usertest.NewService()
	s.projects = project.NewService(s.users)

	ctx = helpers.WithTx(ctx, s.db)
	s.Context = ctx
}

func (s *ProjectServiceTestSuite) TearDownTest() {
}

func (s *ProjectServiceTestSuite) TestGetProjects() {
	user := domain.User{
		Name:       "1dennispark",
		AuthUserId: uuid.NewString(),
		Role:       domain.UserRoleAdmin,
	}
	s.Require().NoError(user.Save(s.db))

	uid, err := uuid.NewRandom()
	s.NoError(err)
	var p = domain.Project{
		Name:    uid.String(),
		OwnerID: user.ID,
		Owner:   user,
	}
	s.Require().NoError(p.Save(s.db))

	s.users.On("GetUser", mock.Anything).Return(&user, nil).Times(1)
	projects, err := s.projects.GetProjects(s, project.GetProjectsInput{
		Name: &p.Name,
	})
	s.NoError(err)
	s.Len(projects, 1)
}

func (s *ProjectServiceTestSuite) TestCreateProject() {
	user := domain.User{
		Name:       "1dennispark",
		AuthUserId: uuid.NewString(),
	}
	s.Require().NoError(user.Save(s.db))

	s.users.On("GetUser", mock.Anything).Return(&user, nil).Times(4)
	defer s.users.AssertExpectations(s.T())

	s.Run("when create a project, should be success", func() {
		project, err := s.projects.CreateProject(s,
			"test",
			"",
		)
		s.Require().NoError(err)

		s.Require().NotNil(project)
	})

	s.Run("when create a project with empty name, should be return error", func() {
		project, err := s.projects.CreateProject(s,
			"",
			"",
		)
		s.Require().Error(err)
		s.Require().Nil(project)
	})

	s.Run("when create a project named 51 characters, should be error bad request", func() {
		project, err := s.projects.CreateProject(s,
			"abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890",
			"",
		)
		s.Require().Error(err)
		s.Require().ErrorAs(err, &tclerrors.ErrBadRequest)
		s.Require().Nil(project)
	})

	s.Run("when create a project named non alphanumeric, should be error bad request", func() {
		project, err := s.projects.CreateProject(s,
			"한글 이름 입니다.",
			"",
		)
		s.Require().NoError(err)
		s.Require().NotNil(project)
	})

}
