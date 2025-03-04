package histories_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"time"
)

func (s *HistoriesTestSuite) TestGetTotalRunningTimeForMyInstancesIn() {
	// given
	user := domain.User{
		Name:       "test",
		AuthUserId: "test",
	}
	s.Require().NoError(user.Save(s.db))

	stack := domain.Stack{
		Name: "test-stack",
		Project: domain.Project{
			Name:  "test-project",
			Owner: user,
		},
	}
	s.Require().NoError(stack.Save(s.db))

	instance1 := domain.Instance{
		Name:  "test-instance-1",
		State: domain.InstanceStateRunning,
		Stack: stack,
	}
	s.Require().NoError(instance1.Save(s.db))
	instance2 := domain.Instance{
		Name:  "test-instance-2",
		State: domain.InstanceStateRunning,
		Stack: stack,
	}
	s.Require().NoError(instance2.Save(s.db))
	instance3 := domain.Instance{
		Name:  "test-instance-3",
		State: domain.InstanceStateRunning,
		Stack: stack,
	}
	s.Require().NoError(instance3.Save(s.db))

	s.Require().NoError(s.service.WriteInstanceHistoriesAt(s))
	time.Sleep(5 * time.Second)
	s.Require().NoError(s.service.WriteInstanceHistoriesAt(s))
	time.Sleep(5 * time.Second)
	s.Require().NoError(s.service.WriteInstanceHistoriesAt(s))

	// when
	runningTimes, err := s.service.GetTotalRunningTimeForUser(s, nil, nil, user.ID)

	// then
	s.Require().NoError(err)
	s.Require().Len(runningTimes, 9)
	var totalTime time.Duration
	for _, rt := range runningTimes {
		totalTime += rt.Duration
	}
	s.Require().NotZero(totalTime)

	s.T().Logf("totalTime: %s", totalTime)
	s.Require().Equal(90, int(totalTime.Seconds()))
}
