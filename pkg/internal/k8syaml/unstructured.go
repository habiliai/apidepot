package k8syaml

import (
	"github.com/pkg/errors"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

type (
	K8sObjectDiffKey struct {
		GVK  schema.GroupVersionKind
		Name string
	}
	K8sObjectDiffValue uint
)

const (
	K8sObjectDiffAdded K8sObjectDiffValue = iota
	K8sObjectDiffDeleted
	K8sObjectDiffChanged
)

func ParseK8sYaml(yamlFileContents string) ([]unstructured.Unstructured, error) {
	decoder := yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(yamlFileContents), 256)
	var objects []unstructured.Unstructured
	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrapf(err, "failed to decode")
		}

		obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode")
		}

		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert to unstructured")
		}

		unstructuredObj := unstructured.Unstructured{Object: unstructuredMap}
		objects = append(objects, unstructuredObj)
	}

	return objects, nil
}

func DiffK8sObjects(oldObjects []unstructured.Unstructured, newObjects []unstructured.Unstructured) (
	map[K8sObjectDiffKey]unstructured.Unstructured,
	map[K8sObjectDiffKey]unstructured.Unstructured,
	map[K8sObjectDiffKey]K8sObjectDiffValue,
	error,
) {
	diff := make(map[K8sObjectDiffKey]K8sObjectDiffValue)
	oldObjectMap := make(map[K8sObjectDiffKey]unstructured.Unstructured)
	newObjectMap := make(map[K8sObjectDiffKey]unstructured.Unstructured)
	namespace := ""

	onBeforeAdding := func(obj *unstructured.Unstructured) error {
		if namespace == "" {
			namespace = obj.GetNamespace()
			return nil
		}

		if namespace != obj.GetNamespace() {
			return errors.Errorf(
				"namespace mismatch: %s != %s. resource name: %s, kind: %s",
				namespace,
				obj.GetNamespace(),
				obj.GetName(),
				obj.GetKind(),
			)
		}

		return nil
	}

	for _, obj := range oldObjects {
		if err := onBeforeAdding(&obj); err != nil {
			return nil, nil, nil, err
		}

		key := K8sObjectDiffKey{
			GVK:  obj.GroupVersionKind(),
			Name: obj.GetName(),
		}
		oldObjectMap[key] = obj
	}

	for _, obj := range newObjects {
		if err := onBeforeAdding(&obj); err != nil {
			return nil, nil, nil, err
		}

		key := K8sObjectDiffKey{
			GVK:  obj.GroupVersionKind(),
			Name: obj.GetName(),
		}
		newObjectMap[key] = obj
	}

	for key := range oldObjectMap {
		if _, ok := newObjectMap[key]; ok {
			diff[key] = K8sObjectDiffChanged
		} else {
			diff[key] = K8sObjectDiffDeleted
		}
	}

	for key := range newObjectMap {
		if _, ok := oldObjectMap[key]; !ok {
			diff[key] = K8sObjectDiffAdded
		}
	}

	return oldObjectMap, newObjectMap, diff, nil
}
