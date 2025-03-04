package domain

import (
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"strings"
)

type Organization struct {
	Model
	soft_delete.DeletedAt

	Name string `gorm:"index:organizations_name_idx,unique,where:deleted_at=0"`
}

func (o *Organization) String() string {
	var builder strings.Builder
	builder.WriteString("Organization{")
	builder.WriteString("ID: ")
	builder.WriteString(fmt.Sprintf("%d", o.ID))
	builder.WriteString(", Name: ")
	builder.WriteString("'" + o.Name + "'")
	builder.WriteString("}")
	return builder.String()
}

func (o *Organization) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(o).Error, "failed to save organization %v", o)
}

func (o *Organization) IsMember(user User) bool {
	//if user.SelfOrganizationId == o.ID {
	//	return true
	//}
	//
	//for _, member := range o.Members {
	//	if member.ID == user.ID {
	//		return true
	//	}
	//}

	return false
}

func GetOrganizationById(tx *gorm.DB, id uint) (organization Organization, err error) {
	err = errors.Wrapf(tx.First(&organization, id).Error, "failed to find organization by id %d", id)
	return
}
