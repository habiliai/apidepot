package vapi_test

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/stretchr/testify/mock"
)

func (s *VapiTestSuite) TestGetMigrationsForDB() {
	// given
	s.deployTestVapis()

	// when
	vapiRel, err := domain.FindVapiReleaseByPackageNameAndVersion(
		s.db,
		"helloworld",
		"0.1.0",
	)
	s.Require().NoError(err)
	migrations, err := s.vapiService.GetDBMigrations(s.Context(), *vapiRel)
	s.Require().NoError(err)

	// then
	s.Require().Len(migrations, 1)
	s.Equal("240326135200", migrations[0].Version.Format("060102150405"))
	s.T().Logf("query: %s", migrations[0].Query)
}

func (s *VapiTestSuite) TestGetRelease() {
	// given
	s.deployTestVapis()

	// when
	vapiRel, err := s.vapiService.GetRelease(s.Context(), 1)

	// then
	s.Require().NoError(err)
	s.Require().NotNil(vapiRel)
	s.Equal("sns", vapiRel.Package.Name)
	s.Equal("1.2.0", vapiRel.Version)
}

func (s *VapiTestSuite) TestDeleteRelease() {
	// given
	s.deployTestVapis()
	s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Once()
	defer s.users.AssertExpectations(s.T())
	vapiRel, err := s.vapiService.GetRelease(s.Context(), 2)

	// when
	err = s.vapiService.DeleteRelease(s.Context(), 2)

	// then
	s.Require().NoError(err)
	_, err = s.storage.DownloadFile(s.Context(), constants.VapiBucketId, vapiRel.TarFilePath)
	s.Require().Error(err)

	vapiRel, err = s.vapiService.GetRelease(s.Context(), 2)
	s.Require().Error(err)
	s.Require().Nil(vapiRel)
	s.Require().ErrorIs(err, tclerrors.ErrNotFound)
}

func (s *VapiTestSuite) TestGivenDeployTestVapisWhenDeleteReleasesByPackageIdThenThePackageHasNotReleases() {
	// given
	s.deployTestVapis()
	s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Once()
	defer s.users.AssertExpectations(s.T())

	// when
	err := s.vapiService.DeleteReleasesByPackageId(s.Context(), 2)

	// then
	s.Require().NoError(err)
	rels, err := domain.FindVapiReleasesByPackageID(s.db, 2)
	s.Require().NoError(err)
	s.Require().Len(rels, 0)
}

func (s *VapiTestSuite) TestGivenDeployTestVapisWhenDeleteAllReleasesThenAllReleasesAreDeleted() {
	// given
	s.deployTestVapis()
	s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Once()
	defer s.users.AssertExpectations(s.T())

	// when
	err := s.vapiService.DeleteAllReleases(s.Context())

	// then
	s.Require().NoError(err)
	rels, err := domain.FindVapiReleases(s.db)
	s.Require().NoError(err)
	s.Require().Len(rels, 0)
}
