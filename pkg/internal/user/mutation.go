package user

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (s *service) AddMemberToOrg(ctx context.Context, orgId uint, memberOwnerId string) error {
	var org domain.Organization
	if err := helpers.GetTx(ctx).First(&org, orgId).Error; err != nil {
		return errors.Wrapf(err, "failed to find organization by id")
	}

	user, err := s.GetUserByAuthUserId(ctx, memberOwnerId)
	if err != nil {
		return err
	}

	return helpers.GetTx(ctx).Transaction(func(tx *gorm.DB) error {
		return errors.Wrapf(
			tx.Model(&org).Association("Members").Append(user),
			"failed to add member to organization",
		)
	})
}

func (s *service) UpdateGithubInstallationId(ctx context.Context, installationId int64) error {
	user, err := s.GetUser(ctx)
	if err != nil {
		return err
	}

	user.GithubInstallationId = installationId
	if err := user.Save(helpers.GetTx(ctx)); err != nil {
		return errors.Wrapf(err, "failed to save user")
	}

	return nil
}

func (s *service) UpdateGithubAccessToken(ctx context.Context, accessToken string) error {
	user, err := s.GetUser(ctx)
	if err != nil {
		return err
	}

	user.GithubAccessToken = accessToken
	if err := user.Save(helpers.GetTx(ctx)); err != nil {
		return errors.Wrapf(err, "failed to save user")
	}

	return nil
}
