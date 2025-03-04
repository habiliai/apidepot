package svctpl

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/habiliai/apidepot/pkg/internal/util/functx/v2"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

var (
	dotEnvTpl = template.Must(template.New("dotEnv").Parse(strings.TrimSpace(`
NEXT_PUBLIC_APIDEPOT_URL={{ .Endpoint }}
NEXT_PUBLIC_APIDEPOT_ANON_KEY={{ .AnonApiKey }}
APIDEPOT_URL={{ .Endpoint }}
APIDEPOT_ANON_KEY={{ .AnonApiKey }}
`)))
)

func (s *serviceImpl) CreateStackFromServiceTemplate(
	ctx context.Context,
	serviceTemplateId uint,
	input stack.CreateStackInput,
) (*domain.Stack, error) {
	user, err := s.users.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	installationId := user.GithubInstallationId
	if installationId == 0 {
		return nil, errors.Wrapf(tclerrors.ErrPreconditionRequired, "installationId is empty")
	}

	accessToken, err := s.githubClient.GenerateInstallationAccessToken(ctx, installationId)
	if err != nil {
		return nil, err
	}

	tx := helpers.GetTx(ctx)
	ctx, fDone := functx.WithFuncTx(ctx)
	defer fDone(ctx, true)

	serviceTpl, err := domain.FindServiceTemplateByID(tx, serviceTemplateId)
	if err != nil {
		return nil, err
	}

	// Create a new stack
	st, err := s.stacks.CreateStack(ctx, input)
	if err != nil {
		return nil, err
	}
	functx.AddRollback(ctx, func(ctx context.Context) {
		if err := s.stacks.DeleteStack(ctx, st.ID); err != nil {
			logger.Error("failed to delete stack", "stack_id", st.ID, "error", err)
		}
	})

	// Install Base APIS: auth, storage and postgrest
	if err := s.stacks.EnableOrUpdateAuth(
		ctx,
		st.ID,
		stack.EnableOrUpdateAuthInput{
			AuthInput: stack.AuthInput{
				MailerAutoConfirm: gog.PtrOf(true),
				EmailEnabled:      gog.PtrOf(true),
			},
		}, true,
	); err != nil {
		return nil, err
	}
	if err := s.stacks.EnableOrUpdateStorage(
		ctx,
		st.ID,
		stack.EnableOrUpdateStorageInput{
			StorageInput: stack.StorageInput{
				TenantID: &input.Name,
			},
		},
		true,
	); err != nil {
		return nil, err
	}
	if err := s.stacks.EnableOrUpdatePostgrest(
		ctx,
		st.ID,
		stack.EnableOrUpdatePostgrestInput{
			PostgrestInput: stack.PostgrestInput{
				Schemas: []string{"public"},
			},
		},
		true,
	); err != nil {
		return nil, err
	}

	// Install service template's VAPIs to the stack
	for _, vapiId := range serviceTpl.VapiIds {
		if _, err := s.stacks.EnableVapi(ctx, st.ID, stack.EnableVapiInput{
			VapiID: vapiId,
		}); err != nil {
			return nil, err
		}
		logger.Info("installed VAPI for stack", "vapi_id", vapiId, "stack_id", st.ID)
	}

	// Clone the repository
	dstGitUrl := fmt.Sprintf("https://%s@github.com/%s", accessToken, input.GitRepo)
	repo, err := s.gitService.Clone(ctx, dstGitUrl, input.GitBranch, services.GitCloneOptionsAllowCommit())
	if err != nil {
		return nil, err
	}

	conf, err := repo.Config()
	if err != nil {
		return nil, err
	}

	conf.User.Name = "JDL Support"
	conf.User.Email = "support@joat.land"
	if err := repo.SetConfig(conf); err != nil {
		return nil, err
	}

	st, err = s.stacks.GetStack(ctx, st.ID)
	if err != nil {
		return nil, err
	}

	// Fill .env file contents with stack information
	var bb bytes.Buffer
	if dotEnvBody, err := s.gitService.ReadFile(repo, ".env.tmpl"); err == nil {
		tpl, err := template.New("").Parse(string(dotEnvBody))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse .env file")
		}
		if err := tpl.Execute(&bb, st); err != nil {
			return nil, errors.Wrapf(err, "failed to execute .env template")
		}

	} else if errors.Is(err, object.ErrFileNotFound) {
		if err := dotEnvTpl.Execute(&bb, st); err != nil {
			return nil, errors.Wrapf(err, "failed to execute .env template")
		}
	} else {
		return nil, err
	}

	// Commit the .env file
	if err := s.gitService.CommitFile(repo, ".env", bb.Bytes(), "chore: add .env file"); err != nil {
		return nil, err
	}

	// Push the changes
	pushOption := &git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "git",
			Password: accessToken,
		},
	}
	if err := repo.PushContext(ctx, pushOption); err != nil {
		return nil, errors.Wrapf(err, "failed to push changes")
	}

	st.ServiceTemplateID = &serviceTemplateId
	st.ServiceTemplate = serviceTpl
	if err := st.Save(tx); err != nil {
		return nil, err
	}

	fDone(ctx, false)

	return st, nil
}
