package stack

import (
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"time"
)

type ShapleServiceType string

const (
	ShapleServiceTypeAuth      ShapleServiceType = "auth"
	ShapleServiceTypeStorage   ShapleServiceType = "storage"
	ShapleServiceTypePostgrest ShapleServiceType = "postgrest"
	ShapleServiceTypeVapis     ShapleServiceType = "vapis"
)

type CreateStackInput struct {
	ProjectID     uint
	Name          string
	SiteURL       string
	Description   string
	LogoImageUrl  string
	DefaultRegion tcltypes.InstanceZone
	GitRepo       string
	GitBranch     string
}

type PatchStackInput struct {
	SiteURL      *string
	Name         *string
	Description  *string
	LogoImageUrl *string
}

type GetStatusOutput struct {
	HealthAuthService      bool `json:"health_auth_service"`
	HealthStorageService   bool `json:"health_storage_service"`
	HealthPostgrestService bool `json:"health_postgrest_service"`
	HealthVapiService      bool `json:"health_vapi_service"`
}

type Migration struct {
	Version time.Time `json:"version"` // format: yymmddHHMMSS
	Query   string    `json:"query"`
}

type MigrateDatabaseInput struct {
	Migrations []Migration `json:"migrations"`
}
