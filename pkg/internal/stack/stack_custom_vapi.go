package stack

import (
	"bytes"
	"context"
	"fmt"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"regexp"
)

type (
	CreateCustomVapiInput struct {
		Name string
	}

	UpdateCustomVapiInput struct {
		NewName       *string
		UpdateTarFile *bool
	}
)

func (s *service) EnableCustomVapi(
	ctx context.Context,
	stackId uint,
	input CreateCustomVapiInput,
) (*domain.CustomVapi, error) {
	tx := helpers.GetTx(ctx)

	stack, err := s.GetStack(ctx, stackId)
	if err != nil {
		return nil, err
	}

	if matched, err := regexp.MatchString("^[a-z0-9_-]+$", input.Name); err != nil {
		return nil, errors.Wrapf(err, "failed to regex")
	} else if !matched {
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "Invalid name")
	} else if err := stack.ValidateVapiNameUniqueness(tx, input.Name); err != nil {
		return nil, err
	}

	tarFilePath, err := s.createVapiTarFile(ctx, stack, input.Name)
	if err != nil {
		return nil, err
	}

	customVapi := domain.CustomVapi{
		StackID:     stackId,
		Name:        input.Name,
		TarFilePath: tarFilePath,
	}

	if err := customVapi.Save(tx); err != nil {
		return nil, err
	}

	return &customVapi, nil
}

func (s *service) UpdateCustomVapi(
	ctx context.Context,
	stackId uint,
	name string,
	input UpdateCustomVapiInput,
) error {
	tx := helpers.GetTx(ctx)

	customVapi, err := s.GetCustomVapiByName(ctx, stackId, name)
	if err != nil {
		return err
	}

	isChanged := false
	if input.NewName != nil {
		name := *input.NewName
		if matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", name); err != nil {
			return errors.Wrapf(err, "failed to regex")
		} else if !matched {
			return errors.Wrapf(tclerrors.ErrBadRequest, "Invalid name")
		} else if err := customVapi.Stack.ValidateVapiNameUniqueness(tx, name); err != nil {
			return err
		}
		customVapi.Name = name
		isChanged = true
	}

	if input.UpdateTarFile != nil && *input.UpdateTarFile {
		tarFilePath, err := s.createVapiTarFile(ctx, &customVapi.Stack, customVapi.Name)
		if err != nil {
			return err
		}

		customVapi.TarFilePath = tarFilePath
		isChanged = true
	}

	if !isChanged {
		return nil
	}

	return customVapi.Save(tx)
}

func (s *service) DisableCustomVapi(ctx context.Context, stackId uint, name string) error {
	tx := helpers.GetTx(ctx)

	customVapi, err := s.GetCustomVapiByName(ctx, stackId, name)
	if err != nil {
		return err
	}

	return customVapi.Delete(tx)
}

func (s *service) GetCustomVapiByName(
	ctx context.Context,
	stackId uint,
	name string,
) (*domain.CustomVapi, error) {
	st, err := s.GetStack(ctx, stackId)
	if err != nil {
		return nil, err
	}

	var customVapi domain.CustomVapi
	if r := helpers.GetTx(ctx).
		First(&customVapi, "stack_id = ? AND name = ?", st.ID, name); r.Error != nil {
		return nil, errors.Wrapf(r.Error, "not found custom vapi")
	}

	customVapi.Stack = *st
	return &customVapi, nil
}

func (s *service) GetCustomVapis(
	ctx context.Context,
	stackId uint,
) (customVapis []domain.CustomVapi, err error) {
	st, err := s.GetStack(ctx, stackId)
	if err != nil {
		return nil, err
	}

	tx := helpers.GetTx(ctx)
	if r := tx.
		Where("stack_id = ?", st.ID).
		Find(&customVapis); r.Error != nil {
		return nil, errors.Wrapf(r.Error, "failed to get custom vapis")
	}

	return customVapis, nil
}

func (s *service) createVapiTarFile(ctx context.Context, stack *domain.Stack, name string) (string, error) {
	user, err := s.users.GetUser(ctx)
	if err != nil {
		return "", err
	}

	accessToken, err := user.GetGithubAccessToken(ctx)
	if err != nil {
		return "", err
	}

	gitUrl := fmt.Sprintf("https://%s@github.com/%s", accessToken, stack.GitRepo)

	repo, err := s.git.Clone(ctx, gitUrl, stack.GitBranch)
	if err != nil {
		return "", err
	}

	tree, err := s.git.OpenDir(repo, fmt.Sprintf("/vapis/%s", name))
	if err != nil {
		return "", err
	}

	var tarFileBuf bytes.Buffer
	if err := util.ArchiveGitTree(tree, &tarFileBuf); err != nil {
		return "", err
	}

	objectPath := fmt.Sprintf("%d/%s.tar", stack.ID, name)
	if _, err := s.storageClient.UploadFile(ctx, constants.CustomVapiBucketId, objectPath, &tarFileBuf, storage.FileOptions{
		Upsert: gog.PtrOf(true),
	}); err != nil {
		return "", errors.Wrapf(err, "failed to upload custom vapi tar file")
	}

	return objectPath, nil
}
