package proto

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/instance"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*Instance, error) {
	var zone tcltypes.InstanceZone
	switch req.Zone {
	case Instance_InstanceZoneNone:
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "zone is required")
	case Instance_InstanceZoneDefault:
		zone = tcltypes.InstanceZoneDefault
	case Instance_InstanceZoneOciApSeoul:
		zone = tcltypes.InstanceZoneOciApSeoul
	case Instance_InstanceZoneOciSingapore:
		zone = tcltypes.InstanceZoneOciSingapore
	default:
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "invalid zone: %s", req.Zone.String())
	}

	inst, err := s.instanceService.CreateInstance(ctx, instance.CreateInstanceInput{
		Name:    req.Name,
		StackID: uint(req.StackId),
		Zone:    zone,
	})
	if err != nil {
		return nil, err
	}

	return newInstancePbFromDb(inst), nil
}

func (s *apiDepotServer) GetInstanceById(ctx context.Context, id *InstanceId) (*Instance, error) {
	inst, err := s.instanceService.GetInstance(ctx, uint(id.GetId()))
	if err != nil {
		return nil, err
	}

	return newInstancePbFromDb(inst), nil
}

func (s *apiDepotServer) EditInstance(ctx context.Context, req *EditInstanceRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.instanceService.EditInstance(ctx, uint(req.GetId()), instance.EditInstanceInput{
		Name: req.Name,
	})
}

func (s *apiDepotServer) DeleteInstance(ctx context.Context, id *InstanceId) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.instanceService.DeleteInstance(ctx, uint(id.Id), false)
}

func (s *apiDepotServer) DeployStack(ctx context.Context, req *DeployStackRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.instanceService.DeployStack(ctx, uint(req.Id), instance.DeployStackInput{
		Timeout: req.Timeout,
	})
}

func (s *apiDepotServer) LaunchInstance(ctx context.Context, id *InstanceId) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.instanceService.LaunchInstance(ctx, uint(id.Id))
}

func (s *apiDepotServer) StopInstance(ctx context.Context, id *InstanceId) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.instanceService.StopInstance(ctx, uint(id.Id))
}
