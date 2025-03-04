package organization

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (s *service) UpdateOrganization(ctx context.Context, input CreateOrUpdateOrganizationInput) (uint, error) {
	var org domain.Organization
	if err := helpers.GetTx(ctx).Transaction(func(tx *gorm.DB) error {
		if input.Id != nil {
			if r := tx.Find(&org, *input.Id); r.Error != nil {
				return errors.Wrapf(r.Error, "failed to find organization by id")
			} else if r.RowsAffected == 0 && input.NoCreate {
				return errors.Wrapf(tclerrors.ErrPreconditionRequired, "organization not found")
			}
		}

		logger.Debug("organization found")

		if input.Name != nil {
			org.Name = *input.Name
		}

		return errors.Wrapf(tx.Save(&org).Error, "failed to save organization")
	}); err != nil {
		return 0, err
	}

	return org.ID, nil
}

func (s *service) DeleteOrganization(ctx context.Context, id uint) error {
	return helpers.GetTx(ctx).Transaction(func(tx *gorm.DB) error {
		if r := tx.Delete(&domain.Organization{}, "id = ?", id); r.Error != nil {
			return errors.Wrapf(r.Error, "failed to delete organization")
		} else if r.RowsAffected == 0 {
			return errors.Wrapf(tclerrors.ErrNotFound, "organization not found")
		}

		return nil
	})
}
