package histories

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"time"
)

type (
	InstanceRunningTime struct {
		InstanceId uint
		Duration   time.Duration
	}
)

func (s *service) GetTotalRunningTimeForUser(
	ctx context.Context,
	begin *time.Time,
	end *time.Time,
	userId uint,
) (instanceRunningTimes []InstanceRunningTime, err error) {
	tx := helpers.GetTx(ctx)

	var histories []domain.InstanceHistory
	stmt := tx.
		InnerJoins("Instance").
		InnerJoins("Instance.Stack").
		InnerJoins("Instance.Stack.Project").
		InnerJoins("Instance.Stack.Project.Owner", "Instance__Stack__Project__Owner.id = ?", userId)
	if begin != nil && end != nil {
		stmt = stmt.Where("instance_histories.created_at BETWEEN ? AND ?", *begin, *end)
	}
	stmt = stmt.Find(&histories, "instance_histories.running = true")
	if err = errors.Wrapf(stmt.Error, "failed to find instance histories"); err != nil {
		return
	}

	minTimes := map[uint]time.Time{}
	maxTimes := map[uint]time.Time{}
	for _, hist := range histories {
		minTime, ok := minTimes[hist.InstanceID]
		createdAt := hist.CreatedAt
		if !ok {
			minTime = createdAt
		} else if createdAt.Before(minTime) {
			minTime = createdAt
		}
		minTimes[hist.InstanceID] = minTime

		maxTime, ok := maxTimes[hist.InstanceID]
		if !ok {
			maxTime = createdAt
		} else if createdAt.After(maxTime) {
			maxTime = createdAt
		}
		maxTimes[hist.InstanceID] = maxTime
	}

	for _, hist := range histories {
		instanceRunningTimes = append(instanceRunningTimes, InstanceRunningTime{
			InstanceId: hist.InstanceID,
			Duration:   maxTimes[hist.InstanceID].Sub(minTimes[hist.InstanceID]),
		})
	}

	return
}
