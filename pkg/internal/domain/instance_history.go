package domain

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type InstanceHistory struct {
	Model
	soft_delete.DeletedAt

	InstanceID uint
	Instance   Instance `gorm:"foreignKey:InstanceID"`

	Running bool
	Billed  bool
}

func (h *InstanceHistory) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(h).Error, "failed to save instance history")
}

func (h *InstanceHistory) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(h).Error, "failed to delete instance history")
}
