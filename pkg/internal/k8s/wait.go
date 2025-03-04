package k8s

import (
	"context"
	"fmt"
	"github.com/habiliai/apidepot/pkg/errors"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

func (k *client) Wait(
	ctx context.Context,
	kind,
	namespace string,
	selector string,
	forCondition string,
) error {
	dri, err := k.getResourceInterface(kind, namespace)
	if err != nil {
		return err
	}

	objects, err := dri.List(ctx, v1.ListOptions{
		LabelSelector: selector,
		Watch:         false,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list")
	}

	if len(objects.Items) == 0 {
		return errors.Wrapf(tclerrors.ErrNotFound, "no resources found")
	}

	met := map[string]bool{}
	for _, object := range objects.Items {
		met[object.GetName()] = false
	}

	wi, err := dri.Watch(ctx, v1.ListOptions{
		LabelSelector: selector,
		Watch:         true,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to watch")
	}
	defer wi.Stop()

	allMet := func() bool {
		for _, v := range met {
			if !v {
				return false
			}
		}
		logger.Debug("all", "met", met)
		return true
	}

	for !allMet() {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				if errors.As(err, &tclerrors.ErrTimeout) {
					return errors.Wrapf(err, "timed out")
				}
				logger.Warn("context done", "err", err)
			}
			return nil
		case event, ok := <-wi.ResultChan():
			if !ok {
				return errors.Wrapf(tclerrors.ErrTimeout, "watch channel closed")
			}

			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(event.Object)
			if err != nil {
				return errors.Wrapf(err, "failed to convert to unstructured")
			}

			name, found, err := unstructured.NestedString(obj, "metadata", "name")
			if err != nil {
				return errors.Wrapf(err, "failed to get name")
			}
			if !found {
				return errors.Wrapf(tclerrors.ErrRuntime, "no name found")
			}
			if _, ok := met[name]; !ok {
				continue
			}

			phase, found, err := unstructured.NestedString(obj, "status", "phase")
			if err != nil {
				return errors.Wrapf(err, "failed to get phase")
			} else if found {
				if phase == "Pending" {
					continue
				}
			}

			conditions, found, err := unstructured.NestedSlice(obj, "status", "conditions")
			if err != nil {
				return errors.Wrapf(err, "failed to get conditions")
			} else if found {
				for _, condition := range conditions {
					logger.Debug("print status", "condition", condition)
					conditionMap, ok := condition.(map[string]interface{})
					if !ok {
						return errors.Wrapf(tclerrors.ErrRuntime, "failed to cast condition to map")
					}

					if conditionMap["type"] == nil {
						return errors.Wrapf(tclerrors.ErrRuntime, "condition type is nil")
					}
					if conditionMap["status"] == nil {
						return errors.Wrapf(tclerrors.ErrRuntime, "condition status is nil")
					}

					if strings.EqualFold(conditionMap["type"].(string), forCondition) &&
						strings.EqualFold(conditionMap["status"].(string), "true") {
						met[name] = true

						logger.Info(fmt.Sprintf("%s/%s is met in condition='%s'", kind, name, forCondition))
						break
					}
				}
			}
		}
	}

	return nil
}
