package apidepotctl_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ApiDepotCtlTestSuite) TestSetStackEnvCmd() {
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
	s.cloudServer.On("SetStackEnv", mock.Anything, mock.MatchedBy(func(req *proto.SetStackEnvRequest) bool {
		s.Equal(int32(1), req.StackId)
		if s.Len(req.EnvVars, 1) {
			s.Equal("exampleKey", req.EnvVars[0].Name)
			s.Equal("exampleValue", req.EnvVars[0].Value)
		}

		return true
	})).Return(&emptypb.Empty{}, nil).Once()
	defer s.cloudServer.AssertExpectations(s.T())

	cmd := s.cli.NewRootCmd()
	cmd.SetArgs([]string{
		"stack", "env", "set", "exampleKey=exampleValue",
		"-f", "./testdata/stack_cmd_test.yaml",
		"--stack.name", "test-stack",
	})

	err := cmd.Execute()
	s.NoError(err)
}

func (s *ApiDepotCtlTestSuite) TestUnsetStackEnv() {
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
	s.cloudServer.On("UnsetStackEnv", mock.Anything, mock.MatchedBy(func(req *proto.UnsetStackEnvRequest) bool {
		s.Equal(int32(1), req.StackId)
		if s.Len(req.EnvVarNames, 1) {
			s.Equal("exampleKey", req.EnvVarNames[0])
		}

		return true
	})).Return(&emptypb.Empty{}, nil).Once()
	defer s.cloudServer.AssertExpectations(s.T())

	cmd := s.cli.NewRootCmd()
	cmd.SetArgs([]string{
		"stack", "env", "unset", "exampleKey",
		"-f", "./testdata/stack_cmd_test.yaml",
		"--stack.name", "test-stack",
	})

	err := cmd.Execute()
	s.NoError(err)
}
