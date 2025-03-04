package domain

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/soft_delete"
)

type (
	VapiPackage struct {
		Model
		DeletedAt soft_delete.DeletedAt `json:"-"`

		Name        string `gorm:"index:pkg_name_idx,unique,where:deleted_at=0" json:"name"`
		GitRepo     string `json:"git_repo"`
		GitBranch   string `json:"git_branch"`
		Domains     datatypes.JSONSlice[string]
		Description string
		Homepage    string

		OwnerId uint `json:"owner_id"`
		Owner   User `gorm:"foreignKey:OwnerId" json:"owner"`

		Releases   []VapiRelease `gorm:"foreignKey:PackageID" json:"releases"`
		VapiPoolId string
	}

	VapiEnvVar struct {
		Name        string `json:"name"`
		Default     string `json:"default"`
		Description string `json:"description"`
	}

	VapiRelease struct {
		Model
		DeletedAt soft_delete.DeletedAt `json:"-"`

		GitHash     string `json:"git_hash"`
		Version     string `gorm:"index:release_version_idx,unique,where:deleted_at=0" json:"version"`
		Description string `json:"description"`
		TarFilePath string `json:"tar_file_path"`
		Deprecated  bool   `json:"deprecated"`
		Suspended   bool   `json:"suspended"`
		Published   bool   `json:"published"`
		Domains     datatypes.JSONSlice[string]
		Homepage    string

		Dependencies []VapiRelease `gorm:"many2many:vapi_releases_dependencies" json:"-"`

		PackageID uint        `gorm:"index:release_version_idx,unique,where:deleted_at=0" json:"package_id"`
		Package   VapiPackage `gorm:"foreignKey:PackageID" json:"package"`

		EnvVars datatypes.JSONSlice[VapiEnvVar]
	}

	dfsFunc func(rel VapiRelease, parent *VapiRelease) error
)

func (v VapiRelease) Slug() string {
	alias := v.Package.Name

	major := semver.Major("v" + v.Version)
	if major != "v0" {
		alias += "/" + major
	}

	return alias
}

func (v VapiRelease) MajorVersion() string {
	return semver.Major("v" + v.Version)
}

func (v *VapiPackage) Save(tx *gorm.DB) error {
	return errors.WithStack(tx.Save(v).Error)
}

func (v *VapiPackage) Delete(tx *gorm.DB) error {
	return errors.WithStack(tx.Delete(v).Error)
}

func (v *VapiRelease) Save(tx *gorm.DB) error {
	return errors.WithStack(tx.Save(v).Error)
}

func (v *VapiRelease) Delete(tx *gorm.DB) error {
	return errors.WithStack(tx.Delete(v).Error)
}

func (v VapiRelease) dfs(
	tx *gorm.DB,
	dfsFunc dfsFunc,
	visited map[uint]struct{},
	parent *VapiRelease,
) error {
	if visited != nil {
		if _, ok := visited[v.ID]; ok {
			return nil
		}
		visited[v.ID] = struct{}{}
	}

	if err := tx.
		Model(v).
		Preload("Package").
		Association("Dependencies").
		Find(&v.Dependencies); err != nil {
		return errors.Wrapf(err, "failed to find dependencies")
	}

	for _, dep := range v.Dependencies {
		if err := dep.dfs(tx, dfsFunc, visited, &v); err != nil {
			return err
		}
	}

	return dfsFunc(v, parent)
}

func (v *VapiRelease) DFS(
	tx *gorm.DB,
	dfsFunc dfsFunc,
	options ...VapiReleaseDFSOptionFunc,
) error {
	opt := mergeVapiReleaseDFSOptions(options...)
	var visited map[uint]struct{}
	if opt.skipVisited {
		visited = map[uint]struct{}{}
	}
	return v.dfs(tx, dfsFunc, visited, nil)
}

