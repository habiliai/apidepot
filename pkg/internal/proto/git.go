package proto

import (
	"context"
	"fmt"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) HardForkGitRepo(
	ctx context.Context,
	req *HardForkGitRepoRequest,
) (*emptypb.Empty, error) {
	githubToken := helpers.GetGithubToken(ctx)
	if githubToken == "" {
		return nil, errors.Wrapf(tclerrors.ErrForbidden, "Unable to proceed: GitHub token is missing. Please provide a valid token and try again.")
	}

	srcGitUrl := fmt.Sprintf("https://github.com/%s", req.SrcGitRepo)
	dstGitUrl := fmt.Sprintf("https://%s@github.com/%s", githubToken, req.DstGitRepo)

	if err := s.gitService.CopyRepo(ctx, srcGitUrl, dstGitUrl); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
