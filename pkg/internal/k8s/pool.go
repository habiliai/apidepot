package k8s

import (
	"github.com/habiliai/apidepot/pkg/config"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
)

type ClientPool struct {
	clients map[tcltypes.InstanceZone]Client
}

func (p *ClientPool) GetClient(zone tcltypes.InstanceZone) (Client, error) {
	if zone == tcltypes.InstanceZoneDefault {
		zone = tcltypes.InstanceZoneOciApSeoul
	} else if zone == tcltypes.InstanceZoneMulti {
		return nil, errors.Errorf("zone %s is not supported", zone)
	}

	client, ok := p.clients[zone]
	if !ok {
		return nil, errors.Errorf("client not found for zone %s", zone)
	}

	return client, nil
}

func newK8sClientPool(
	conf config.KubernetesConfig,
	regions []tcltypes.InstanceZone,
) (*ClientPool, error) {
	clients := map[tcltypes.InstanceZone]Client{}

	for _, region := range regions {
		client, err := newK8sClient(conf, region)
		if err != nil {
			return nil, err
		}
		clients[region] = client
	}

	return &ClientPool{
		clients: clients,
	}, nil
}
