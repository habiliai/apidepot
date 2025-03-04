package organization_test

import (
	"github.com/habiliai/apidepot/pkg/internal/organization"
	"github.com/mokiat/gog"
)

func (s *OrganizationTestSuite) TestUpdateOrganization() {
	// Given
	input := organization.CreateOrUpdateOrganizationInput{
		Name: gog.PtrOf("test"),
	}

	// When
	orgId, err := s.orgService.UpdateOrganization(s.context, input)

	// Then
	s.Require().NoError(err)
	org, err := s.orgService.GetOrganizationById(s.context, orgId)
	s.Require().NoError(err)
	s.Require().Equal("test", org.Name)
}

func (s *OrganizationTestSuite) TestUpdateOrganizationGivenNoCreateWhenCreateOrganizationShouldBeError() {
	// Given
	input := organization.CreateOrUpdateOrganizationInput{
		Id:       gog.PtrOf(uint(1)),
		Name:     gog.PtrOf("test1"),
		NoCreate: true,
	}

	// When
	_, err := s.orgService.UpdateOrganization(s.context, input)

	// Then
	s.Require().Error(err)
}
