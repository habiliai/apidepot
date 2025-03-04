package domain

import (
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type TelegramMiniappPromotion struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index:telegram_miniapp_promotions_created_at_idx"`
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"index:telegram_miniapp_promotions_deleted_at_idx"`

	StackID uint  `gorm:"index:telegram_miniapp_promotions_stack_id_uniq_idx,unique,where:deleted_at=0"`
	Stack   Stack `gorm:"foreignKey:StackID"`

	Public                 bool
	AppBannerImageUrl      string
	AppScreenshotImageUrls datatypes.JSONSlice[string]
	Link                   string
	AppTitle               string
	AppDescription         string
	AppIconImageUrl        string

	Views []TelegramMiniappPromotionView `gorm:"foreignKey:TelegramMiniappPromotionID"`
}

type TelegramMiniappPromotionView struct {
	CreatedAt time.Time

	DeviceID string `gorm:"primaryKey"`

	TelegramMiniappPromotionID uint                     `gorm:"primaryKey"`
	TelegramMiniappPromotion   TelegramMiniappPromotion `gorm:"foreignKey:TelegramMiniappPromotionID"`

	NumViews uint
}

func (t *TelegramMiniappPromotion) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(t).Error, "failed to save telegram miniapp promotion")
}

func (t *TelegramMiniappPromotion) GetNumUniqueViews() int {
	return len(t.Views)
}
