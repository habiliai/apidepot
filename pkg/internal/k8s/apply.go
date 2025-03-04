package k8s

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	"github.com/habiliai/apidepot/pkg/internal/util/functx/v2"
	"github.com/pkg/errors"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"strings"
)

func (k *client) ApplyYamlFile(ctx context.Context, yamlFileContents string) error {
	yamlFileContents = strings.TrimSpace(yamlFileContents)
	if yamlFileContents == "" || yamlFileContents == "---" {
		return nil
	}

	objects, err := k8syaml.ParseK8sYaml(yamlFileContents)
	if err != nil {
		return err
	}

	return k.Apply(ctx, objects)
}

func (k *client) Apply(ctx context.Context, objects []unstructured.Unstructured) error {
	for i := range objects {
		object := &objects[i]
		gvk := object.GroupVersionKind()

		gr, err := restmapper.GetAPIGroupResources(k.client.Discovery())
		if err != nil {
			return errors.Wrapf(err, "failed to get api group resources")
		}

		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return errors.Wrapf(err, "failed to get rest mapping")
		}

		var dri dynamic.ResourceInterface
		if object.GetNamespace() == "" {
			dri = k.dynamicClient.Resource(mapping.Resource)
		} else {
			dri = k.dynamicClient.Resource(mapping.Resource).Namespace(object.GetNamespace())
		}

		// client-side apply logic
		if _, err := dri.Create(ctx, object, v1.CreateOptions{}); err != nil {
			if !errors2.IsAlreadyExists(err) {
				return errors.Wrapf(err, "failed to apply")
			}

			beforeObject, err := dri.Get(ctx, object.GetName(), v1.GetOptions{})
			if err != nil {
				return errors.Wrapf(err, "failed to get")
			}

			object.SetResourceVersion(beforeObject.GetResourceVersion())

			if _, err := dri.Update(ctx, object, v1.UpdateOptions{}); err != nil {
				return errors.Wrapf(err, "failed to update")
			}
		}
	}

	return nil
}

func (k *client) Upgrade(
	ctx context.Context,
	oldObjects []unstructured.Unstructured,
	newObjects []unstructured.Unstructured,
	options ...OptionsFunc,
) error {
	ctx, fDone := functx.WithFuncTx(ctx)
	defer fDone(ctx, true)
	logger.Info("upgrade")
	option := mergeK8sClientOptions(options...)

	oldObjectMap, newObjectMap, markings, err := k8syaml.DiffK8sObjects(oldObjects, newObjects)
	if err != nil {
		return err
	}

	createTargets := make([]unstructured.Unstructured, 0, len(newObjects))
	updateNewTargets := make([]unstructured.Unstructured, 0, len(newObjects))
	updateOldTargets := make([]unstructured.Unstructured, 0, len(oldObjects))
	deleteTargets := make([]unstructured.Unstructured, 0, len(oldObjects))
	for key, value := range markings {
		switch value {
		case k8syaml.K8sObjectDiffAdded:
			obj, ok := newObjectMap[key]
			if !ok {
				return errors.Errorf("failed to get new object")
			}

			createTargets = append(createTargets, obj)
		case k8syaml.K8sObjectDiffChanged:
			newObj, ok := newObjectMap[key]
			if !ok {
				return errors.Errorf("failed to get new object")
			}

			oldObj, ok := oldObjectMap[key]
			if !ok {
				return errors.Errorf("failed to get old object")
			}

			updateNewTargets = append(updateNewTargets, newObj)
			updateOldTargets = append(updateOldTargets, oldObj)
		case k8syaml.K8sObjectDiffDeleted:
			if value != k8syaml.K8sObjectDiffDeleted {
				continue
			}

			obj, ok := oldObjectMap[key]
			if !ok {
				return errors.Errorf("failed to get old object")
			}

			deleteTargets = append(deleteTargets, obj)
		}
	}

	functx.AddRollback(ctx, func(ctx context.Context) {
		if err := k.Delete(ctx, createTargets, true, options...); err != nil {
			logger.Warn("failed to rollback create")
		}
	})
	if err := k.Apply(ctx, createTargets); err != nil {
		return err
	}

	functx.AddRollback(ctx, func(ctx context.Context) {
		if err := k.Apply(ctx, updateOldTargets); err != nil {
			logger.Warn("failed to rollback update old")
		}
	})
	if err := k.Apply(ctx, updateNewTargets); err != nil {
		return err
	}

	if option.ApplyCheckFn != nil {
		if err := option.ApplyCheckFn(ctx); err != nil {
			return err
		}
	}

	functx.AddRollback(ctx, func(ctx context.Context) {
		if err := k.Apply(ctx, deleteTargets); err != nil {
			logger.Warn("failed to rollback delete")
		}
	})
	if err := k.Delete(ctx, deleteTargets, true, options...); err != nil {
		return err
	}

	fDone(ctx, false)
	return nil
}