func (v *VapiPackage) IsPermittedToEdit(user *User) error {
	if user == nil {
		return errors.Wrapf(tclerrors.ErrForbidden, "you are not authorized")
	}

	if user.ID != v.OwnerId {
		return errors.Wrapf(tclerrors.ErrForbidden, "you are not owned this package")
	}

	return nil
}

func GetVapiReleaseByID(tx *gorm.DB, id uint) (*VapiRelease, error) {
	var vapi VapiRelease
	if r := tx.
		Preload("Package").
		Find(&vapi, id); r.Error != nil {
		return nil, errors.Wrapf(r.Error, "failed to find vapi by id")
	} else if r.RowsAffected == 0 {
		return nil, errors.Wrapf(tclerrors.ErrNotFound, "vapi not found")
	}

	return &vapi, nil
}

func FindVapiReleaseByPackageIDAndVersion(
	db *gorm.DB,
	packageId uint,
	version string,
) (*VapiRelease, error) {
	var vapi VapiRelease
	if err := db.
		InnerJoins("Package").
		First(&vapi, "package_id = ? AND version = ?", packageId, version).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapi by package id and version")
	}

	return &vapi, nil
}

func FindVapiReleaseByPackageNameAndVersion(
	db *gorm.DB,
	name string,
	version string,
) (*VapiRelease, error) {
	var vapi VapiRelease
	if r := db.
		Table(
			"(?) as p, (?) as r",
			db.Model(&VapiPackage{}).Where("name = ?", name),
			db.Model(&VapiRelease{}).Where("version = ?", version),
		).
		Where(
			"p.id = r.package_id",
		).
		Select("r.*").
		Limit(1).
		Preload("Package").
		Find(&vapi); r.Error != nil {
		return nil, errors.Wrapf(r.Error, "failed to find vapi by name and version")
	} else if r.RowsAffected == 0 {
		return nil, errors.Wrapf(tclerrors.ErrNotFound, "vapi not found")
	}

	return &vapi, nil
}

func FindLatestVapiReleaseByPackageID(
	db *gorm.DB,
	packageId uint,
) (*VapiRelease, error) {
	var vapi VapiRelease
	if err := db.
		Where("package_id = ?", packageId).
		Order("apidepot.version_to_int(version) DESC").
		First(&vapi).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find latest vapi release by package id")
	}

	return &vapi, nil
}

func FindVapiPackageByNameAndProjectID(
	db *gorm.DB,
	name string,
	projectId uint,
) (*VapiPackage, error) {
	var vapi VapiPackage
	if err := db.
		Preload(clause.Associations).
		Where("name = ? AND project_id = ?", name, projectId).
		First(&vapi).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapi by name and project id")
	}

	return &vapi, nil
}

func FindVapiPackageByID(
	db *gorm.DB,
	id uint,
) (*VapiPackage, error) {
	var vapi VapiPackage
	if err := db.
		Preload(clause.Associations).
		First(&vapi, id).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapi by id")
	}

	return &vapi, nil
}

func FindVapiReleasesByPackageID(
	db *gorm.DB,
	packageId uint,
) ([]VapiRelease, error) {
	var vapis []VapiRelease
	if err := db.
		Preload("Package").
		Where("package_id = ?", packageId).
		Find(&vapis).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapis by package id")
	}

	return vapis, nil
}

func FindVapiReleases(db *gorm.DB) ([]VapiRelease, error) {
	var vapis []VapiRelease
	if err := db.Preload("Package").Find(&vapis).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapis")
	}

	return vapis, nil
}

func FindVapiPackagesByProjectID(db *gorm.DB, projectId uint) ([]VapiPackage, error) {
	var vapis []VapiPackage
	if err := db.
		Where("project_id = ?", projectId).
		Find(&vapis).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapis by project id")
	}

	return vapis, nil
}

func FindVapiPackages(db *gorm.DB) ([]VapiPackage, error) {
	var vapis []VapiPackage
	if err := db.Find(&vapis).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find vapis")
	}

	return vapis, nil
}
