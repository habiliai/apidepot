package instance

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type (
	DeployStackInput struct {
		Timeout *string `json:"timeout"`
	}
)

func (s *service) LaunchInstance(
	ctx context.Context,
	instanceId uint,
) error {
	tx := helpers.GetTx(ctx)

	instance, err := s.GetInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	k8sClient, err := s.k8sClientPool.GetClient(instance.Zone)
	if err != nil {
		return err
	}

	if instance.State != domain.InstanceStateReady {
		return errors.Wrapf(tclerrors.ErrForbidden, "instance is not running")
	}

	return tx.Transaction(func(tx *gorm.DB) (error error) {
		instance.State = domain.InstanceStateRunning
		if err := instance.Save(tx); err != nil {
			return err
		}

		if instance.AppliedK8sYaml == "" {
			return errors.Wrapf(tclerrors.ErrForbidden, "stack is not deployed at instance")
		}

		objects, err := k8syaml.ParseK8sYaml(instance.AppliedK8sYaml)
		if err != nil {
			return err
		}

		defer func() {
			if error == nil {
				return
			}

			if err := k8sClient.Apply(ctx, objects); err != nil {
				logger.Error("failed to apply k8s objects", "err", err)
			}
		}()
		if err := k8sClient.Delete(ctx, objects, true); err != nil {
			return err
		}

		return nil
	})
}

func (s *service) StopInstance(
	ctx context.Context,
	instanceId uint,
) error {
	tx := helpers.GetTx(ctx)

	instance, err := s.GetInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	k8sClient, err := s.k8sClientPool.GetClient(instance.Zone)
	if err != nil {
		return err
	}

	if instance.State != domain.InstanceStateRunning {
		return errors.Wrapf(tclerrors.ErrForbidden, "instance is not running")
	}

	return tx.Transaction(func(tx *gorm.DB) (error error) {
		instance.State = domain.InstanceStateReady
		if err := instance.Save(tx); err != nil {
			return err
		}

		if instance.AppliedK8sYaml != "" {
			objects, err := k8syaml.ParseK8sYaml(instance.AppliedK8sYaml)
			if err != nil {
				return err
			}

			defer func() {
				if error == nil {
					return
				}

				if err := k8sClient.Apply(ctx, objects); err != nil {
					logger.Error("failed to apply k8s objects", "err", err)
				}
			}()
			if err := k8sClient.Delete(ctx, objects, true); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *service) RestartInstance(
	ctx context.Context,
	instanceId uint,
) error {
	if err := s.StopInstance(ctx, instanceId); err != nil {
		return err
	}

	time.Sleep(150 * time.Millisecond)

	return s.LaunchInstance(ctx, instanceId)
}
