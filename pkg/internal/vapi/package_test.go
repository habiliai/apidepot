package vapi_test

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/stretchr/testify/mock"
)

func (s *VapiTestSuite) TestGivenDeployedTestVapisAndNotEmptyReleasesWhenDeletePackageThenError() {
	// given
	s.deployTestVapis()
	s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Once()
	defer s.users.AssertExpectations(s.T())

	// when
	err := s.vapiService.DeletePackage(s.Context(), 2)

	// then
	s.Require().Error(err)
	s.Require().ErrorIs(err, tclerrors.ErrPreconditionRequired)
}

func (s *VapiTestSuite) TestGivenDeployedTestVapisWhenGetPackageThenOK() {
	// given
	s.deployTestVapis()

	// when
	pkg, err := s.vapiService.GetPackage(s.Context(), 1)

	// then
	s.Require().NoError(err)
	s.Require().NotNil(pkg)
	s.Equal("sns", pkg.Name)
	s.Equal("1.2.0", pkg.Releases[len(pkg.Releases)-1].Version)
}

func (s *VapiTestSuite) TestGivenDeployedTestVapisWhenDeleteAllPackagesThenDeployedPackagesDeleted() {
	// given
	s.deployTestVapis()
	s.users.On("GetUser", mock.Anything).Return(&s.user, nil).Once()
	defer s.users.AssertExpectations(s.T())

	// when
	err := s.vapiService.DeleteAllPackages(s.Context(), 0)

	// then
	s.Require().NoError(err)
	pkgs, err := domain.FindVapiPackages(s.db)
	s.Require().NoError(err)
	s.Require().Len(pkgs, 0)
}

func (s *VapiTestSuite) TestGivenDeployedTestVapisWhenGetPackagesThenOK() {
	// given
	var (
		name = "helloworld"
	)
	s.deployTestVapis()

	// when
	pkgs, err := s.vapiService.GetPackages(s.Context(), vapi.GetPackagesInput{
		Name: &name,
	})

	// then
	s.Require().NoError(err)
	s.Require().Len(pkgs, 1)
	s.Equal(name, pkgs[0].Name)
}
