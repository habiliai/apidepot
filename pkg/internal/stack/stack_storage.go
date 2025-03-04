package stack

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StorageInput struct {
	TenantID *string `json:"tenant_id"`
}

type EnableOrUpdateStorageInput struct {
	StorageInput
}

func (ss *service) EnableOrUpdateStorage(
	ctx context.Context,
	stackId uint,
	input EnableOrUpdateStorageInput,
	isCreate bool,
) error {
	tx := helpers.GetTx(ctx)

	stack, err := ss.GetStack(ctx, stackId)
	if err != nil {
		return err
	}

	if err := ss.hasPermission(ctx, stack.ProjectID); err != nil {
		return err
	}

	if !stack.AuthEnabled {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "required to enable auth")
	}
	if stack.StorageEnabled && isCreate {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "storage is already created")
	} else if !stack.StorageEnabled && !isCreate {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "storage is not enabled")
	}

	stack.StorageEnabled = true

	// prepare default values for first enable storage as a creating
	storage := domain.Storage{
		S3Bucket: stack.Domain,
		TenantID: "root",
	}
	if !isCreate {
		storage = stack.Storage.Data()
	}

	// override default values if input is not empty
	if input.TenantID != nil && *input.TenantID != "" {
		storage.TenantID = *input.TenantID
	}

	stack.Storage = datatypes.NewJSONType(storage)

	if err := tx.Transaction(func(tx *gorm.DB) (error error) {
		if err := tx.Omit(clause.Associations).Save(stack).Error; err != nil {
			return err
		}

		defer func() {
			if error == nil || !isCreate {
				return
			}

			if err := ss.bucketService.DeleteBucket(ctx, stack.DefaultRegion, stack.Name); err != nil {
				logger.Warn("failed to delete bucket", "err", err)
			}
		}()
		if isCreate {
			if err := ss.bucketService.CreateBucket(ctx, stack.DefaultRegion, storage.S3Bucket); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {

		return err
	}

	return nil
}

func (ss *service) DisableStorage(ctx context.Context, stackId uint) error {
	tx := helpers.GetTx(ctx)

	stack, err := ss.GetStack(ctx, stackId)
	if err != nil {
		return err
	}

	if !stack.StorageEnabled {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "storage is not enabled")
	}

	stack.StorageEnabled = false

	return tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Save(stack).Error; err != nil {
			return err
		}

		if err := ss.bucketService.DeleteBucket(ctx, stack.DefaultRegion, stack.Domain); err != nil {
			return err
		}

		return nil
	})
}

func (ss *service) GetStorageUsage(ctx context.Context, stackId uint) (int64, error) {
	stack, err := ss.GetStack(ctx, stackId)
	if err != nil {
		return 0, err
	}

	if !stack.StorageEnabled {
		return 0, errors.Wrapf(tclerrors.ErrPreconditionRequired, "storage is not enabled")
	}

	return ss.bucketService.GetTotalSize(ctx, stack.DefaultRegion, stack.Domain)
}

func (s *service) GetMyTotalStorageUsage(
	ctx context.Context,
) (int64, error) {
	tx := helpers.GetTx(ctx)

	user, err := s.users.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	var stacks []domain.Stack
	if err := tx.InnerJoins("Project", tx.Where("Project.OwnerID = ?", user.ID)).
		Find(&stacks).Error; err != nil {
		return 0, errors.Wrapf(err, "failed to find stacks")
	}

	if len(stacks) == 0 {
		return 0, nil
	}

	totalUsage := int64(0)
	for _, stack := range stacks {
		usage, err := s.GetStorageUsage(ctx, stack.ID)
		if err != nil {
			return 0, err
		}

		totalUsage += usage
	}

	return totalUsage, nil
}
