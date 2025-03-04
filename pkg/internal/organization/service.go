package organization

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
)

type (
	Service interface {
		UpdateOrganization(ctx context.Context, input CreateOrUpdateOrganizationInput) (uint, error)
		GetOrganizationById(ctx context.Context, id uint) (domain.Organization, error)
		DeleteOrganization(ctx context.Context, id uint) error
		GetOrganizations(ctx context.Context, memberOwnerId *string) ([]domain.Organization, error)
	}
	service struct {
	}
)

const (
	ServiceKey = "organization.Service"
)

var (
	_      Service = (*service)(nil)
	logger         = tclog.GetLogger()
)

func init() {
	digo.ProvideService(ServiceKey, func(serviceContainer *digo.Container) (interface{}, error) {
		return &service{}, nil
	})
}
