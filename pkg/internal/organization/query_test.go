package organization_test

import (
	"github.com/habiliai/apidepot/pkg/internal/organization"
	"github.com/mokiat/gog"
)

func (s *OrganizationTestSuite) TestGetAllOrganizationsGivenTwoCreatedOrganizationWhenGetAllOrganizationsShouldBeReturnedTwo() {
	// Given
	{
		input := organization.CreateOrUpdateOrganizationInput{
			Name: gog.PtrOf("test"),
		}
		_, err := s.orgService.UpdateOrganization(s.context, input)
		s.Require().NoError(err)
	}
	{
		input := organization.CreateOrUpdateOrganizationInput{
			Name: gog.PtrOf("test1"),
		}
		_, err := s.orgService.UpdateOrganization(s.context, input)
		s.Require().NoError(err)
	}

	// When
	orgs, err := s.orgService.GetOrganizations(s.context, nil)

	// Then
	s.Require().NoError(err)
	s.Require().Len(orgs, 2)
	s.T().Logf("orgs: %v", orgs)
	s.Require().Equal("test", orgs[0].Name)
	s.Require().Equal("test1", orgs[1].Name)
}
