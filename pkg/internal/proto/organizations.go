package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/organization"
	"github.com/mokiat/gog"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) UpsertOrganization(ctx context.Context, request *UpsertOrganizationRequest) (*OrganizationId, error) {
	var id *uint
	if request.Id != nil {
		id = gog.PtrOf(uint(*request.Id))
	}

	orgId, err := s.orgService.UpdateOrganization(ctx, organization.CreateOrUpdateOrganizationInput{
		Id:   id,
		Name: request.Name,
	})
	if err != nil {
		return nil, err
	}

	return &OrganizationId{Id: int32(orgId)}, nil
}

func (s *apiDepotServer) DeleteOrganization(ctx context.Context, id *OrganizationId) (*emptypb.Empty, error) {
	err := s.orgService.DeleteOrganization(ctx, uint(id.Id))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetOrganization(ctx context.Context, id *OrganizationId) (*Organization, error) {
	org, err := s.orgService.GetOrganizationById(ctx, uint(id.Id))
	if err != nil {
		return nil, err
	}

	return newOrganizationPbFromDb(org), nil
}

func (s *apiDepotServer) GetAllOrganizations(ctx context.Context, req *GetAllOrganizationsRequest) (*GetAllOrganizationsResponse, error) {
	orgs, err := s.orgService.GetOrganizations(ctx, req.MemberId)
	if err != nil {
		return nil, err
	}

	organizations := make([]*Organization, len(orgs))
	for i, org := range orgs {
		organizations[i] = newOrganizationPbFromDb(org)
	}

	return &GetAllOrganizationsResponse{
		Organizations: organizations,
	}, nil
}
