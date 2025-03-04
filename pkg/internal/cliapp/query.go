package cliapp

import (
	"context"
	"encoding/base64"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type (
	VerifyCliAppResponse struct {
		AccessToken string
	}
)

func (s *service) VerifyCliApp(
	ctx context.Context,
	appId string,
	appSecret string,
) (*VerifyCliAppResponse, error) {
	tx := helpers.GetTx(ctx)

	cliApp, err := domain.GetCliAppByAppId(tx, appId)
	if err != nil {
		return nil, err
	}

	appSecretBin, err := base64.StdEncoding.DecodeString(appSecret)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := bcrypt.CompareHashAndPassword(cliApp.AppSecret, appSecretBin); err != nil {
		return nil, errors.Wrapf(tclerrors.ErrUnauthorized, "invalid app secret")
	}

	session, err := s.gotrueClient.RefreshToken(cliApp.RefreshToken)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		cliApp.RefreshToken = session.RefreshToken
		return cliApp.Save(tx)
	}); err != nil {
		return nil, err
	}

	return &VerifyCliAppResponse{
		AccessToken: session.AccessToken,
	}, nil
}
