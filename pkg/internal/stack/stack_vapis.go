package stack

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type EnableVapiInput struct {
	VapiID uint `json:"vapi_id"`
}

type UpdateVapiInput struct {
	Version string `json:"version"`
}

func (i EnableVapiInput) Validate() error {
	if i.VapiID == 0 {
		return errors.Wrapf(tclerrors.ErrBadRequest, "vapi_id is required")
	}

	return nil
}

func (ss *service) DisableVapi(
	ctx context.Context,
	stackId uint,
	vapiId uint,
) error {
	stack, err := ss.GetStack(ctx, stackId)
	if err != nil {
		return err
	}

	if err := ss.hasPermission(ctx, stack.ProjectID); err != nil {
		return err
	}

	tx := helpers.GetTx(ctx)
	stackVapi, err := domain.GetStackVapiByStackIDAndVapiID(tx, stackId, vapiId)
	if err != nil {
		return err
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		return errors.Wrapf(stackVapi.Delete(tx), "failed to delete stack vapi")
	})
}

func (ss *service) EnableVapi(
	ctx context.Context,
	stackId uint,
	input EnableVapiInput,
) (*domain.StackVapi, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	stack, err := ss.GetStack(ctx, stackId)
	if err != nil {
		return nil, err
	}

	tx := helpers.GetTx(ctx)

	vapiRelease, err := domain.GetVapiReleaseByID(tx, input.VapiID)
	if err != nil {
		return nil, err
	}

	if err := stack.ValidateVapiNameUniqueness(tx, vapiRelease.Package.Name); err != nil {
		return nil, err
	}

	stackVapi := domain.StackVapi{
		StackID: stackId,
		VapiID:  input.VapiID,
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		return errors.Wrapf(stackVapi.Create(tx), "failed to create stack vapi")
	}); err != nil {
		return nil, err
	}

	return &stackVapi, nil
}

func (ss *service) UpdateVapi(
	ctx context.Context,
	stackId uint,
	vapiId uint,
	input UpdateVapiInput,
) (*domain.StackVapi, error) {
	if _, err := ss.GetStack(ctx, stackId); err != nil {
		return nil, err
	}

	tx := helpers.GetTx(ctx)

	stackVapi, err := domain.GetStackVapiByStackIDAndVapiID(tx, stackId, vapiId)
	if err != nil {
		return nil, err
	}

	st := &stackVapi.Stack
	if err := tx.Model(st).Association("Vapis").Find(&st.Vapis); err != nil {
		return nil, errors.Wrapf(err, "failed to find stack vapis")
	}

	vapiRelease, err := domain.FindVapiReleaseByPackageIDAndVersion(
		tx,
		stackVapi.Vapi.PackageID,
		input.Version,
	)
	if err != nil {
		return nil, err
	}

	if err := st.ValidateVapiNameUniqueness(tx, vapiRelease.Package.Name); err != nil {
		return nil, err
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(stackVapi).Association("Vapi").Replace(vapiRelease); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *service) SetVapiEnv(
	ctx context.Context,
	stackId uint,
	env map[string]string,
) error {
	stack, err := s.GetStack(ctx, stackId)
	if err != nil {
		return err
	}

	envVars := stack.GetVapiEnvVarsMap()
	for key, value := range env {
		envVars[key] = value
	}
	stack.SetVapiEnvVarsMap(envVars)

	return helpers.GetTx(ctx).Transaction(func(tx *gorm.DB) error {
		return stack.Save(tx)
	})
}

func (s *service) UnsetVapiEnv(ctx context.Context, stackId uint, names []string) error {
	stack, err := s.GetStack(ctx, stackId)
	if err != nil {
		return err
	}

	envVars := stack.GetVapiEnvVarsMap()
	for _, name := range names {
		if _, ok := envVars[name]; !ok {
			continue
		}
		delete(envVars, name)
	}
	stack.SetVapiEnvVarsMap(envVars)

	return helpers.GetTx(ctx).Transaction(func(tx *gorm.DB) error {
		return stack.Save(tx)
	})
}
