package instance

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type (
	EditInstanceInput struct {
		Name *string `json:"name"`
	}

	CreateInstanceInput struct {
		StackID uint                  `json:"stack_id"`
		Name    string                `json:"name"`
		Zone    tcltypes.InstanceZone `json:"zone"`
	}
)

func (s *service) CreateInstance(
	ctx context.Context,
	input CreateInstanceInput,
) (*domain.Instance, error) {
	if input.StackID == 0 {
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "stack id is required")
	}

	tx := helpers.GetTx(ctx)

	stack, err := domain.FindStackByID(tx, input.StackID)
	if err != nil {
		return nil, err
	}

	zone := tcltypes.InstanceZoneDefault
	if input.Zone != "" {
		zone = input.Zone
	}

	if input.Name == "" {
		input.Name = stack.Name + "-" + zone.String()
	}

	instance := domain.Instance{
		StackID: input.StackID,
		Zone:    zone,
		Name:    input.Name,
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		return instance.Save(tx)
	}); err != nil {
		return nil, err
	}

	return &instance, nil
}

func (s *service) DeleteInstance(
	ctx context.Context,
	instanceId uint,
	force bool,
) error {
	tx := helpers.GetTx(ctx)

	instance, err := domain.FindInstanceById(
		tx.
			Preload("Stack").
			Preload("Stack.Project"),
		instanceId,
	)
	if err != nil {
		return err
	}

	k8sClient, err := s.k8sClientPool.GetClient(instance.Zone)
	if err != nil {
		return err
	}

	if !force && instance.State == domain.InstanceStateRunning {
		return errors.Wrapf(tclerrors.ErrBadRequest, "instance is running")
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		instance.State = domain.InstanceStateNone
		if err := instance.Save(tx); err != nil {
			return err
		}

		if err := instance.Delete(tx); err != nil {
			return err
		}

		if err := s.deleteNamespace(ctx, k8sClient, &instance.Stack, force); err != nil {
			return err
		}

		return nil
	})
}

func (s *service) EditInstance(
	ctx context.Context,
	instanceId uint,
	input EditInstanceInput,
) error {
	tx := helpers.GetTx(ctx)

	instance, err := domain.FindInstanceById(tx, instanceId)
	if err != nil {
		return err
	}

	if input.Name != nil {
		instance.Name = *input.Name
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		return instance.Save(tx)
	})
}
