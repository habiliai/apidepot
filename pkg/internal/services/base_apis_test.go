package services_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/stretchr/testify/suite"
	"testing"
)

type StorageClientTestSuite struct {
	suite.Suite
	context.Context

	storage *storage.Client
}

func TestStorageClient(t *testing.T) {
	suite.Run(t, new(StorageClientTestSuite))
}

func (s *StorageClientTestSuite) SetupTest() {
	var err error
	s.Context = context.TODO()
	container := digo.NewContainer(
		s,
		digo.EnvTest,
		nil,
	)

	s.storage, err = digo.Get[*storage.Client](container, services.ServiceKeyStorageClient)
	s.Require().NoError(err)
}

func (s *StorageClientTestSuite) TestStorageClient_NotFoundBucket() {
	_, err := s.storage.GetBucket(s, "not-found-bucket")
	s.Require().Error(err)
	storageErr := &storage.StorageError{}
	s.Require().ErrorAs(err, &storageErr)
	s.Equalf(404, storageErr.Status, "storageErr: %v", storageErr)
}
