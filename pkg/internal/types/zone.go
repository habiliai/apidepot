package tcltypes

import "github.com/habiliai/apidepot/pkg/internal/constants"

type (
	InstanceZone string
)

const (
	InstanceZoneMulti        InstanceZone = "multi"
	InstanceZoneOciApSeoul   InstanceZone = "oci-ap-seoul-1"
	InstanceZoneOciSingapore InstanceZone = "oci-ap-singapore-1"
	InstanceZoneDefault                   = InstanceZoneOciApSeoul
)

var InstanceZones = []InstanceZone{
	InstanceZoneOciApSeoul,
	InstanceZoneOciSingapore,
}

func (z InstanceZone) ToS3Region() string {
	switch z {
	case InstanceZoneOciApSeoul:
		return constants.S3_REGION_SEOUL
	case InstanceZoneOciSingapore:
		return constants.S3_REGION_SINGAPORE
	default:
		panic("invalid instance zone")
	}
}

func (z InstanceZone) String() string {
	return string(z)
}
