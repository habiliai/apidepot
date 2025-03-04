package proto_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/stretchr/testify/mock"
)

func (s *ProtoTestSuite) TestStackEnableVapi() {
	var (
		stackId uint = 2
		vapiId  uint = 3
	)

	s.stacks.On(
		"EnableVapi",
		mock.Anything,
		mock.MatchedBy(func(arg uint) bool {
			return stackId == arg
		}),
		mock.MatchedBy(func(input stack.EnableVapiInput) bool {
			return vapiId == input.VapiID
		}),
	).Return(&domain.StackVapi{
		StackID: stackId,
		VapiID:  vapiId,
	}, nil).Once()
	defer s.stacks.AssertExpectations(s.T())

	client, dispose := s.newClient()
	defer dispose()

	resp, err := client.InstallVapi(s.Context(), &proto.InstallVapiRequest{
		StackId: int32(stackId),
		VapiId:  int32(vapiId),
	})

	s.Require().NoError(err)

	s.Equal(stackId, uint(resp.StackId))
	s.Equal(vapiId, uint(resp.VapiId))
}

func (s *ProtoTestSuite) TestDisableVapis() {
	var (
		stackId uint = 2
		vapiId  uint = 1
	)

	s.stacks.On(
		"DisableVapi",
		mock.Anything,
		stackId,
		vapiId,
	).Return(nil).Once()
	defer s.stacks.AssertExpectations(s.T())

	client, dispose := s.newClient()
	defer dispose()

	resp, err := client.UninstallVapi(s.Context(), &proto.UninstallVapiRequest{
		StackId: int32(stackId),
		VapiId:  int32(vapiId),
	})
	s.Require().NoError(err)
	s.NotNil(resp)
}
