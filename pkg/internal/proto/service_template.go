package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/mokiat/gog"
)

func (s *apiDepotServer) CreateStackFromServiceTemplate(
	ctx context.Context,
	req *CreateStackFromServiceTemplateRequest,
) (*Stack, error) {
	st, err := s.svctplService.CreateStackFromServiceTemplate(
		ctx,
		uint(req.ServiceTemplateId),
		stack.CreateStackInput{
			Name:          req.CreateStackRequest.Name,
			Description:   req.CreateStackRequest.Description,
			LogoImageUrl:  req.CreateStackRequest.LogoImageUrl,
			DefaultRegion: req.CreateStackRequest.DefaultRegion.ToDomain(),
			SiteURL:       req.CreateStackRequest.SiteUrl,
			ProjectID:     uint(req.CreateStackRequest.ProjectId),
			GitRepo:       req.GitRepo,
			GitBranch:     "main",
		},
	)
	if err != nil {
		return nil, err
	}

	return newStackPbFromDb(*st), nil
}

func (s *apiDepotServer) GetServiceTemplateById(
	ctx context.Context,
	req *ServiceTemplateId,
) (*ServiceTemplate, error) {
	st, err := s.svctplService.GetServiceTemplateByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}

	return newServiceTemplatePbFromDb(*st), nil
}

func (s *apiDepotServer) SearchServiceTemplates(
	ctx context.Context,
	req *SearchServiceTemplatesRequest,
) (*SearchServiceTemplatesResponse, error) {
	var cursor uint = 0
	if req.Cursor != nil {
		cursor = uint(*req.Cursor)
	}

	var limit uint = 0
	if req.Limit != nil {
		limit = uint(*req.Limit)
	}

	var searchQuery string
	if req.SearchQuery != nil {
		searchQuery = *req.SearchQuery
	}

	output, err := s.svctplService.SearchServiceTemplates(
		ctx,
		cursor,
		limit,
		searchQuery,
	)
	if err != nil {
		return nil, err
	}

	return &SearchServiceTemplatesResponse{
		NextCursor:       int32(output.NextCursor),
		NumTotal:         output.NumTotal,
		ServiceTemplates: gog.Map(output.ServiceTemplates, newServiceTemplatePbFromDb),
	}, nil
}
