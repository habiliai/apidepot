package domain

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type (
	CustomVapi struct {
		Model
		DeletedAt soft_delete.DeletedAt `gorm:"index"`

		Name string `gorm:"uniqueIndex:custom_vapis_name_idx_uniq,where:deleted_at=0"`

		StackID uint  `gorm:"uniqueIndex:custom_vapis_name_idx_uniq,where:deleted_at=0"`
		Stack   Stack `gorm:"foreignKey:StackID"`

		TarFilePath string
	}
)

func (v *CustomVapi) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(v).Error, "failed to save custom vapi")
}

func (v *CustomVapi) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(v).Error, "failed to delete custom vapi")
}
