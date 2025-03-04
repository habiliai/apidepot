package vapi

import (
	"context"
	"fmt"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/semver"
)

type SearchVapisInput struct {
	Name    *string `json:"name" form:"name" binding:"omitempty"`
	Version *string `json:"version" form:"version" binding:"omitempty"`

	PageNum  int `json:"page_num" form:"page_num" binding:"omitempty"`
	PageSize int `json:"page_size" form:"page_size" binding:"omitempty"`
}

type SearchVapisOutput struct {
	Releases []domain.VapiRelease `json:"releases"`
	NumTotal int64                `json:"num_total"`
	NextPage *int                 `json:"next_page"`
}

func (s *service) FindVapiReleaseOnStack(
	ctx context.Context,
	stackId uint,
	name string,
	version string,
	projectId uint,
) (*domain.VapiRelease, error) {
	tx := helpers.GetTx(ctx)

	var vapi domain.VapiRelease
	if err := tx.
		Table(
			"(?) as p, (?) as r",
			tx.Model(&domain.VapiPackage{}).Where("name = ? AND project_id = ?", name, projectId),
			tx.Model(&domain.VapiRelease{}).Where("version = ?", version),
		).
		Where(
			"p.id = r.package_id AND ((r.access = 'public' AND r.published = true) OR (? IN (?)))",
			stackId,
			tx.Table("vapi_packages_borrowers as b").Select("b.stack_id").Where("b.vapi_package_id = p.id"),
		).
		Select("r.*").
		Limit(1).
		Preload("Package").
		First(&vapi).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapi by name and version")
	}

	return &vapi, nil
}

func (s *service) SearchVapis(
	ctx context.Context,
	input SearchVapisInput,
) (output SearchVapisOutput, err error) {
	tx := helpers.GetTx(ctx)

	stmt := tx.Model(&domain.VapiPackage{})

	if input.Name != nil {
		stmt = stmt.
			Select(
				"vapi_packages.*, ts_rank_cd(to_tsvector(vapi_packages.name), to_tsquery(?)) as rank",
				*input.Name,
			).
			Order("rank DESC").
			Where("to_tsvector(vapi_packages.name) @@ to_tsquery(?)", *input.Name)
	} else {
		stmt = stmt.Order("vapi_packages.id DESC")
	}

	pageNum := 1
	if input.PageNum > 0 {
		pageNum = input.PageNum
	}

	pageSize := 10
	if input.PageSize > 0 {
		pageSize = input.PageSize
	}

	stmt = stmt.Count(&output.NumTotal)
	if err := stmt.Error; err != nil {
		return output, errors.Wrapf(err, "failed to count vapi packages")
	}

	nextCount := int64(0)
	stmt = stmt.Offset(pageNum * pageSize).Limit(pageSize).Count(&nextCount)
	if err := stmt.Error; err != nil {
		return output, errors.Wrapf(err, "failed to count next vapi packages")
	}

	if nextCount > 0 {
		output.NextPage = gog.PtrOf(pageNum + 1)
	}

	offset := (pageNum - 1) * pageSize
	if offset == 0 {
		offset = -1 // to avoid offset ZeroValue
	}
	stmt = stmt.Offset(offset).Limit(pageSize)

	var packages []domain.VapiPackage
	stmt = stmt.Find(&packages)

	if err := stmt.Error; err != nil {
		return output, errors.Wrapf(err, "failed to search vapi packages")
	}

	for _, pkg := range packages {
		var rel domain.VapiRelease
		if input.Version == nil {
			if err := tx.Order("apidepot.version_to_int(version) DESC").
				Preload("Package").
				First(&rel, "package_id = ?", pkg.ID).
				Error; err != nil {
				return output, errors.Wrapf(err, "failed to find latest release")
			}
		} else {
			if err := tx.First(&rel, "package_id = ? AND version = ?", pkg.ID, *input.Version).
				Preload("Package").
				Error; err != nil {
				return output, errors.Wrapf(err, "failed to find release")
			}
		}
		output.Releases = append(output.Releases, rel)
	}

	return
}

func (s *service) GetPackagesByOwnerId(ctx context.Context, ownerId uint) ([]domain.VapiPackage, error) {
	tx := helpers.GetTx(ctx)

	var pkgs []domain.VapiPackage
	err := errors.WithStack(tx.Find(&pkgs, "owner_id = ?", ownerId).Error)

	return pkgs, err
}

func (s *service) GetAllDependenciesOfVapiReleases(
	ctx context.Context,
	vapiReleases []domain.VapiRelease,
) ([]domain.VapiRelease, error) {
	tx := helpers.GetTx(ctx)

	var foundVapiReleases []domain.VapiRelease
	for _, rel := range vapiReleases {
		if err := rel.DFS(tx, func(dep domain.VapiRelease, _ *domain.VapiRelease) error {
			logger.Debug("found dependency", "packageName", dep.Package.Name, "version", dep.Version)
			foundVapiReleases = append(foundVapiReleases, dep)
			return nil
		}, domain.SkipVisited()); err != nil {
			return nil, err
		}
	}

	logger.Debug("found vapi", "num", len(foundVapiReleases), "releases", foundVapiReleases)

	set := map[string]domain.VapiRelease{}
	for _, rel := range foundVapiReleases {
		logger.Debug("checking vapi", "name", rel.Package.Name, "version", rel.Version)

		major := semver.Major("v" + rel.Version)
		key := fmt.Sprintf("%s/%s", rel.Package.Name, major)

		old, ok := set[key]
		if !ok {
			set[key] = rel
			continue
		}

		if semver.Compare("v"+old.Version, "v"+rel.Version) >= 0 {
			continue
		}

		set[key] = rel
	}

	logger.Debug("print", "set", set)

	return maps.Values(set), nil
}
