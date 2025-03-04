package vapi_test

import (
	"archive/tar"
	"bytes"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
	"github.com/stretchr/testify/mock"
	"os"
)

func (s *VapiTestSuite) TestRegister() {
	s.users.On("GetUser", mock.Anything).Return(&domain.User{
		Model: domain.Model{
			ID: 1,
		},
		GithubInstallationId: 1,
	}, nil).Once()
	defer s.users.AssertExpectations(s.T())

	rel, err := s.vapiService.Register(
		s.Context(),
		"habiliai/vapi-user-management",
		"main",
		"user-management",
		"User management",
		[]string{"Social"},
		"",
		"https://habili.ai",
	)
	s.Require().NoError(err)
	s.Require().NotNil(rel)

	s.Require().NotZero(rel.ID)
	s.Require().Equal("0.1.0", rel.Version)

	{
		content, err := s.storage.DownloadFile(s.Context(), constants.VapiBucketId, "user-management/v0.1.0.tar")
		s.Require().NoError(err)
		s.Require().NotEmpty(content)
	}

	vapiRel, err := domain.FindVapiReleaseByPackageNameAndVersion(
		s.db,
		"user-management",
		"0.1.0",
	)
	s.Require().NoError(err)

	{
		content, err := s.storage.DownloadFile(s.Context(), constants.VapiBucketId, vapiRel.TarFilePath)
		s.Require().NoError(err)
		s.Require().NotEmpty(content)

		tfs := tarfs.New(tar.NewReader(bytes.NewBuffer(content)))
		s.Require().NoError(afero.Walk(tfs, "./", func(path string, info os.FileInfo, err error) error {
			s.T().Logf("path: %s", path)
			return nil
		}))
		{
			fileInfo, err := tfs.Stat("delete-user")
			s.Require().NoError(err)

			s.True(fileInfo.IsDir())
		}
		{
			fileInfo, err := tfs.Stat("get-user-info")
			s.Require().NoError(err)

			s.True(fileInfo.IsDir())
		}
		{
			fileInfo, err := tfs.Stat("update-user")
			s.Require().NoError(err)

			s.True(fileInfo.IsDir())
		}
	}

}
