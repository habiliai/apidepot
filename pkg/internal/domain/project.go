package domain

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Project struct {
	Model
	DeletedAt soft_delete.DeletedAt

	OwnerID uint
	Owner   User `gorm:"foreignKey:OwnerID"`

	Name        string
	Description string

	Stacks []Stack `gorm:"foreignKey:ProjectID"`
}

func (p *Project) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(p).Error, "failed to save project")
}

func (p *Project) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(p).Error, "failed to delete project")
}

func GetProjectByID(db *gorm.DB, id uint) (Project, error) {
	var project Project
	if err := db.
		Preload("Stacks").
		First(&project, id).Error; err != nil {
		return project, err
	}

	return project, nil
}

func FindProjects(db *gorm.DB) ([]Project, error) {
	var projects []Project
	if err := db.Preload("Stacks").Find(&projects).Error; err != nil {
		return nil, err
	}

	return projects, nil
}

func DeleteProjectByID(db *gorm.DB, id uint) error {
	return errors.Wrapf(db.Delete(&Project{}, id).Error, "failed to delete project")
}

func FindProjectById(db *gorm.DB, id uint) (*Project, error) {
	var project Project
	if r := db.Find(&project, "id = ?", id); r.Error != nil {
		return nil, r.Error
	} else if r.RowsAffected == 0 {
		return nil, errors.Wrapf(tclerrors.ErrNotFound, "project not found")
	}

	return &project, nil
}
