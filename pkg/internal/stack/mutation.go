package stack

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/Masterminds/goutils"
	"github.com/google/uuid"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/martinlindhe/base36"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"regexp"
	"strings"
)

//go:embed data/init.sql
var initSQL string

func (ss *service) CreateStack(
	ctx context.Context,
	input CreateStackInput,
) (*domain.Stack, error) {
	{ // validate input
		if input.ProjectID == 0 {
			return nil, errors.Wrapf(tclerrors.ErrBadRequest, "project_id is required")
		}

		if input.Name == "" {
			return nil, errors.Wrapf(tclerrors.ErrBadRequest, "name is required")
		} else if nameLen := len(input.Name); nameLen < 3 || nameLen > 50 {
			logger.Debug("print", "name", input.Name)
			return nil, errors.Wrapf(tclerrors.ErrBadRequest, "name must be between 3 and 50 characters")
		}

		if matched, err := regexp.MatchString("^[a-zA-Z0-9-_ ]+$", input.Name); err != nil {
			return nil, errors.Wrapf(err, "failed to validate name")
		} else if !matched {
			return nil, errors.Wrapf(tclerrors.ErrBadRequest, "name must contain only alphanumeric characters and hyphens")
		}

		if input.SiteURL == "" {
			return nil, errors.Wrapf(tclerrors.ErrBadRequest, "site_url is required")
		}

		if input.DefaultRegion == "" {
			input.DefaultRegion = tcltypes.InstanceZoneDefault
		}
	}

	tx := helpers.GetTx(ctx)
	regionalStackConfig := ss.stackConfig.GetRegionalConfig(input.DefaultRegion)

	if err := ss.hasPermission(ctx, input.ProjectID); err != nil {
		return nil, err
	}

	dbPassword, err := goutils.CryptoRandomAlphaNumeric(32)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate random db password")
	}

	hash, err := getRandomName()
	dbName := "db_" + hash
	dbUsername := "user_" + hash
	domainName := hash + "." + regionalStackConfig.Domain

	stack := &domain.Stack{
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Hash:      hash,
		Domain:    domainName,
		Scheme:    regionalStackConfig.Scheme,
		SiteURL:   input.SiteURL,
		DB: datatypes.NewJSONType(domain.DB{
			Name:     dbName,
			Username: dbUsername,
			Password: dbPassword,
		}),
		Description:   input.Description,
		LogoImageUrl:  input.LogoImageUrl,
		DefaultRegion: input.DefaultRegion,
		GitRepo:       input.GitRepo,
		GitBranch:     input.GitBranch,
	}

	if stack.SiteURL == "" {
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "site url is empty")
	}

	if err := ss.runtimeSchema.CreateUserAndDB(ctx, stack.DefaultRegion, stack.DB.Data().Username, stack.DB.Data().Password, stack.DB.Data().Name); err != nil {
		return nil, err
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		if err = stack.Save(tx.Preload("Project")); err != nil {
			return err
		}

		if err = ss.applySchema(ctx, stack); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return stack, nil
}

func (ss *service) DeleteStack(ctx context.Context, id uint) error {
	tx := helpers.GetTx(ctx)
	stack, err := domain.FindStackByID(
		tx.
			Preload("Instances"),
		id,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to find stack by id")
	}

	if user, err := ss.users.GetUser(ctx); err != nil {
		return err
	} else if stack.Project.OwnerID != user.ID && !user.IsSuperuser() {
		return errors.Wrapf(tclerrors.ErrForbidden, "you are not allowed to access this stack")
	}

	logger.Debug("print", "numInstances", len(stack.Instances))
	if len(stack.Instances) > 0 {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "stack has instances")
	}

	if stack.PostgrestEnabled {
		if err := ss.DisablePostgrest(ctx, stack.ID); err != nil {
			return err
		}
	}

	if stack.StorageEnabled {
		if err := ss.DisableStorage(ctx, stack.ID); err != nil {
			return err
		}
	}

	if stack.AuthEnabled {
		if err := ss.DisableAuth(ctx, stack.ID); err != nil {
			return err
		}
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		if err := stack.Delete(tx); err != nil {
			return errors.Wrapf(err, "failed to delete stack")
		}

		return nil
	}); err != nil {
		return err
	}

	if err := ss.runtimeSchema.DropUserAndDB(ctx, stack.DefaultRegion, stack.DB.Data().Username, stack.DB.Data().Name); err != nil {
		return err
	}

	return nil
}

func (ss *service) PatchStack(
	ctx context.Context,
	id uint,
	input PatchStackInput,
) error {
	tx := helpers.GetTx(ctx)
	stack, err := domain.FindStackByID(tx.Preload("Project"), id)
	if err != nil {
		return errors.Wrapf(err, "failed to find stack by id")
	}

	if err := ss.hasPermission(ctx, stack.ProjectID); err != nil {
		return err
	}

	if input.SiteURL != nil {
		stack.SiteURL = *input.SiteURL
	}

	if input.Name != nil {
		stack.Name = *input.Name
	}

	if input.Description != nil {
		stack.Description = *input.Description
	}

	if input.LogoImageUrl != nil {
		stack.LogoImageUrl = *input.LogoImageUrl
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		if err := stack.Save(tx.Omit(clause.Associations)); err != nil {
			return errors.Wrapf(err, "failed to save stack")
		}

		return nil
	})
}

func (ss *service) applySchema(ctx context.Context, stack *domain.Stack) error {
	regionalDbConfig := ss.dbConfig.GetRegionalConfig(stack.DefaultRegion)
	stackDB := stack.DB.Data()
	conn, err := pgx.Connect(
		ctx,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			stackDB.Username,
			stackDB.Password,
			regionalDbConfig.Host,
			regionalDbConfig.Port,
			stackDB.Name,
		),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to db")
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, initSQL); err != nil {
		return errors.Wrapf(err, "failed to execute sql")
	}

	return nil
}

func getRandomName() (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate uuid")
	}
	return strings.ToLower(base36.EncodeBytes(uid[:])), nil
}
