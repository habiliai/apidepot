package organization

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
)

func (s *service) GetOrganizationById(ctx context.Context, id uint) (org domain.Organization, err error) {
	err = errors.Wrapf(helpers.GetTx(ctx).Where("id = ?", id).First(&org).Error, "failed to find organization by id")
	return
}

func (s *service) GetOrganizations(ctx context.Context, memberOwnerId *string) (orgs []domain.Organization, err error) {
	tx := helpers.GetTx(ctx)

	if memberOwnerId != nil {
		tx = tx.Joins("JOIN organization_members ON organizations.id = organization_members.organization_id").
			Joins("JOIN users ON organization_members.user_id = users.id").
			Where("users.auth_user_id = ?", *memberOwnerId)
	}

	tx = tx.Order("organizations.id ASC").Find(&orgs)
	err = errors.Wrapf(tx.Error, "failed to find organizations")
	return
}
