package organization_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/organization"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/user"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type OrganizationTestSuite struct {
	suite.Suite

	context context.Context
	cancel  context.CancelFunc

	orgService      organization.Service
	userServiceMock *usertest.Service
}

func TestOrganization(t *testing.T) {
	suite.Run(t, new(OrganizationTestSuite))
}

func (s *OrganizationTestSuite) SetupTest() {
	tclog.Init("debug", "")

	s.context, s.cancel = context.WithCancel(context.Background())
	container := digo.NewContainer(
		s.context,
		digo.EnvTest,
		nil,
	)

	db, err := digo.Get[*gorm.DB](container, services.ServiceKeyDB)
	s.Require().NoError(err)
	s.context = helpers.WithTx(s.context, db)

	s.userServiceMock = usertest.NewService()
	digo.Set(container, user.ServiceKey, s.userServiceMock)

	s.orgService, err = digo.Get[organization.Service](container, organization.ServiceKey)
	s.Require().NoError(err)
}

func (s *OrganizationTestSuite) TearDownTest() {
	defer s.cancel()
}
