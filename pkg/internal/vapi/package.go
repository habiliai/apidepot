package vapi

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type (
	GetPackagesInput struct {
		Name *string `json:"name" form:"name"`
	}
)

func (s *service) GetPackage(
	ctx context.Context,
	id uint,
) (*domain.VapiPackage, error) {
	tx := helpers.GetTx(ctx)
	pkg, err := domain.FindVapiPackageByID(tx, id)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

func (s *service) GetPackages(
	ctx context.Context,
	input GetPackagesInput,
) ([]domain.VapiPackage, error) {
	stmt := helpers.GetTx(ctx)

	if input.Name != nil {
		stmt = stmt.Where("name = ?", *input.Name)
	}

	var pkgs []domain.VapiPackage
	if err := stmt.Find(&pkgs).Error; err != nil {
		return nil, err
	}

	return pkgs, nil
}

func (s *service) DeletePackage(
	ctx context.Context,
	id uint,
) error {
	tx := helpers.GetTx(ctx)

	pkg, err := domain.FindVapiPackageByID(tx, id)
	if err != nil {
		return err
	}

	me, err := s.users.GetUser(ctx)
	if err != nil {
		return err
	}

	if err := pkg.IsPermittedToEdit(me); err != nil {
		return err
	}

	if len(pkg.Releases) > 0 {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "package is able to delete only if it has no releases")
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		return pkg.Delete(tx)
	})
}

func (s *service) DeleteAllPackages(
	ctx context.Context,
	projectId uint,
) error {
	tx := helpers.GetTx(ctx)

	var (
		pkgs []domain.VapiPackage
		err  error
	)
	if projectId > 0 {
		pkgs, err = domain.FindVapiPackagesByProjectID(tx, projectId)
	} else {
		pkgs, err = domain.FindVapiPackages(tx)
	}
	if err != nil {
		return err
	}

	me, err := s.users.GetUser(ctx)
	if err != nil {
		return err
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		for _, pkg := range pkgs {
			if err := pkg.IsPermittedToEdit(me); err != nil {
				return err
			}

			if len(pkg.Releases) > 0 {
				return errors.Wrapf(tclerrors.ErrPreconditionRequired, "package is able to delete only if it has no releases")
			}

			if err := pkg.Delete(tx); err != nil {
				return err
			}
		}

		return nil
	})
}
