package domain

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type ServiceTemplate struct {
	Model
	DeletedAt soft_delete.DeletedAt

	Name            string
	ConceptImageUrl string
	PrimaryImageUrl string

	Description string
	Detail      string

	GitRepo string
	GitHash string

	TSV string `gorm:"type:tsvector; index:service_templates_tsv_idx,type:gin,option:gin_tsv_ops,where:deleted_at=0"`

	VapiIds datatypes.JSONSlice[uint] // VapiRelease IDs
}

func (s *ServiceTemplate) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(s).Error, "failed to save service template")
}

func (s *ServiceTemplate) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(s).Error, "failed to delete service template")
}

func FindServiceTemplateByID(db *gorm.DB, id uint) (*ServiceTemplate, error) {
	var st ServiceTemplate
	if r := db.Find(&st, id); r.Error != nil {
		return nil, errors.Wrapf(r.Error, "failed to find service template by id: %d", id)
	} else if r.RowsAffected == 0 {
		return nil, errors.Wrapf(tclerrors.ErrNotFound, "service template not found by id: %d", id)
	}

	return &st, nil
}

func init() {
	functionsOnAfterMigration = append(functionsOnAfterMigration, func(db *gorm.DB) error {
		if err := db.Exec(`
create or replace function service_templates_tsv_generator()
returns trigger as $$

begin
    new.tsv := to_tsvector(new.name || ' ' || new.detail);
    return new;
end;
$$
language plpgsql;

create or replace trigger service_templates_tsv_generator before insert or update on service_templates
    for each row execute function service_templates_tsv_generator();
`).Error; err != nil {
			return errors.Wrapf(err, "failed to create service_templates_tsv_generator function")
		}

		return nil
	})
}
