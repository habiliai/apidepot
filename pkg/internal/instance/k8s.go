package instance

import (
	"context"
	"fmt"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/k8s"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	"github.com/habiliai/apidepot/pkg/internal/util/functx/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

func (s *service) applyK8s(
	ctx context.Context,
	instance *domain.Instance,
	timeout time.Duration,
) error {
	tx := helpers.GetTx(ctx)
	ctx, fDone := functx.WithFuncTx(ctx)
	defer fDone(ctx, true)
	k8sClient, err := s.k8sClientPool.GetClient(instance.Zone)
	if err != nil {
		return err
	}

	oldK8sYaml := instance.AppliedK8sYaml
	newK8sYaml, err := s.renderK8sYamlValues(ctx, &instance.Stack)
	if err != nil {
		return err
	}

	if oldK8sYaml == newK8sYaml {
		return nil
	}

	oldObjects, err := k8syaml.ParseK8sYaml(oldK8sYaml)
	if err != nil {
		return err
	}
	newObjects, err := k8syaml.ParseK8sYaml(newK8sYaml)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeoutCause(ctx, timeout, tclerrors.ErrTimeout)
	defer cancel()

	if err := s.applyNamespace(ctx, k8sClient, &instance.Stack); err != nil {
		return err
	}

	if err := k8sClient.Upgrade(ctx, oldObjects, newObjects, k8s.WithApplyCheckFn(func(ctx context.Context) error {
		for {
			if err := k8sClient.Wait(
				ctx,
				"pod",
				instance.Stack.Namespace(),
				fmt.Sprintf("shaple.io/project.id=%d,shaple.io/stack.id=%d", instance.Stack.Project.ID, instance.Stack.ID),
				"ready",
			); err != nil {
				if !errors.Is(err, tclerrors.ErrNotFound) {
					return err
				}
				continue
			}

			break
		}

		if s.stackConfig.SkipHealthCheck {
			return nil
		}

		for allOk := false; !allOk; {
			results, err := s.isAvailable(ctx, instance, 500*time.Millisecond)
			if err != nil {
				return err
			}

			allOk = true
			for _, ok := range results {
				if !ok {
					allOk = false
					time.Sleep(250 * time.Millisecond)
					break
				}
			}
		}

		return nil
	})); err != nil {
		return err
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		instance.AppliedK8sYaml = newK8sYaml
		return instance.Save(tx)
	}); err != nil {
		return err
	}

	fDone(ctx, false)
	return nil
}

func (s *service) applyNamespace(ctx context.Context, k8sClient k8s.Client, stack *domain.Stack) error {
	logger.Info("apply", "namespace", stack.Namespace())
	objects, err := s.k8sYamlService.RenderYaml([]string{
		"common/namespace.yaml",
	}, s.k8sYamlService.NewValuesFromStack(stack))
	if err != nil {
		return err
	}

	if err := k8sClient.ApplyYamlFile(ctx, objects); err != nil {
		return err
	}

	return nil
}

func (s *service) deleteNamespace(ctx context.Context, k8sClient k8s.Client, stack *domain.Stack, force bool) error {
	logger.Info("delete", "namespace", stack.Namespace())
	objects, err := s.k8sYamlService.RenderYaml([]string{
		"common/namespace.yaml",
	}, s.k8sYamlService.NewValuesFromStack(stack))
	if err != nil {
		return err
	}

	if err := k8sClient.DeleteYamlFile(ctx, objects, true, k8s.WithForce(s.stackConfig.ForceDelete || force)); err != nil {
		return err
	}

	return nil
}

func (s *service) renderK8sYamlValues(ctx context.Context, stack *domain.Stack) (string, error) {
	values := s.k8sYamlService.NewValuesFromStack(stack)
	k8sYamlFiles := []string{
		"common/network-policy.yaml",
		"common/ingress.yaml",
		"common/configmap.yaml",
		"database/configmap.yaml",
		"database/secret.yaml",
	}

	if stack.AuthEnabled {
		values = values.WithAuth(s.smtpConfig)
		k8sYamlFiles = append(k8sYamlFiles, "auth/configmap.yaml", "auth/secret.yaml", "auth/deployment.yaml", "auth/service.yaml")
	}

	if stack.StorageEnabled {
		values = values.WithStorage(s.s3Config, stack.DefaultRegion)
		k8sYamlFiles = append(k8sYamlFiles, "storage/service.yaml", "storage/secret.yaml", "storage/deployment.yaml", "storage/configmap.yaml")
	}

	if stack.PostgrestEnabled {
		values = values.WithPostgrest()
		k8sYamlFiles = append(k8sYamlFiles, "postgrest/deployment.yaml", "postgrest/service.yaml")
	}

	if len(stack.Vapis) > 0 {
		vapiReleases, err := s.vapis.GetAllDependenciesOfVapiReleases(ctx, stack.GetVapiReleases())
		if err != nil {
			return "", err
		}

		vapiValues, err := s.k8sYamlService.GetVapiYamlValues(ctx, vapiReleases, stack.VapiEnvVars)
		if err != nil {
			return "", err
		}
		values = values.WithVapis(vapiValues)
		k8sYamlFiles = append(k8sYamlFiles,
			"vapi/configmap.yaml",
			"vapi/secret.yaml",
			"vapi/service.yaml",
			"vapi/deployment.yaml",
		)
	}

	k8sYaml, err := s.k8sYamlService.RenderYaml(k8sYamlFiles, values)
	if err != nil {
		return "", err
	}

	return k8sYaml, nil
}
