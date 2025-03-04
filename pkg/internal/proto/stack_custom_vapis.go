package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/mokiat/gog"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) InstallCustomVapi(ctx context.Context, req *InstallCustomVapiRequest) (*emptypb.Empty, error) {
	if _, err := s.stackService.EnableCustomVapi(ctx, uint(req.StackId), stack.CreateCustomVapiInput{
		Name: req.Name,
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) UninstallCustomVapi(ctx context.Context, req *UninstallCustomVapiRequest) (*emptypb.Empty, error) {
	if err := s.stackService.DisableCustomVapi(ctx, uint(req.StackId), req.Name); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetCustomVapiByNameOnStack(ctx context.Context, req *GetCustomVapiByNameOnStackRequest) (*CustomVapi, error) {
	customVapi, err := s.stackService.GetCustomVapiByName(ctx, uint(req.StackId), req.Name)
	if err != nil {
		return nil, err
	}

	return newCustomVapiPbFromDb(*customVapi), nil
}

func (s *apiDepotServer) UpdateCustomVapi(ctx context.Context, req *UpdateCustomVapiRequest) (*emptypb.Empty, error) {
	if err := s.stackService.UpdateCustomVapi(ctx, uint(req.StackId), req.Name, stack.UpdateCustomVapiInput{
		NewName:       req.NewName,
		UpdateTarFile: req.UpdateTarFile,
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetCustomVapisOnStack(ctx context.Context, req *StackId) (*GetCustomVapisOnStackResponse, error) {
	customVapis, err := s.stackService.GetCustomVapis(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}

	return &GetCustomVapisOnStackResponse{
		CustomVapis: gog.Map(customVapis, newCustomVapiPbFromDb),
	}, nil
}
