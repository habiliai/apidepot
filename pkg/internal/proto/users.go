package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *apiDepotServer) GetUser(ctx context.Context, _ *emptypb.Empty) (*User, error) {
	user, err := s.userService.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	return newUserPbFromDb(*user), nil
}

func (s *apiDepotServer) UpdateUserProfile(
	ctx context.Context,
	req *UpdateUserProfileRequest,
) (*emptypb.Empty, error) {
	tx := helpers.GetTx(ctx)
	user, err := s.userService.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	user.Name = req.Profile.Name
	user.Description = req.Profile.Description
	user.GithubEmail = req.Profile.GithubEmail
	user.GithubUsername = req.Profile.GithubUsername
	user.MediumLink = req.Profile.MediumLink
	user.AvatarUrl = req.Profile.AvatarUrl

	logger.Debug("updating user profile", "user", user)

	if err := user.Save(tx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) UpdateGithubInstallationInfo(
	ctx context.Context,
	req *UpdateUserGithubInstallationInfoRequest,
) (*emptypb.Empty, error) {
	accessToken, err := s.githubClient.FetchUserAccessToken(ctx, req.AuthCode)
	if err != nil {
		return nil, err
	}

	verified, err := s.githubClient.VerifyInstallation(ctx, accessToken, req.InstallationId)
	if err != nil {
		return nil, err
	}

	if !verified {
		return nil, errors.New("failed to verify installation")
	}

	if err := s.userService.UpdateGithubAccessToken(ctx, accessToken); err != nil {
		return nil, err
	}

	if err := s.userService.UpdateGithubInstallationId(ctx, req.InstallationId); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GenerateInstallationAccessToken(ctx context.Context, _ *emptypb.Empty) (*GenerateInstallationAccessTokenResponse, error) {
	user, err := s.userService.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	installationId := user.GithubInstallationId
	if installationId == 0 {
		return nil, errors.New("installation id is not set")
	}

	accessToken, err := s.githubClient.GenerateInstallationAccessToken(ctx, installationId)
	if err != nil {
		return nil, err
	}

	return &GenerateInstallationAccessTokenResponse{
		Token: accessToken,
	}, nil
}

func (s *apiDepotServer) SyncExistingInstallation(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	githubToken := helpers.GetGithubToken(ctx)
	if githubToken == "" {
		logger.Debug("github token is not found")
		return &emptypb.Empty{}, nil
	}

	user, err := s.userService.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	installationId := user.GithubInstallationId
	if installationId != 0 {
		logger.Debug("installation id is already set", "installationId", installationId)
		return &emptypb.Empty{}, nil
	}

	githubUser, err := s.githubClient.GetUser(ctx, githubToken)
	if err != nil {
		return nil, err
	}
	existingInstallationId, err := s.githubClient.GetExistingInstallationId(ctx, githubUser.GetLogin())
	if err != nil {
		return nil, err
	}

	if existingInstallationId == 0 {
		logger.Debug("installation id is not found from github")
		return &emptypb.Empty{}, nil
	}

	if err := s.userService.UpdateGithubInstallationId(ctx, existingInstallationId); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetUserStorageUsages(ctx context.Context, _ *emptypb.Empty) (*GetUserStorageUsagesResponse, error) {
	usages, err := s.userService.GetStorageUsagesLatest(ctx)
	if err != nil {
		return nil, err
	}

	resp := GetUserStorageUsagesResponse{
		Average: usages.Average,
		Overage: usages.Overage,
	}

	resp.AveragesInPeriod = make([]*GetUserStorageUsagesResponse_DailyAverage, 0, len(usages.AverageInPeriod))
	for _, aip := range usages.AverageInPeriod {
		resp.AveragesInPeriod = append(resp.AveragesInPeriod, &GetUserStorageUsagesResponse_DailyAverage{
			Average: aip.Average,
			Date:    timestamppb.New(aip.Date),
		})
	}

	return &resp, nil
}
