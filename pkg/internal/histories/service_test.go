package histories_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/histories"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/habiliai/apidepot/pkg/internal/user"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type HistoriesTestSuite struct {
	suite.Suite
	context.Context

	service       histories.Service
	db            *gorm.DB
	userService   *usertest.Service
	stackService  stack.Service
	bucketService services.BucketService
}

func (s *HistoriesTestSuite) SetupTest() {
	s.Context = context.TODO()

	container := digo.NewContainer(s, digo.EnvTest, nil)
	s.db = digo.MustGet[*gorm.DB](container, services.ServiceKeyDB)
	s.Context = helpers.WithTx(s.Context, s.db)
	s.userService = usertest.NewService()
	digo.Set(container, user.ServiceKey, s.userService)
	s.service = digo.MustGet[histories.Service](container, histories.ServiceKey)
	s.stackService = digo.MustGet[stack.Service](container, stack.ServiceKey)
	s.bucketService = digo.MustGet[services.BucketService](container, services.ServiceKeyBucketService)
}

func TestInstanceHistoryService(t *testing.T) {
	suite.Run(t, new(HistoriesTestSuite))
}
