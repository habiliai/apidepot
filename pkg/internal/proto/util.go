package proto

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/pkg/errors"
	"net/url"
)

func (s *apiDepotServer) CreateImageUploadUrl(ctx context.Context, req *CreateImageUploadUrlRequest) (*CreateImageUploadUrlResponse, error) {
	if _, err := s.userService.GetUser(ctx); err != nil {
		return nil, err
	}

	var ext string
	switch req.ImageType {
	case constants.ImageTypePNG:
		ext = ".png"
	case constants.ImageTypeJPG:
		ext = ".jpg"
	default:
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "unsupported image type: %s", req.ImageType)
	}

	imageId := uuid.NewString()
	filepath := fmt.Sprintf("images/%s%s", imageId, ext)

	var bucket string
	switch req.Category {
	case CreateImageUploadUrlRequest_STACK_LOGO:
		bucket = constants.StackLogoBucketId
	case CreateImageUploadUrlRequest_USER_PROFILE_IMAGE:
		bucket = constants.ProfileImageBucketId
	case CreateImageUploadUrlRequest_TAPP_BANNER_IMAGE:
		bucket = constants.TappBannerImageBucketId
	case CreateImageUploadUrlRequest_TAPP_SCREENSHOT_IMAGE:
		bucket = constants.TappScreenshotImageBucketId
	default:
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "unsupported category: %s", req.Category)
	}

	resp, err := s.storageClient.CreateSignedUploadUrl(ctx, bucket, filepath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create signed upload url")
	}

	signedUrl := resp.Url
	var token string
	if url, err := url.Parse(resp.Url); err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", resp.Url)
	} else if !url.Query().Has("token") {
		return nil, errors.Errorf("token is not found in url: %s", resp.Url)
	} else {
		token = url.Query().Get("token")
	}

	return &CreateImageUploadUrlResponse{
		Path:      filepath,
		SignedUrl: signedUrl,
		Token:     token,
		Bucket:    bucket,
	}, nil
}
