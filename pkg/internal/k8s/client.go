package k8s

import (
	"context"
	"github.com/habiliai/apidepot/pkg/config"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type (
	Options struct {
		Force        bool
		ApplyCheckFn func(ctx context.Context) error
	}
	OptionsFunc func(*Options)
	Client      interface {
		ApplyYamlFile(ctx context.Context, yamlFileContents string) error
		DeleteYamlFile(ctx context.Context, yamlFileContents string, wait bool, options ...OptionsFunc) error
		Apply(ctx context.Context, objects []unstructured.Unstructured) error
		Delete(ctx context.Context, objects []unstructured.Unstructured, wait bool, options ...OptionsFunc) error
		GetResource(ctx context.Context, kind, name, namespace string) (*unstructured.Unstructured, error)
		Wait(ctx context.Context, kind, namespace string, selector string, forCondition string) error
		GetLogs(
			ctx context.Context,
			selector, namespace string,
		) (map[string]string, error)
		Upgrade(
			ctx context.Context,
			oldObjects []unstructured.Unstructured,
			newObjects []unstructured.Unstructured,
			options ...OptionsFunc,
		) error
	}

	client struct {
		client        *kubernetes.Clientset
		dynamicClient *dynamic.DynamicClient
	}
)

var (
	_ Client = (*client)(nil)
)

func newK8sClient(
	conf config.KubernetesConfig,
	region tcltypes.InstanceZone,
) (Client, error) {
	regionalConf := conf.GetRegionalConfig(region)

	kubeConfig := conf.KubeConfig
	if kubeConfig == "" {
		homeDir := homedir.HomeDir()
		kubeConfig = homeDir + "/.kube/config"
	}

	var (
		config *rest.Config
		err    error
	)
	if kubeConfig == "inCluster" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfig},
			&clientcmd.ConfigOverrides{
				CurrentContext: regionalConf.Context,
			}).ClientConfig()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build config from flags")
	}
	if conf.Burst > 0 {
		config.Burst = conf.Burst
	}
	if conf.QPS > 0 {
		config.QPS = conf.QPS
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create client from config")
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create dynamic client from config")
	}

	return &client{
		client:        k8sClient,
		dynamicClient: dynamicClient,
	}, nil
}

func WithForce(force bool) OptionsFunc {
	return func(options *Options) {
		options.Force = force
	}
}

func WithApplyCheckFn(fn func(ctx context.Context) error) OptionsFunc {
	return func(options *Options) {
		options.ApplyCheckFn = fn
	}
}

func mergeK8sClientOptions(options ...OptionsFunc) Options {
	opts := Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}
