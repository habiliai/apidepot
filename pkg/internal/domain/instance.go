package domain

import (
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type (
	Instance struct {
		Model
		DeletedAt soft_delete.DeletedAt

		StackID uint  `gorm:"uniqueIndex:instances_stack_zone_uniq_idx,where:deleted_at is null"`
		Stack   Stack `gorm:"foreignKey:StackID"`

		Zone        tcltypes.InstanceZone `gorm:"uniqueIndex:instances_stack_zone_uniq_idx,where:deleted_at is null"`
		NumReplicas uint                  `gorm:"default:0"`
		MaxReplicas uint                  `gorm:"default:1"`
		State       InstanceState         `gorm:"default:0"`
		Name        string

		AppliedK8sYaml string
	}

	InstanceState uint
)

const (
	InstanceStateNone InstanceState = iota
	InstanceStateRunning
	InstanceStateInitialize
	InstanceStateReady
)

var (
	InstanceStates = []InstanceState{
		InstanceStateNone,
		InstanceStateRunning,
		InstanceStateInitialize,
		InstanceStateReady,
	}
)

func (i *Instance) Save(tx *gorm.DB) error {
	return errors.Wrapf(tx.Save(i).Error, "failed to save instance")
}

func (i *Instance) Delete(tx *gorm.DB) error {
	return errors.Wrapf(tx.Delete(i).Error, "failed to delete instance")
}

func (i *Instance) updateState(tx *gorm.DB, state InstanceState) error {
	if state == InstanceStateNone {
		return errors.Errorf("invalid state: %v", state)
	}

	i.State = state
	return errors.Wrapf(tx.Model(&Instance{}).Where("id = ?", i.ID).Update("state", state).Error, "failed to update instance state")
}

func (i *Instance) TransitionToRunning(tx *gorm.DB) error {
	return i.updateState(tx, InstanceStateRunning)
}

func (i *Instance) TransitionToInitialize(tx *gorm.DB) error {
	return i.updateState(tx, InstanceStateInitialize)
}

func (i *Instance) TransitionToReady(tx *gorm.DB) error {
	return i.updateState(tx, InstanceStateReady)
}

func FindInstanceById(tx *gorm.DB, id uint) (*Instance, error) {
	var instance Instance
	if err := tx.First(&instance, id).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &instance, nil
}

func FindInstancesByStackId(tx *gorm.DB, id uint) ([]Instance, error) {
	var instances []Instance
	if err := tx.Find(&instances, "stack_id = ?", id).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return instances, nil
}
