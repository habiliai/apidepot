package k8s

import (
	"context"
	"github.com/habiliai/apidepot/pkg/errors"
	"github.com/pkg/errors"
	"io"
	v2 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"strings"
)

func (k *client) GetLogs(
	ctx context.Context,
	selector, namespace string,
) (map[string]string, error) {
	pods, err := k.client.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list pods")
	}

	results := map[string]string{}
	for _, pod := range pods.Items {
		req := k.client.CoreV1().Pods(namespace).GetLogs(pod.GetName(), &v2.PodLogOptions{})
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to stream logs")
		}
		defer podLogs.Close()

		buf := new(strings.Builder)

		if _, err := io.Copy(buf, podLogs); err != nil {
			return nil, errors.Wrapf(err, "failed to copy logs")
		}

		results[pod.GetName()] = buf.String()
	}

	return results, nil
}

func (k *client) getResourceInterface(kind, namespace string) (dynamic.ResourceInterface, error) {
	groupResources, err := restmapper.GetAPIGroupResources(k.client.Discovery())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get api group resources")
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	var gvk schema.GroupVersionKind
	switch strings.ToLower(kind) {
	case "configmap":
		gvk = schema.FromAPIVersionAndKind("v1", "ConfigMap")
	case "secret":
		gvk = schema.FromAPIVersionAndKind("v1", "Secret")
	case "namespace":
		gvk = schema.FromAPIVersionAndKind("v1", "Namespace")
	case "deployment":
		gvk = schema.FromAPIVersionAndKind("apps/v1", "Deployment")
	case "service":
		gvk = schema.FromAPIVersionAndKind("v1", "Service")
	case "ingress":
		gvk = schema.FromAPIVersionAndKind("networking.k8s.io/v1", "Ingress")
	case "pod":
		gvk = schema.FromAPIVersionAndKind("v1", "Pod")
	case "job":
		gvk = schema.FromAPIVersionAndKind("batch/v1", "Job")
	default:
		return nil, errors.Wrapf(tclerrors.ErrRuntime, "i don't know group kind")
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get rest mapping")
	}

	var dri dynamic.ResourceInterface
	if namespace == "" {
		dri = k.dynamicClient.Resource(mapping.Resource)
	} else {
		dri = k.dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	}

	return dri, nil
}

func (k *client) GetResource(ctx context.Context, kind, name, namespace string) (*unstructured.Unstructured, error) {
	dri, err := k.getResourceInterface(kind, namespace)
	if err != nil {
		return nil, err
	}

	object, err := dri.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get. kind=%s, name=%s, namespace=%s", kind, name, namespace)
	}

	return object, nil
}
