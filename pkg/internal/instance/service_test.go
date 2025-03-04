package instance_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/instance"
	"github.com/habiliai/apidepot/pkg/internal/k8s"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/habiliai/apidepot/pkg/internal/user"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type InstanceServiceTestSuite struct {
	suite.Suite
	context.Context

	instances     instance.Service
	db            *gorm.DB
	k8sClientPool *k8s.ClientPool
	stacks        stack.Service
	users         *usertest.Service
	vapis         vapi.Service

	project domain.Project
	stack   *domain.Stack
	user    domain.User
}

func (s *InstanceServiceTestSuite) SetupTest() {
	ctx := context.TODO()
	container := digo.NewContainer(
		ctx,
		digo.EnvTest,
		nil,
	)

	s.users = usertest.NewService()
	s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Maybe()
	digo.Set(container, user.ServiceKey, s.users)
	s.db = digo.MustGet[*gorm.DB](container, services.ServiceKeyDB)

	s.vapis = digo.MustGet[vapi.Service](container, vapi.ServiceKey)
	s.k8sClientPool = digo.MustGet[*k8s.ClientPool](container, k8s.ServiceKeyK8sClientPool)
	s.instances = digo.MustGet[instance.Service](container, instance.ServiceKey)
	s.stacks = digo.MustGet[stack.Service](container, stack.ServiceKey)

	ctx = helpers.WithTx(ctx, s.db)
	s.Context = ctx

	{
		s.T().Log("-- add user")
		s.user = domain.User{
			Name:                 "1dennispark",
			GithubInstallationId: 1,
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
		s.T().Log("-- add stack by project")
		name, err := uuid.NewRandom()
		s.Require().NoError(err)
		s.stack, err = s.stacks.CreateStack(s.Context, stack.CreateStackInput{
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

func (s *InstanceServiceTestSuite) TearDownTest() {
	defer s.users.AssertExpectations(s.T())

	if s.stack != nil {
		var instances []domain.Instance
		s.Require().NoError(s.db.Find(&instances, "stack_id = ?", s.stack.ID).Error)
		s.T().Logf("remove instances number of %d\n", len(instances))
		for _, inst := range instances {
			s.instances.DeleteInstance(s, inst.ID, true)
			s.db.Unscoped().Delete(&inst)
		}

		s.Require().NoError(s.db.Find(&instances, "stack_id = ?", s.stack.ID).Error)
		s.T().Logf("instances number of %d\n", len(instances))

		s.Require().NoError(s.stacks.DeleteStack(s, s.stack.ID))
	}
}

func TestInstanceService(t *testing.T) {
	suite.Run(t, new(InstanceServiceTestSuite))
}
