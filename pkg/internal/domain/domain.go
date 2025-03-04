package domain

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	functionsOnAfterMigration []func(db *gorm.DB) error
)

func AutoMigrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE SCHEMA IF NOT EXISTS apidepot`).Error; err != nil {
		return errors.Wrapf(err, "failed to create schema")
	}

	if err := db.Exec(`
CREATE OR REPLACE FUNCTION apidepot.version_to_int(version TEXT) RETURNS INTEGER AS $$
DECLARE
  ver_arr TEXT[];
  ver_int INTEGER;
BEGIN
    ver_arr := regexp_split_to_array(version, '\.');
    ver_int := ver_arr[1]::INTEGER * 100000 + ver_arr[2]::INTEGER * 1000 + ver_arr[3]::INTEGER;
    RETURN ver_int;
END;
$$ LANGUAGE plpgsql;
`).Error; err != nil {
		return errors.Wrapf(err, "failed to create version_to_int function")
	}

	if err := errors.Wrapf(db.
		AutoMigrate(
			&Organization{},
			&User{},
			&VapiPackage{},
			&VapiRelease{},
			&Project{},
			&Stack{},
			&Instance{},
			&StackVapi{},
			&CliApp{},
			&ServiceTemplate{},
			&InstanceHistory{},
			&StackHistory{},
			&CustomVapi{},
			&TelegramMiniappPromotion{},
			&TelegramMiniappPromotionView{},
		), "failed to auto migrate"); err != nil {
		return err
	}

	for _, f := range functionsOnAfterMigration {
		if err := f(db); err != nil {
			return err
		}
	}

	return nil
}

func DropAll(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&TelegramMiniappPromotionView{},
		&TelegramMiniappPromotion{},
		&CustomVapi{},
		&StackHistory{},
		&InstanceHistory{},
		"service_template_vapi_releases",
		&ServiceTemplate{},
		&CliApp{},
		&StackVapi{},
		&Instance{},
		&Stack{},
		&Project{},
		"vapi_releases_dependencies",
		&VapiRelease{},
		&VapiPackage{},
		&User{},
		&Organization{},
	)
}
