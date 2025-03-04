package proto_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/instance"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/stretchr/testify/mock"
)

func (s *ProtoTestSuite) TestCreateInstance() {
	s.instances.On("CreateInstance", mock.Anything, instance.CreateInstanceInput{
		Name:    "test",
		StackID: 1,
		Zone:    tcltypes.InstanceZoneOciApSeoul,
	}).Return(&domain.Instance{
		Model: domain.Model{
			ID: 1,
		},
		Zone:           tcltypes.InstanceZoneOciApSeoul,
		NumReplicas:    1,
		MaxReplicas:    1,
		State:          domain.InstanceStateReady,
		Name:           "test",
		AppliedK8sYaml: "",
		StackID:        1,
	}, nil).Once()
	defer s.instances.AssertExpectations(s.T())

	client, dispose := s.newClient()
	defer dispose()

	resp, err := client.CreateInstance(s.Context(), &proto.CreateInstanceRequest{
		Name:    "test",
		StackId: 1,
		Zone:    proto.Instance_InstanceZoneOciApSeoul,
	})
	s.Require().NoError(err)

	s.Equal("test", resp.Name)
	s.Equal(int32(1), resp.Id)
	s.Equal(int32(1), resp.StackId)
	s.Equal(int32(1), resp.NumReplicas)
	s.Equal(int32(1), resp.MaxReplicas)
	s.Equal(proto.Instance_InstanceZoneOciApSeoul, resp.Zone)
	s.Equal(proto.Instance_InstanceStateReady, resp.State)
}
