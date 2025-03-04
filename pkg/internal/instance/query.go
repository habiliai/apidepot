package instance

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
)

func (s *service) GetInstance(
	ctx context.Context,
	instanceId uint,
) (*domain.Instance, error) {
	tx := helpers.GetTx(ctx)

	user, err := s.users.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	instance, err := domain.FindInstanceById(
		tx.Preload("Stack").
			Preload("Stack.Project").
			Preload("Stack.Vapis").
			Preload("Stack.Vapis.Vapi").
			Preload("Stack.Vapis.Vapi.Package"),
		instanceId,
	)
	if err != nil {
		return nil, err
	}

	if instance.Stack.Project.OwnerID != user.ID {
		return nil, errors.Wrapf(tclerrors.ErrForbidden, "user is not owner of the project")
	}

	return instance, nil
}

func (s *service) GetInstancesInStack(
	ctx context.Context,
	stackId uint,
) ([]domain.Instance, error) {
	tx := helpers.GetTx(ctx)

	if _, err := s.stacks.GetStack(ctx, stackId); err != nil {
		return nil, err
	}

	return domain.FindInstancesByStackId(tx, stackId)
}
