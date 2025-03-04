package svctpl_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/svctpl"
	"github.com/habiliai/apidepot/pkg/internal/user"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type ServiceTestSuite struct {
	suite.Suite
	context.Context

	users   *usertest.Service
	svctpls svctpl.Service
	vapis   vapi.Service
	db      *gorm.DB
}

func (s *ServiceTestSuite) SetupTest() {
	ctx := context.TODO()

	container := digo.NewContainer(ctx, digo.EnvTest, nil)
	s.users = usertest.NewService()
	digo.Set(container, user.ServiceKey, s.users)
	s.db = digo.MustGet[*gorm.DB](container, services.ServiceKeyDB)
	ctx = helpers.WithTx(ctx, s.db)
	s.svctpls = digo.MustGet[svctpl.Service](container, svctpl.ServiceKey)
	s.vapis = digo.MustGet[vapi.Service](container, vapi.ServiceKey)

	s.Context = ctx
}

func (s *ServiceTestSuite) TearDownTest() {
	defer s.users.AssertExpectations(s.T())
}

func TestServiceTemplateService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
