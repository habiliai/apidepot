package histories

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

func (s *service) WriteInstanceHistoriesAt(
	ctx context.Context,
) (err error) {
	tx := helpers.GetTx(ctx)

	return tx.Transaction(func(tx *gorm.DB) (err error) {
		cursor := uint(0)
		var eg errgroup.Group

		for {
			var instances []domain.Instance
			if err = tx.
				Order("id ASC").
				Limit(250).
				Find(&instances, "state = ? AND id > ?", domain.InstanceStateRunning, cursor).
				Error; err != nil {
				return errors.Wrapf(err, "failed to find instances")
			}
			if len(instances) == 0 {
				break
			}
			cursor = instances[len(instances)-1].ID

			eg.Go(func() (err error) {
				for _, instance := range instances {
					history := domain.InstanceHistory{
						InstanceID: instance.ID,
						Instance:   instance,
					}
					if instance.State == domain.InstanceStateRunning {
						history.Running = true
					}
					if err = history.Save(tx); err != nil {
						return
					}
				}

				return nil
			})
		}

		return eg.Wait()
	})
}

func (s *service) WriteStackHistoriesAt(
	ctx context.Context,
) error {
	tx := helpers.GetTx(ctx)

	return tx.Transaction(func(tx *gorm.DB) error {
		cursor := uint(0)
		var eg errgroup.Group

		for {
			var stacks []domain.Stack
			if err := tx.Order("id ASC").Limit(100).Find(&stacks, "id > ?", cursor).Error; err != nil {
				return errors.WithStack(err)
			}

			if len(stacks) == 0 {
				break
			}

			eg.Go(func() error {
				for _, stack := range stacks {
					storageSize, err := s.bucketService.GetTotalSize(ctx, stack.DefaultRegion, stack.Domain)
					if err != nil {
						return err
					}

					history := domain.StackHistory{
						StackID: stack.ID,
						Stack:   stack,

						StorageSize: int(storageSize),
					}
					if err := history.Save(tx); err != nil {
						return err
					}
				}

				return nil
			})

			cursor = stacks[len(stacks)-1].ID
		}

		return eg.Wait()
	})
}
