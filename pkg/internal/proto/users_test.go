package proto_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/stretchr/testify/mock"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ProtoTestSuite) signUp() *gotruetypes.Session {
	resp, err := s.gotrueClient.Signup(gotruetypes.SignupRequest{
		Email:    "test@test.com",
		Password: "test1234",
	})
	s.Require().NoError(err)

	return &resp.Session
}

func (s *ProtoTestSuite) TestGetUser() {
	tcc, dispose := s.newClient()
	defer dispose()

	gotrueSession := s.signUp()

	expectedUser := domain.User{
		Name:           "test-1234",
		GithubEmail:    "test@test.com",
		GithubUsername: "JCooky",
		AuthUserId:     gotrueSession.User.ID.String(),
	}
	s.Require().NoError(expectedUser.Save(s.db))
	s.users.On("GetUser", mock.Anything).Return(&expectedUser, nil)
	defer s.users.AssertExpectations(s.T())

	targetUser, err := tcc.GetUser(
		s.ctx,
		&emptypb.Empty{},
		grpc.Header(&metadata.MD{
			"Authorization": {"Bearer " + gotrueSession.AccessToken},
		}),
	)
	s.Require().NoError(err)

	s.Equal(expectedUser.ID, uint(targetUser.GetId()))
}
