package stack

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
)

func (ss *service) GetStacks(
	ctx context.Context,
	projectId uint,
	name *string,
	cursor uint,
	limit int,
) ([]domain.Stack, error) {
	tx := helpers.GetTx(ctx)

	project, err := domain.FindProjectById(tx, projectId)
	if err != nil {
		return nil, err
	}

	if user, err := ss.users.GetUser(ctx); err != nil {
		return nil, err
	} else if project.OwnerID != user.ID && !user.IsSuperuser() {
		return nil, errors.Wrapf(tclerrors.ErrForbidden, "you are not allowed to access this project")
	}

	tx = tx.Where("project_id = ? and id > ?", project.ID, cursor)
	if name != nil {
		tx = tx.Where("name = ?", *name)
	}
	if limit == 0 {
		limit = constants.DefaultPageLimit
	}
	stacks, err := domain.FindStacks(tx.Limit(limit))

	if err != nil {
		return nil, errors.Wrapf(err, "failed to find stacks")
	}

	return stacks, nil
}

func (ss *service) GetStack(ctx context.Context, id uint) (*domain.Stack, error) {
	tx := helpers.GetTx(ctx)
	st, err := domain.FindStackByID(
		tx.Preload("Project"),
		id,
	)
	if err != nil {
		return nil, err
	}

	if user, err := ss.users.GetUser(ctx); err != nil {
		return nil, err
	} else if st.Project.OwnerID != user.ID && !user.IsSuperuser() {
		return nil, errors.Wrapf(tclerrors.ErrForbidden, "you are not allowed to access this stack")
	}

	return st, nil
}
