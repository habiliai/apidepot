package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/project"
	"github.com/mokiat/gog"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateProject 함수
func (s *apiDepotServer) CreateProject(ctx context.Context, req *CreateProjectRequest) (*Project, error) {
	project, err := s.projectService.CreateProject(ctx, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return newProjectPbFromDb(project), nil
}

// GetProjectById 함수
func (s *apiDepotServer) GetProjectById(ctx context.Context, req *ProjectId) (*Project, error) {
	id := req.GetId()

	project, err := s.projectService.GetProject(ctx, uint(id))
	if err != nil {
		return nil, err
	}

	return newProjectPbFromDb(project), nil
}

// DeleteProjectById 함수
func (s *apiDepotServer) DeleteProjectById(ctx context.Context, req *ProjectId) (*emptypb.Empty, error) {
	id := req.GetId()

	if err := s.projectService.DeleteProject(ctx, uint(id)); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetProjects 함수
func (s *apiDepotServer) GetProjects(ctx context.Context, req *GetProjectsRequest) (*GetProjectsResponse, error) {
	var input = project.GetProjectsInput{
		Name:    req.Name,
		Page:    int(req.Page),
		PerPage: int(req.PerPage),
	}

	projects, err := s.projectService.GetProjects(ctx, input)
	if err != nil {
		return nil, err
	}

	return &GetProjectsResponse{
		Projects: gog.Map(projects, func(p domain.Project) *Project {
			return newProjectPbFromDb(&p)
		}),
	}, nil
}

func (s *apiDepotServer) DeleteProject(ctx context.Context, id *ProjectId) (*emptypb.Empty, error) {
	if err := s.projectService.DeleteProject(ctx, uint(id.GetId())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
