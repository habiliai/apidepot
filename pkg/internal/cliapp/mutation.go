package cliapp

import (
	"context"
	"crypto"
	"encoding/base64"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type (
	AddCliAppOutput struct {
		AppId     string
		AppSecret string
	}
)

func (s *service) RegisterCliApp(
	ctx context.Context,
	host string,
	refreshToken string,
) (*AddCliAppOutput, error) {
	user, err := s.users.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	tx := helpers.GetTx(ctx)

	appIdBin, err := util.GenerateRandomHash(crypto.SHA1)
	if err != nil {
		return nil, err
	}
	appId := base64.StdEncoding.EncodeToString(appIdBin)
	appSecretBin, err := util.GenerateRandomHash(crypto.SHA1)
	if err != nil {
		return nil, err
	}
	appSecret := base64.StdEncoding.EncodeToString(appSecretBin)
	encAppSecret, err := bcrypt.GenerateFromPassword(appSecretBin, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	cliApp := domain.CliApp{
		Host:         host,
		RefreshToken: refreshToken,

		AppId:     appId,
		AppSecret: encAppSecret,

		Owner:   *user,
		OwnerID: user.ID,
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		return cliApp.Save(tx)
	}); err != nil {
		return nil, err
	}

	return &AddCliAppOutput{
		AppId:     appId,
		AppSecret: appSecret,
	}, nil
}

func (s *service) DeleteCliApp(
	ctx context.Context,
	appId string,
) error {
	user, err := s.users.GetUser(ctx)
	if err != nil {
		return err
	}

	tx := helpers.GetTx(ctx)

	cliApp, err := domain.GetCliAppByAppId(tx, appId)
	if err != nil {
		return err
	}

	if cliApp.OwnerID != user.ID {
		return errors.Wrapf(tclerrors.ErrForbidden, "cli app not owned by user")
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		if err := cliApp.Delete(tx); err != nil {

			return err
		}
		return nil
	})
}
