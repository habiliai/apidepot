package k8s

import (
	"github.com/habiliai/apidepot/pkg/config"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	logger = tclog.GetLogger()
)

const ServiceKeyK8sClientPool digo.ObjectKey = "k8sClientPool"

func init() {
	digo.ProvideService(ServiceKeyK8sClientPool, func(ctx *digo.Container) (any, error) {
		switch ctx.Env {
		case "prod":
			return newK8sClientPool(ctx.Config.K8s, []tcltypes.InstanceZone{tcltypes.InstanceZoneOciApSeoul, tcltypes.InstanceZoneOciSingapore})
		case "test":
			// uses only kind cluster for testing
			return newK8sClientPool(config.KubernetesConfig{
				KubeConfig: "",
				Burst:      20000,
				QPS:        10000,
				Seoul: config.RegionalKubernetesConfig{
					Context: "kind-apidepot",
				},
				Singapore: config.RegionalKubernetesConfig{
					Context: "kind-apidepot",
				},
			}, []tcltypes.InstanceZone{tcltypes.InstanceZoneDefault})
		default:
			return nil, errors.New("unknown env")
		}
	})
}
