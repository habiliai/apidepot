package constants

import "github.com/docker/go-units"

const (
	DefaultImageMaxSize         = 10 * units.MB
	ProfileImageBucketId        = "profile-images"
	ProfileImageMaxSize         = DefaultImageMaxSize
	StackLogoBucketId           = "stack-logos"
	StackLogoMaxSize            = DefaultImageMaxSize
	TappBannerImageBucketId     = "tapp-banner-images"
	TappIconImageBucketId       = "tapp-icon-images"
	TappScreenshotImageBucketId = "tapp-screenshot-images"
)
