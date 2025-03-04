package domain_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
)

func (s *DomainTestSuite) TestVapiRelease_DFS() {
	// Given
	user := domain.User{
		AuthUserId: "test-123123",
	}
	s.Require().NoError(user.Save(s.db))

	vapiRelease := domain.VapiRelease{
		Version: "1.0.0",
		Package: domain.VapiPackage{
			Name:    "vapi-1",
			OwnerId: user.ID,
		},
	}
	s.Require().NoError(vapiRelease.Save(s.db))

	vapiRelease2 := domain.VapiRelease{
		Version: "2.0.0",
		Package: domain.VapiPackage{
			Name:    "vapi-2",
			OwnerId: user.ID,
		},
	}
	s.Require().NoError(vapiRelease2.Save(s.db))

	vapiRelease3 := domain.VapiRelease{
		Version: "3.0.0",
		Package: domain.VapiPackage{
			Name:    "vapi-3",
			OwnerId: user.ID,
		},
	}
	s.Require().NoError(vapiRelease3.Save(s.db))

	vapiRelease4 := domain.VapiRelease{
		Version: "4.0.0",
		Package: domain.VapiPackage{
			Name:    "vapi-4",
			OwnerId: user.ID,
		},
	}
	s.Require().NoError(vapiRelease4.Save(s.db))

	vapiRelease.Dependencies = append(vapiRelease.Dependencies, vapiRelease2, vapiRelease3)
	vapiRelease2.Dependencies = append(vapiRelease2.Dependencies, vapiRelease4)
	vapiRelease3.Dependencies = append(vapiRelease3.Dependencies, vapiRelease4)
	s.Require().NoError(vapiRelease.Save(s.db))
	s.Require().NoError(vapiRelease2.Save(s.db))
	s.Require().NoError(vapiRelease3.Save(s.db))

	deps := map[uint][]uint{}

	// When
	err := vapiRelease.DFS(s.db, func(v domain.VapiRelease, parent *domain.VapiRelease) error {
		var parentID uint = 0
		if parent != nil {
			parentID = parent.ID
		}
		deps[parentID] = append(deps[parentID], v.ID)
		return nil
	})

	// Then
	s.Require().NoError(err)
	s.Require().Len(deps, 4)
	s.Contains(deps[0], vapiRelease.ID)
	s.Contains(deps[vapiRelease.ID], vapiRelease2.ID)
	s.Contains(deps[vapiRelease.ID], vapiRelease3.ID)
	s.Contains(deps[vapiRelease2.ID], vapiRelease4.ID)
	s.Contains(deps[vapiRelease3.ID], vapiRelease4.ID)
}
