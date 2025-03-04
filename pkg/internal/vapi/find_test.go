package vapi_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/mokiat/gog"
	"slices"
)

func (s *VapiTestSuite) TestSearchVapis() {
	// Given
	user := domain.User{
		AuthUserId: "test-123123",
	}
	s.Require().NoError(user.Save(s.db))

	pkg1 := domain.VapiPackage{
		Name:    "sns-api",
		OwnerId: user.ID,
	}
	s.Require().NoError(pkg1.Save(s.db))

	rel1 := domain.VapiRelease{
		Version:   "1.0.0",
		PackageID: pkg1.ID,
		Package:   pkg1,
		Published: false,
	}
	s.Require().NoError(rel1.Save(s.db))

	pkg2 := domain.VapiPackage{
		Name:    "profile-sns-api",
		OwnerId: user.ID,
	}
	s.Require().NoError(pkg2.Save(s.db))

	rel2 := domain.VapiRelease{
		Version:   "0.2.0",
		PackageID: pkg2.ID,
		Package:   pkg2,
		Published: false,
	}
	s.Require().NoError(rel2.Save(s.db))

	// When
	output, err := s.vapiService.SearchVapis(s.Context(), vapi.SearchVapisInput{
		Name:    gog.PtrOf("sns"),
		PageNum: 1,
	})
	s.Require().NoError(err)

	// Then
	s.Len(output.Releases, 2)
	slices.SortStableFunc(output.Releases, func(lhs, rhs domain.VapiRelease) int {
		if lhs.Package.ID < rhs.Package.ID {
			return -1
		} else if lhs.Package.ID > rhs.Package.ID {
			return 1
		} else {
			return 0
		}
	})
	s.Equal(pkg1.ID, output.Releases[0].ID)
	s.Equal(uint(1), output.Releases[0].Package.ID)
}

func (s *VapiTestSuite) TestFindAllDependenciesForStack() {
	pkg1 := domain.VapiPackage{
		Name:  "test-vapi1",
		Owner: s.user,
	}
	s.Require().NoError(pkg1.Save(s.db))

	rel1 := domain.VapiRelease{
		Version:   "1.0.0",
		PackageID: pkg1.ID,
		Package:   pkg1,
		Published: false,
	}
	s.Require().NoError(rel1.Save(s.db))

	pkg2 := domain.VapiPackage{
		Name:  "test-vapi2",
		Owner: s.user,
	}
	s.Require().NoError(pkg2.Save(s.db))

	rel2 := domain.VapiRelease{
		Version:      "1.0.0",
		PackageID:    pkg2.ID,
		Package:      pkg2,
		Published:    false,
		Dependencies: []domain.VapiRelease{rel1},
	}
	s.Require().NoError(rel2.Save(s.db))

	pkg3 := domain.VapiPackage{
		Name:  "test-vapi3",
		Owner: s.user,
	}
	s.Require().NoError(pkg3.Save(s.db))

	rel3 := domain.VapiRelease{
		Version:      "1.0.0",
		PackageID:    pkg3.ID,
		Package:      pkg3,
		Published:    false,
		Dependencies: []domain.VapiRelease{rel2},
	}
	s.Require().NoError(rel3.Save(s.db))

	vapiReleases, err := s.vapiService.GetAllDependenciesOfVapiReleases(
		s.Context(),
		[]domain.VapiRelease{rel3},
	)
	s.Require().NoError(err)

	s.Equal(3, len(vapiReleases))
}
