package domain_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type DomainTestSuite struct {
	suite.Suite

	db *gorm.DB
}

func (s *DomainTestSuite) SetupTest() {
	initCtx := digo.NewContainer(
		context.Background(),
		digo.EnvTest,
		nil,
	)

	var err error
	s.db, err = digo.Get[*gorm.DB](initCtx, services.ServiceKeyDB)
	s.Require().NoError(err)
}

func TestDomain(t *testing.T) {
	suite.Run(t, new(DomainTestSuite))
}
