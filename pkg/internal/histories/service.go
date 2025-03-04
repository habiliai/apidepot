package histories

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"time"
)

var (
	ServiceKey digo.ObjectKey = "habili.instance-history"
	_          Service        = (*service)(nil)
)

type Service interface {
	GetTotalRunningTimeForUser(ctx context.Context, begin, end *time.Time, userId uint) ([]InstanceRunningTime, error)
	WriteInstanceHistoriesAt(ctx context.Context) error
	WriteStackHistoriesAt(ctx context.Context) error
}

type service struct {
	bucketService services.BucketService
}

func init() {
	digo.ProvideService(ServiceKey, func(container *digo.Container) (any, error) {
		bucketService, err := digo.Get[services.BucketService](container, services.ServiceKeyBucketService)
		if err != nil {
			return nil, err
		}

		return &service{
			bucketService: bucketService,
		}, nil
	})
}
