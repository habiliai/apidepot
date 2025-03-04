package apidepotctl_test

import (
	"context"
	tcli "github.com/habiliai/apidepot/pkg/cli/apidepotctl"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v3"
	"os"
)

func (s *ApiDepotCtlTestSuite) TestUpdateStackCmd() {
	s.Require().NoError(util.CopyFile("./testdata/stack_cmd_test.orig.yaml", "./testdata/stack_cmd_test.yaml", true))

	authTokenMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		token := helpers.GetAuthToken(ctx)
		s.NotEmpty(token)
		return true
	})
	s.cloudServer.On("VerifyCliApp", mock.Anything, mock.Anything).Return(&proto.VerifyCliAppResponse{
		AccessToken: s.session.AccessToken,
	}, nil).Once()
	s.cloudServer.On("GetProjects", authTokenMatcher, mock.Anything).Return(&proto.GetProjectsResponse{
		Projects: []*proto.Project{
			{
				Id: 1,
			},
		},
	}, nil).Once()
	s.cloudServer.On("GetStacks", authTokenMatcher, mock.MatchedBy(func(req *proto.GetStacksRequest) bool {
		s.Equal(int32(1), req.ProjectId)
		s.Equal("test-stack", *req.Name)

		return true
	})).Return(&proto.GetStacksResponse{
		Stacks: []*proto.Stack{
			{
				Id: 1,
			},
		},
	}, nil).Once()
	s.cloudServer.On("UpdateStack", authTokenMatcher, mock.MatchedBy(func(req *proto.UpdateStackRequest) bool {
		s.Equal(int32(1), req.StackId)
		if s.NotNil(req.SiteUrl) {
			s.Equal("http://localhost:8080", *req.SiteUrl)
		}
		return true
	})).Return(&emptypb.Empty{}, nil).Once()
	s.cloudServer.On("InstallAuth", authTokenMatcher, mock.MatchedBy(func(req *proto.InstallAuthRequest) bool {
		s.Equal(int32(1), req.Id)
		s.True(req.IsUpdate)
		if s.NotNil(req.MailerAutoConfirm) {
			s.True(*req.MailerAutoConfirm)
		}
		return true
	})).Return(&emptypb.Empty{}, nil).Once()
	s.cloudServer.On("InstallStorage", authTokenMatcher, mock.MatchedBy(func(req *proto.InstallStorageRequest) bool {
		s.Equal(int32(1), req.Id)
		s.True(req.IsUpdate)
		return true
	})).Return(&emptypb.Empty{}, nil).Once()
	s.cloudServer.On("InstallPostgrest", authTokenMatcher, mock.MatchedBy(func(req *proto.InstallPostgrestRequest) bool {
		s.Equal(int32(1), req.Id)
		s.True(req.IsUpdate)
		return true
	})).Return(&emptypb.Empty{}, nil).Once()
	defer s.cloudServer.AssertExpectations(s.T())

	// when
	cmd := s.cli.NewRootCmd()
	cmd.SetArgs([]string{
		"stack",
		"update",
		"--server", s.grpcAddr,
		"--apiKey", constants.GotrueAnonKeyForTest,
		"--stack.name", "test-stack",
		"--stack.siteUrl", "http://localhost:8080",
		"-f", "./testdata/stack_cmd_test.yaml",
		"--stack.auth.mailer.autoConfirm", "true",
	})

	// then
	s.Require().NoError(cmd.ExecuteContext(s))

	contents, err := os.ReadFile("./testdata/stack_cmd_test.yaml")
	s.Require().NoError(err)
	var config tcli.CliConfig
	s.Require().NoError(yaml.Unmarshal(contents, &config))

	s.True(config.Stack.Auth.Mailer.AutoConfirm)
}
