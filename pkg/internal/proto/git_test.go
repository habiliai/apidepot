package proto_test

import (
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

func (s *ProtoTestSuite) TestHardForkGitRepo() {
	// Given
	ctx := s.ctx
	s.git.On("CopyRepo", mock.Anything, "https://github.com/source/repo", "https://<token>@github.com/dest/repo").Return(nil).Once()
	defer s.git.AssertExpectations(s.T())

	client, closeConn := s.newClient()
	defer func() { s.Require().NoError(closeConn()) }()

	ctx = metadata.AppendToOutgoingContext(ctx, "x-github-token", "<token>")

	// When
	_, err := client.HardForkGitRepo(ctx, &proto.HardForkGitRepoRequest{
		SrcGitRepo: "source/repo",
		DstGitRepo: "dest/repo",
	})

	// Then
	s.Require().NoError(err)
}
