package k8s

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	"github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"strings"
	"sync"
	"time"
)

func (k *client) DeleteYamlFile(
	ctx context.Context,
	yamlFileContents string,
	wait bool,
	options ...OptionsFunc,
) error {
	yamlFileContents = strings.TrimSpace(yamlFileContents)
	if yamlFileContents == "" || yamlFileContents == "---" {
		return nil
	}

	objects, err := k8syaml.ParseK8sYaml(yamlFileContents)
	if err != nil {
		return err
	}

	return k.Delete(ctx, objects, wait, options...)
}

func (k *client) Delete(
	ctx context.Context,
	objects []unstructured.Unstructured,
	wait bool,
	options ...OptionsFunc,
) error {
	option := mergeK8sClientOptions(options...)

	var wg sync.WaitGroup
	for i := 0; i < len(objects); i++ {
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

		dri := k.dynamicClient.Resource(mapping.Resource).Namespace(object.GetNamespace())
		deleteOptions := v1.DeleteOptions{
			// follow kubectl default option for deletion objects
			// see also, https://github.com/kubernetes/kubernetes/issues/59850
			PropagationPolicy: gog.PtrOf(v1.DeletePropagationForeground),
		}
		if option.Force {
			deleteOptions.GracePeriodSeconds = gog.PtrOf[int64](0)
		}
		if err := dri.Delete(ctx, object.GetName(), deleteOptions); err != nil {
			return errors.Wrapf(err, "failed to delete")
		}

		if wait {
			wg.Add(1)
			go func(dri dynamic.ResourceInterface) {
				doIt := func() bool {
					_, err := dri.Get(ctx, object.GetName(), v1.GetOptions{})
					return errors2.IsNotFound(err)
				}
				defer wg.Done()
				for deleted := doIt(); !deleted; {
					select {
					case <-ctx.Done():
						if err := ctx.Err(); err != nil {
							logger.Error("deletion go routing error", tclog.Err(errors.Wrapf(err, "context done")))
						}
						return
					case <-time.After(200 * time.Millisecond):
						deleted = doIt()
					}
				}
			}(dri)
		}
	}

	wg.Wait()

	return nil
}
