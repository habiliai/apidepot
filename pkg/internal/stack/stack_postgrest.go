package stack

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

type PostgrestInput struct {
	Schemas []string `json:"schemas"`
}

type EnableOrUpdatePostgrestInput struct {
	PostgrestInput
}

func (ss *service) EnableOrUpdatePostgrest(
	ctx context.Context,
	stackId uint,
	input EnableOrUpdatePostgrestInput,
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

	if stack.PostgrestEnabled && isCreate {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "postgrest is already created")
	} else if !stack.PostgrestEnabled && !isCreate {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "postgrest is not enabled")
	}

	postgrest := stack.Postgrest.Data()
	if isCreate {
		postgrest = domain.Postgrest{
			Schemas: []string{
				"public",
				"api",
			},
		}
	}

	if len(input.Schemas) > 0 {
		postgrest.Schemas = input.Schemas
	}

	stack.PostgrestEnabled = true
	stack.Postgrest = datatypes.NewJSONType(postgrest)
	if err := tx.Transaction(func(tx *gorm.DB) error {
		if err := stack.Save(tx.Omit(clause.Associations)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (ss *service) DisablePostgrest(
	ctx context.Context,
	stackId uint,
) error {
	tx := helpers.GetTx(ctx)
	stack, err := ss.GetStack(ctx, stackId)
	if err != nil {
		return err
	}

	if !stack.PostgrestEnabled {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "postgrest is not enabled")
	}

	stack.PostgrestEnabled = false
	return tx.Transaction(func(tx *gorm.DB) error {
		if err := stack.Save(tx.Omit(clause.Associations)); err != nil {
			return err
		}

		return nil
	})
}

func (ss *service) waitPostgrestForReady(ctx context.Context, stack *domain.Stack, timeout time.Duration) error {
	ctx, cancel := context.WithTimeoutCause(ctx, timeout, tclerrors.ErrTimeout)
	defer cancel()
	if !stack.PostgrestEnabled {
		return nil
	}

	for ready := false; !ready; {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				if errors.Is(err, tclerrors.ErrTimeout) {
					return errors.Wrapf(err, "timeout waiting for postgrest to be ready")
				}
				logger.Warn("context cancelled", "err", err)
			}
			return nil
		case <-time.After(500 * time.Millisecond):
			// ready-path is checked end of reloading schema
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, stack.ServicePath(constants.PathPostgrestReady), nil)
			if err != nil {
				return errors.Wrapf(err, "failed to create request")
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errors.Wrapf(err, "failed to do request")
			}

			ready = resp.StatusCode == http.StatusOK
		}
	}

	return nil
}
