package vapi

import (
	"archive/tar"
	"bytes"
	"context"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Migration struct {
	Version time.Time `json:"version"` // format: yymmddHHMMSS
	Query   string    `json:"query"`
}

func (s *service) GetDBMigrations(
	ctx context.Context,
	vapiRel domain.VapiRelease,
) ([]Migration, error) {
	tarFile, err := s.storage.DownloadFile(ctx, constants.VapiBucketId, vapiRel.TarFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download tar file")
	}

	tfs := tarfs.New(tar.NewReader(bytes.NewBuffer(tarFile)))
	if ok, err := afero.DirExists(tfs, "/migrations"); err != nil {
		return nil, errors.Wrapf(err, "failed to check migrations directory")
	} else if !ok {
		return nil, nil
	}

	files, err := afero.ReadDir(tfs, "/migrations")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read migrations directory")
	}

	migrations := make([]Migration, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		fileBaseName, ok := strings.CutSuffix(file.Name(), ".sql")
		if !ok {
			continue
		}

		logger.Debug("check", "fileBaseName", fileBaseName)
		versionStr, _, _ := strings.Cut(fileBaseName, "_")

		logger.Debug("extracted", "versionStr", versionStr, "from", fileBaseName)
		version, err := time.Parse("060102150405", versionStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse migration version")
		}

		migration, err := afero.ReadFile(tfs, "/migrations/"+file.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read migration file")
		}

		migrations = append(migrations, Migration{
			Version: version,
			Query:   string(migration),
		})
	}

	return migrations, nil
}

func (s *service) GetRelease(
	ctx context.Context,
	id uint,
) (*domain.VapiRelease, error) {
	return domain.GetVapiReleaseByID(helpers.GetTx(ctx), id)
}

func (s *service) GetReleaseByVersionInPackage(
	ctx context.Context,
	packageId uint,
	version string,
) (*domain.VapiRelease, error) {
	if version == "" {
		version = "latest"
	}

	tx := helpers.GetTx(ctx)
	if version != "latest" {
		return domain.FindVapiReleaseByPackageIDAndVersion(tx, packageId, version)
	} else {
		return domain.FindLatestVapiReleaseByPackageID(tx, packageId)
	}
}

func (s *service) DeleteRelease(
	ctx context.Context,
	id uint,
) error {
	tx := helpers.GetTx(ctx)
	rel, err := domain.GetVapiReleaseByID(tx, id)
	if err != nil {
		return err
	}

	me, err := s.users.GetUser(ctx)
	if err != nil {
		return err
	}

	if err := rel.Package.IsPermittedToEdit(me); err != nil {
		return err
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		if err := rel.Delete(tx); err != nil {
			return err
		}

		_, err := s.storage.RemoveFile(ctx, constants.VapiBucketId, []string{
			rel.TarFilePath,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to remove tar file")
		}

		return nil
	})
}

func (s *service) deleteReleases(
	ctx context.Context,
	tx *gorm.DB,
	rels []domain.VapiRelease,
) error {
	return tx.Transaction(func(tx *gorm.DB) error {
		tarFilePaths := make([]string, 0, len(rels))
		for _, rel := range rels {
			if err := rel.Delete(tx); err != nil {
				return err
			}

			tarFilePaths = append(tarFilePaths, rel.TarFilePath)
		}

		if _, err := s.storage.RemoveFile(ctx, constants.VapiBucketId, tarFilePaths); err != nil {
			return errors.Wrapf(err, "failed to remove tar files, tarFilePaths=%v", tarFilePaths)
		}

		return nil
	})
}

func (s *service) DeleteReleasesByPackageId(
	ctx context.Context,
	packageId uint,
) error {
	tx := helpers.GetTx(ctx)

	pkg, err := domain.FindVapiPackageByID(tx, packageId)
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

	return s.deleteReleases(ctx, tx, pkg.Releases)
}

func (s *service) DeleteAllReleases(
	ctx context.Context,
) error {
	tx := helpers.GetTx(ctx)

	rels, err := domain.FindVapiReleases(tx)
	if err != nil {
		return err
	}

	me, err := s.users.GetUser(ctx)
	if err != nil {
		return err
	}

	for _, rel := range rels {
		if err := rel.Package.IsPermittedToEdit(me); err != nil {
			return err
		}
	}

	return s.deleteReleases(ctx, tx, rels)
}
