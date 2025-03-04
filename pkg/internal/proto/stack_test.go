package proto_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/mokiat/gog"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

func (s *ProtoTestSuite) TestPatchStack() {
	s.stacks.On("PatchStack", mock.Anything, uint(1), mock.MatchedBy(func(input stack.PatchStackInput) bool {
		if !s.NotNil(input.SiteURL) {
			return false
		}
		return s.Equal("http://localhost:3000", *input.SiteURL)
	})).Return(nil).Once()
	defer s.stacks.AssertExpectations(s.T())

	client, dispose := s.newClient()
	defer dispose()

	_, err := client.UpdateStack(s.Context(), &proto.UpdateStackRequest{
		StackId: 1,
		SiteUrl: gog.PtrOf("http://localhost:3000"),
	})
	s.Require().NoError(err)
}

func (s *ProtoTestSuite) TestCreateStack() {
	s.stacks.On("CreateStack", mock.Anything, stack.CreateStackInput{
		ProjectID:   1,
		Name:        "test11",
		SiteURL:     "http://localhost:3000",
		Description: "",
	}).Return(&domain.Stack{
		Model: domain.Model{
			ID: 1,
		},
		ProjectID: 1,
		Name:      "test11",
		SiteURL:   "http://localhost:3000",
	}, nil).Once()
	defer s.stacks.AssertExpectations(s.T())

	input := proto.CreateStackRequest{
		ProjectId: 1,
		Name:      "test11",
		SiteUrl:   "http://localhost:3000",
	}

	client, dispose := s.newClient()
	defer dispose()

	resp, err := client.CreateStack(s.Context(), &input)
	s.Require().NoError(err)

	s.Equal("test11", resp.Name)
	s.Equal(int32(1), resp.ProjectId)
	s.Equal("http://localhost:3000", resp.SiteUrl)
}

func (s *ProtoTestSuite) TestGetStack() {
	s.stacks.On("GetStack", mock.Anything, uint(1)).Return(&domain.Stack{
		Model: domain.Model{
			ID: 1,
		},
		Name: "test",
	}, nil)
	defer s.stacks.AssertExpectations(s.T())

	client, dispose := s.newClient()
	defer dispose()

	resp, err := client.GetStackById(s.Context(), &proto.StackId{
		Id: 1,
	})
	s.Require().NoError(err)

	s.Equal("test", resp.Name)
	s.Equal(int32(1), resp.Id)
}

func (s *ProtoTestSuite) TestDeleteStack() {
	s.stacks.On("DeleteStack", mock.Anything, uint(1)).Return(nil)
	defer s.stacks.AssertExpectations(s.T())

	client, dispose := s.newClient()
	defer dispose()

	_, err := client.DeleteStack(s.Context(), &proto.StackId{
		Id: 1,
	})
	s.Require().NoError(err)
}

func (s *ProtoTestSuite) TestGetStacks() {
	s.Run("given with a stack, when to get a stack, should be ok", func() {
		st := domain.Stack{
			ProjectID: 1,
			Name:      "test-1",
		}
		s.stacks.On("GetStacks", mock.Anything, st.ProjectID, mock.Anything, mock.Anything, mock.Anything).
			Return([]domain.Stack{st}, nil)
		defer s.stacks.AssertExpectations(s.T())

		client, dispose := s.newClient()
		defer dispose()

		resp, err := client.GetStacks(s.Context(), &proto.GetStacksRequest{
			ProjectId: 1,
		})
		s.Require().NoError(err)
		s.Require().Len(resp.Stacks, 1)
	})

	s.Run("given a stack with auth enabled, when to get a stack, should be get stack without jwt secret", func() {
		st := domain.Stack{
			ProjectID:   1,
			Name:        "test-2",
			AuthEnabled: true,
			Auth: datatypes.NewJSONType(domain.Auth{
				JWTSecret:            "test123",
				ExternalEmailEnabled: true,
			}),
		}
		s.stacks.On(
			"GetStacks",
			mock.Anything,
			st.ProjectID,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return([]domain.Stack{st}, nil)
		defer s.stacks.AssertExpectations(s.T())

		client, dispose := s.newClient()
		defer dispose()

		resp, err := client.GetStacks(s.Context(), &proto.GetStacksRequest{
			ProjectId: 1,
		})
		s.Require().NoError(err)
		s.Require().Len(resp.Stacks, 1)
		s.Require().Empty(resp.Stacks[0].Auth.JwtSecret)
	})
}
