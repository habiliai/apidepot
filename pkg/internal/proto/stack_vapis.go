package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) InstallVapi(ctx context.Context, request *InstallVapiRequest) (*StackVapi, error) {
	input := stack.EnableVapiInput{
		VapiID: uint(request.VapiId),
	}

	vapi, err := s.stackService.EnableVapi(ctx, uint(request.StackId), input)
	if err != nil {
		return nil, err
	}

	return newStackVapiPbFromDb(vapi), nil
}

func (s *apiDepotServer) UninstallVapi(ctx context.Context, request *UninstallVapiRequest) (*emptypb.Empty, error) {
	if err := s.stackService.DisableVapi(ctx, uint(request.StackId), uint(request.VapiId)); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) UpdateVapi(ctx context.Context, request *UpdateVapiRequest) (*StackVapi, error) {
	input := stack.UpdateVapiInput{
		Version: request.Version,
	}

	vapi, err := s.stackService.UpdateVapi(ctx, uint(request.StackId), uint(request.VapiId), input)
	if err != nil {
		return nil, err
	}

	return newStackVapiPbFromDb(vapi), nil
}

func (s *apiDepotServer) SetStackEnv(
	ctx context.Context,
	req *SetStackEnvRequest,
) (*emptypb.Empty, error) {
	envVars := make(map[string]string, len(req.EnvVars))
	for _, v := range req.EnvVars {
		envVars[v.Name] = v.Value
	}

	if err := s.stackService.SetVapiEnv(ctx, uint(req.StackId), envVars); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) UnsetStackEnv(
	ctx context.Context,
	req *UnsetStackEnvRequest,
) (*emptypb.Empty, error) {
	if err := s.stackService.UnsetVapiEnv(ctx, uint(req.StackId), req.EnvVarNames); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
