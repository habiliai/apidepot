package stack

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
)

func (s *service) hasPermission(
	ctx context.Context,
	projectId uint,
) error {
	tx := helpers.GetTx(ctx)
	if prj, err := domain.FindProjectById(tx, projectId); err != nil {
		return err
	} else if user, err := s.users.GetUser(ctx); err != nil {
		return err
	} else if user.ID != prj.OwnerID && !user.IsSuperuser() {
		return errors.Wrapf(tclerrors.ErrForbidden, "you are not allowed to access this project")
	} else {
		logger.Debug("edit stack by user", "role", user.Role, "id", user.ID)
	}

	return nil
}
