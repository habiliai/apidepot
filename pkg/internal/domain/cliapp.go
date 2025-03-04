package domain

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type CliApp struct {
	Model
	DeletedAt soft_delete.DeletedAt `gorm:"index"`

	Host         string
	RefreshToken string

	AppId     string `gorm:"uniqueIndex:cli_app_app_id_uniq,where:deleted_at=0,not null"`
	AppSecret []byte

	OwnerID uint
	Owner   User
}

func (c *CliApp) Save(db *gorm.DB) error {
	return errors.WithStack(db.Save(c).Error)
}

func (c *CliApp) Delete(db *gorm.DB) error {
	return errors.WithStack(db.Delete(c).Error)
}

func GetCliAppByAppId(db *gorm.DB, appId string) (*CliApp, error) {
	var cliApp CliApp
	if r := db.
		Preload("Owner").
		Where("app_id = ?", appId).
		Find(&cliApp); r.Error != nil {
		return nil, errors.WithStack(r.Error)
	} else if r.RowsAffected == 0 {
		return nil, errors.Wrapf(tclerrors.ErrNotFound, "cli app not found")
	}

	return &cliApp, nil
}
