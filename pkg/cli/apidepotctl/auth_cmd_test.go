package apidepotctl_test

import (
	"context"
	tcli "github.com/habiliai/apidepot/pkg/cli/apidepotctl"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func (s *ApiDepotCtlTestSuite) TestLoginCmd() {
	// given
	authUserId := s.session.User.ID.String()

	cmd := s.cli.NewRootCmd()
	s.cloudServer.On("GetUser", mock.MatchedBy(func(ctx context.Context) bool {
		token := helpers.GetAuthToken(ctx)
		s.NotEmpty(token)
		return true
	}), mock.Anything).
		Return(&proto.User{
			Id:     1,
			AuthId: authUserId,
			Profile: &proto.UserProfile{
				Name: "Dennis Park",
			},
		}, nil).Once()
	s.cloudServer.On("RegisterCliApp", mock.MatchedBy(func(ctx context.Context) bool {
		token := helpers.GetAuthToken(ctx)
		s.NotEmpty(token)
		return true
	}), mock.MatchedBy(func(req *proto.RegisterCliAppRequest) bool {
		s.NotEmpty(req.Host)
		s.NotEmpty(req.RefreshToken)
		return true
	})).Return(&proto.RegisterCliAppResponse{
		AppId:     "test123",
		AppSecret: "test123123",
	}, nil).Once()
	defer s.cloudServer.AssertExpectations(s.T())

	// when
	configFile := "./testdata/auth_cmd_test.yaml"
	cmd.SetIn(strings.NewReader("test"))
	s.T().Logf("grpcAddr: %s", s.grpcAddr)
	cmd.SetArgs([]string{
		"login",
		"--server", s.grpcAddr,
		"--apiKey", constants.GotrueAnonKeyForTest,
		"--email", s.sessionEmail,
		"--password", s.sessionPassword,
		"-f", configFile,
	})
	if err := os.Remove(configFile); err != nil {
		s.T().Logf("not exists config file: %s", configFile)
	}

	// then
	s.Require().NoError(cmd.ExecuteContext(s))
}

func (s *ApiDepotCtlTestSuite) TestLogoutCmd() {
	// given
	accessToken := s.session.AccessToken
	bin, err := os.ReadFile("./testdata/auth_cmd_logout_before.yaml")
	s.Require().NoError(err)
	configFile := "./testdata/auth_cmd_logout_test.yaml"
	err = os.WriteFile(configFile, bin, 0644)
	s.Require().NoError(err)

	cmd := s.cli.NewRootCmd()
	s.cloudServer.On("VerifyCliApp", mock.Anything, mock.MatchedBy(func(req *proto.VerifyCliAppRequest) bool {
		s.Equal("test123", req.AppId)
		s.Equal("test123123", req.AppSecret)
		return true
	})).Return(&proto.VerifyCliAppResponse{
		AccessToken: accessToken,
	}, nil).Once()
	s.cloudServer.On("DeleteCliApp", mock.MatchedBy(func(ctx context.Context) bool {
		token := helpers.GetAuthToken(ctx)
		s.NotEmpty(token)
		return true
	}), mock.MatchedBy(func(req *proto.DeleteCliAppRequest) bool {
		s.Equal("test123", req.AppId)
		return true
	})).Return(&emptypb.Empty{}, nil).Once()
	defer s.cloudServer.AssertExpectations(s.T())

	// when
	cmd.SetIn(strings.NewReader("test"))
	s.T().Logf("grpcAddr: %s", s.grpcAddr)
	cmd.SetArgs([]string{
		"logout",
		"--server", s.grpcAddr,
		"--apiKey", constants.GotrueAnonKeyForTest,
		"-f", configFile,
	})

	// then
	s.Require().NoError(cmd.ExecuteContext(s))

	var config tcli.CliConfig
	configBin, err := os.ReadFile(configFile)
	s.Require().NoError(err)
	s.Require().NoError(yaml.Unmarshal(configBin, &config))

	s.Empty(config.Session.AppId)
	s.Empty(config.Session.AppSecret)
}
