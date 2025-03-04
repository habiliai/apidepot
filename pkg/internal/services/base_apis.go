package services

import (
	"context"
	"github.com/docker/go-units"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/pkg/errors"
	"github.com/supabase-community/gotrue-go"
)

const (
	ServiceKeyGoTrueClient  = "gotrueClient"
	ServiceKeyStorageClient = "storageClient"
)

func init() {
	digo.ProvideService(ServiceKeyGoTrueClient, func(ctx *digo.Container) (interface{}, error) {

		switch ctx.Env {
		case digo.EnvProd:
			anonKey := ctx.Config.Stoa.AnonKey
			adminKey := ctx.Config.Stoa.AdminKey
			return gotrue.New("apidepot", anonKey).WithCustomGoTrueURL(ctx.Config.Stoa.URL + "/auth/v1").WithToken(adminKey), nil
		case digo.EnvTest:
			anonKey := constants.GotrueAnonKeyForTest
			adminKey := constants.GotrueAdminKeyForTest
			return gotrue.
				New("test", anonKey).
				WithCustomGoTrueURL("http://apidepot-test.local.shaple.io/auth/v1").
				WithToken(adminKey), nil
		}

		return nil, errors.Errorf("unknown env")
	})

	digo.ProvideService(ServiceKeyStorageClient, func(ctx *digo.Container) (interface{}, error) {
		var storageClient *storage.Client

		buckets := []struct {
			name    string
			maxSize float64
			public  bool
		}{
			{constants.VapiBucketId, constants.VapiPackageTarMaxSize, true},
			{constants.ProfileImageBucketId, constants.ProfileImageMaxSize, true},
			{constants.StackLogoBucketId, constants.StackLogoMaxSize, true},
			{constants.CustomVapiBucketId, constants.VapiPackageTarMaxSize, false},
		}

		switch ctx.Env {
		case digo.EnvProd:
			adminKey := ctx.Config.Stoa.AdminKey
			storageClient = storage.NewClient(ctx.Config.Stoa.URL+"/storage/v1", adminKey, nil)
		case digo.EnvTest:
			adminKey := constants.GotrueAdminKeyForTest
			storageClient = storage.NewClient("http://apidepot-test.local.shaple.io/storage/v1", adminKey, nil)

			for _, bucket := range buckets {
				if _, err := storageClient.EmptyBucket(ctx.Context, bucket.name); err != nil {
					logger.Warn("failed to empty bucket", "bucketName", bucket.name, tclog.Err(err))
				}
				if _, err := storageClient.DeleteBucket(ctx.Context, bucket.name); err != nil {
					logger.Warn("failed to delete bucket", "bucketName", bucket.name, tclog.Err(err))
				}
				logger.Info("deleted bucket", "bucketName", bucket.name)
			}

		default:
			return nil, errors.Errorf("unknown env")
		}

		for _, bucket := range buckets {
			if err := maybeCreateBucket(ctx.Context, storageClient, bucket.name, storage.BucketOptions{
				Public:        bucket.public,
				FileSizeLimit: units.HumanSize(bucket.maxSize),
			}); err != nil {
				return nil, err
			}
		}

		return storageClient, nil
	})
}

func maybeCreateBucket(ctx context.Context, storageClient *storage.Client, bucketId string, options storage.BucketOptions) error {
	if _, err := storageClient.CreateBucket(ctx, bucketId, options); err != nil {
		var storageErr *storage.StorageError
		if !errors.As(err, &storageErr) {
			return errors.Wrapf(err, "failed to create bucket: %s", constants.VapiBucketId)
		}

		if storageErr.Status != 200 && storageErr.Status != 201 && storageErr.Status != 409 {
			return errors.Wrapf(
				err,
				"failed to create bucket: %s, message: %s, status: %d",
				constants.VapiBucketId,
				storageErr.Message,
				storageErr.Status,
			)
		}

		if storageErr.Status == 409 {
			logger.Warn("existing bucket", "bucketName", constants.VapiBucketId, tclog.Err(err))
		}
	}

	return nil
}
