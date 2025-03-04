package constants

const (
	PathStorage        = "/storage/v1"
	PathPostgrest      = "/postgrest/v1"
	PathAuth           = "/auth/v1"
	PathVapis          = "/vapis/v1"
	PathPostgrestReady = PathPostgrest + "/ready"
	PathPostgrestLive  = PathPostgrest + "/live"
	PathStorageHealth  = PathStorage + "/health"
	PathAuthHealth     = PathAuth + "/health"
)
