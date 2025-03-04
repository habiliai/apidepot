package constants

import "github.com/docker/go-units"

const (
	VapiBucketId          = "vapis"
	VapiPackageTarMaxSize = 25 * units.MB
	VapiYamlFileName      = "apidepot.yml"
	VapiHealthPath        = "/_internal/health"
	CustomVapiBucketId    = "custom-vapis"
)
