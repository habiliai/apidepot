package histories_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/stretchr/testify/mock"
)

func (s *HistoriesTestSuite) TestScanRunningAllInstances() {
	// given
	st := domain.Stack{
		Name: "test-stack",
		Project: domain.Project{
			Name: "test-project",
			Owner: domain.User{
				Name: "test",
			},
		},
	}
	s.Require().NoError(st.Save(s.db))
	instance1 := domain.Instance{
		Name:  "test-instance-1",
		State: domain.InstanceStateRunning,
		Stack: st,
	}
	s.Require().NoError(instance1.Save(s.db))
	instance2 := domain.Instance{
		Name:  "test-instance-2",
		State: domain.InstanceStateRunning,
		Stack: st,
	}
	s.Require().NoError(instance2.Save(s.db))
	instance3 := domain.Instance{
		Name:  "test-instance-3",
		State: domain.InstanceStateRunning,
		Stack: st,
	}
	s.Require().NoError(instance3.Save(s.db))

	// when
	err := s.service.WriteInstanceHistoriesAt(s)

	// then
	s.Require().NoError(err)

	var instanceHistory domain.InstanceHistory
	s.Require().NoError(s.db.First(&instanceHistory, "instance_id = ?", instance1.ID).Error)
	s.Require().True(instanceHistory.Running)

	instanceHistory = domain.InstanceHistory{}
	s.Require().NoError(s.db.First(&instanceHistory, "instance_id = ?", instance2.ID).Error)
	s.Require().True(instanceHistory.Running)

	instanceHistory = domain.InstanceHistory{}
	s.Require().NoError(s.db.First(&instanceHistory, "instance_id = ?", instance3.ID).Error)
	s.Require().True(instanceHistory.Running)
}

func (s *HistoriesTestSuite) TestWriteStackHistoriesAt() {
	// given
	project := domain.Project{
		Name: "test-project",
		Owner: domain.User{
			Name: "test",
		},
	}
	s.Require().NoError(project.Save(s.db))

	s.userService.On("GetUser", mock.Anything).Return(&project.Owner, nil).Twice()
	defer s.userService.AssertExpectations(s.T())

	stack1, err := s.stackService.CreateStack(s, stack.CreateStackInput{
		ProjectID:     project.ID,
		Name:          "test-stack-1",
		SiteURL:       "localhost:3000",
		DefaultRegion: tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)

	stack2, err := s.stackService.CreateStack(s, stack.CreateStackInput{
		Name:          "test-stack-2",
		ProjectID:     project.ID,
		SiteURL:       "localhost:3001",
		DefaultRegion: tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)

	// when
	err = s.service.WriteStackHistoriesAt(s)

	// then
	s.Require().NoError(err)

	var stackHistory domain.StackHistory
	s.Require().NoError(s.db.First(&stackHistory, "stack_id = ?", stack1.ID).Error)
	s.Require().Equal(stack1.ID, stackHistory.StackID)

	stackHistory = domain.StackHistory{}
	s.Require().NoError(s.db.First(&stackHistory, "stack_id = ?", stack2.ID).Error)
	s.Require().Equal(stack2.ID, stackHistory.StackID)
}
