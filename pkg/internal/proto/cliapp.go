package proto

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) RegisterCliApp(ctx context.Context, req *RegisterCliAppRequest) (*RegisterCliAppResponse, error) {
	output, err := s.cliappService.RegisterCliApp(ctx, req.Host, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &RegisterCliAppResponse{
		AppId:     output.AppId,
		AppSecret: output.AppSecret,
	}, nil
}

func (s *apiDepotServer) DeleteCliApp(ctx context.Context, req *DeleteCliAppRequest) (*emptypb.Empty, error) {
	if err := s.cliappService.DeleteCliApp(ctx, req.AppId); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) VerifyCliApp(ctx context.Context, req *VerifyCliAppRequest) (*VerifyCliAppResponse, error) {
	cliApp, err := s.cliappService.VerifyCliApp(ctx, req.AppId, req.AppSecret)
	if err != nil {
		return nil, err
	}

	return &VerifyCliAppResponse{
		AccessToken: cliApp.AccessToken,
	}, nil
}
