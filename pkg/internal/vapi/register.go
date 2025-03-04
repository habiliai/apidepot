package vapi

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
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type (
	DeployStatus string

	VapiPackageYaml struct {
		Name         string         `yaml:"name"`
		Version      string         `yaml:"version"`
		Dependencies map[string]any `yaml:"dependencies"`
	}
)

func (s *service) Register(
	ctx context.Context,
	gitRepo string,
	gitBranch string,
	name string,
	description string,
	domains []string,
	vapiPoolId string,
	homepage string,
) (*domain.VapiRelease, error) {
	tx := helpers.GetTx(ctx)

	user, err := s.users.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	installationId := user.GithubInstallationId
	if installationId == 0 {
		return nil, errors.New("installationId is empty")
	}

	// validate input parameters
	if gitRepo == "" {
		return nil, errors.New("gitRepo is empty. must be like 'go-git/go-git'")
	}

	if !strings.Contains(gitRepo, "/") {
		return nil, errors.New("gitRepo is invalid. must be like 'go-git/go-git'")
	}

	gitRepoSplitted := strings.Split(gitRepo, "/")
	if len(gitRepoSplitted) != 2 {
		return nil, errors.New("gitRepo is invalid. must be like 'go-git/go-git'")
	}

	if gitRepoSplitted[0] == "" || gitRepoSplitted[1] == "" {
		return nil, errors.Errorf("gitRepo('%s') is invalid. must be like 'go-git/go-git'", gitRepo)
	}

	accessToken := user.GithubAccessToken
	if accessToken == "" {
		accessToken = helpers.GetGithubToken(ctx)
	}

	var gitUrl string
	if accessToken == "" {
		gitUrl = fmt.Sprintf("https://github.com/%s", gitRepo)
	} else {
		gitUrl = fmt.Sprintf("https://%s@github.com/%s", accessToken, gitRepo)
	}
	logger.Debug("check", "gitUrl", gitUrl, "gitBranch", gitBranch)

	repo, err := s.git.Clone(ctx, gitUrl, gitBranch)
	if err != nil {
		return nil, err
	}

	// get VapiPackageYaml struct from apidepot.yml using Git info.
	head, err := repo.Head()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get head")
	}

	gitHash := head.Hash().String()

	var vapiPackageYaml VapiPackageYaml
	if file, err := s.git.ReadFile(repo, constants.VapiYamlFileName); err != nil {
		return nil, errors.Wrapf(err, "failed to read file: %s", constants.VapiYamlFileName)
	} else if err := yaml.Unmarshal(file, &vapiPackageYaml); err != nil {
		return nil, err
	} else if vapiPackageYaml.Version == "" {
		return nil, errors.New("package.json is invalid. version is empty")
	}

	// make tar file from git repo
	dirTree, err := s.git.OpenDir(repo, "")
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to open dir repo='%s'", gitRepo)
	}

	var tarFile bytes.Buffer
	if err := util.ArchiveGitTree(dirTree, &tarFile); err != nil {
		return nil, err
	}

	// make up vapiRelease and dependencies, then upload tar file to storage
	var rel *domain.VapiRelease
	tarObjectPath := fmt.Sprintf("%s/v%s.tar", name, vapiPackageYaml.Version)
	if err := tx.Transaction(func(tx *gorm.DB) error {
		var vapiPackage domain.VapiPackage
		if r := tx.Find(&vapiPackage, "name = ?", name); r.Error != nil {
			return errors.Wrapf(r.Error, "failed to find vapi by name")
		} else {
			vapiPackage.Description = description
			vapiPackage.Domains = domains
			vapiPackage.GitRepo = gitRepo
			vapiPackage.GitBranch = gitBranch
			vapiPackage.Homepage = homepage
			if r.RowsAffected == 0 {
				logger.Debug("not found vapi", "pkgName", name)
				vapiPackage.Name = name
				vapiPackage.OwnerId = user.ID
				vapiPackage.Owner = *user
				vapiPackage.VapiPoolId = vapiPoolId
			}
			if err := vapiPackage.Save(tx); err != nil {
				return err
			}
		}

		if err := vapiPackage.IsPermittedToEdit(user); err != nil {
			return errors.Wrapf(tclerrors.ErrForbidden, "failed to edit package")
		}

		var skip bool
		rel, skip, err = s.createVapiRelease(tx, vapiPackage, tarObjectPath, gitHash, vapiPackageYaml.Version)
		if err != nil {
			return err
		}

		if skip {
			return nil
		}

		dependencies, err := ParseDependencies(vapiPackageYaml.Dependencies)
		if err != nil {
			return err
		}
		for _, dep := range dependencies {
			dep, err := domain.FindVapiReleaseByPackageNameAndVersion(tx, dep.Name, dep.Version)
			if err != nil {
				return err
			}

			if !dep.Published {
				return errors.Errorf("dependency vapi '%s' is not used", dep.Package.Name)
			}

			if err := tx.Model(rel).Association("Dependencies").Append(dep); err != nil {
				return errors.Wrapf(err, "failed to append dependency")
			}
		}

		{
			// upload tar file to storage during 30 seconds with timeout
			ctx, cancel := context.WithTimeoutCause(ctx, 30*time.Second, tclerrors.ErrTimeout)
			defer cancel()

			if _, err := s.storage.UploadFile(
				ctx,
				constants.VapiBucketId,
				tarObjectPath,
				&tarFile,
				storage.FileOptions{
					Upsert: gog.PtrOf(true),
				},
			); err != nil {
				return errors.Wrapf(err, "failed to upload file")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return rel, nil
}

func (s *service) createVapiRelease(
	tx *gorm.DB,
	vapi domain.VapiPackage,
	tarPath string,
	gitHash string,
	version string,
) (*domain.VapiRelease, bool, error) {
	release := domain.VapiRelease{
		Version:     version,
		TarFilePath: tarPath,
		Deprecated:  false,
		Suspended:   false,
		Published:   true, // TODO: it is needed to be changed by auto-preview process
		PackageID:   vapi.ID,
		Package:     vapi,
		GitHash:     gitHash,
	}

	if r := tx.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&release); r.Error != nil {
		return nil, false, errors.Wrapf(r.Error, "failed to create vapi release")
	} else if r.RowsAffected == 0 {
		rel, err := domain.FindVapiReleaseByPackageIDAndVersion(tx, vapi.ID, version)
		if err != nil {
			return nil, false, err
		}

		return rel, true, nil
	}

	logger.Info("created vapi release", "name", vapi.Name, "version", release.Version)

	return &release, false, nil
}
