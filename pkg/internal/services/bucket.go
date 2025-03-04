package services

import (
	"context"
	"github.com/habiliai/apidepot/pkg/config"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"strings"
)

type BucketService interface {
	CreateBucket(ctx context.Context, zone tcltypes.InstanceZone, name string) error
	DeleteBucket(ctx context.Context, zone tcltypes.InstanceZone, name string) error
	IsBucketExists(ctx context.Context, zone tcltypes.InstanceZone, name string) (bool, error)
	GetTotalSize(
		ctx context.Context,
		zone tcltypes.InstanceZone, name string,
	) (totalSize int64, err error)
	PutObject(ctx context.Context, zone tcltypes.InstanceZone, bucket, path string, contents []byte) error
}

type bucketService struct {
	clients map[tcltypes.InstanceZone]*minio.Client // region -> client
}

func NewBucketService(
	s3Conf config.S3Config,
) (BucketService, error) {
	clients := make(map[tcltypes.InstanceZone]*minio.Client, len(tcltypes.InstanceZones))
	for _, zone := range tcltypes.InstanceZones {
		var (
			endpoint = s3Conf.GetRegionalConfig(zone).Endpoint
			secure   bool
		)

		if strings.HasPrefix(endpoint, "https://") {
			endpoint, _ = strings.CutPrefix(endpoint, "https://")
			secure = true
		} else if strings.HasPrefix(endpoint, "http://") {
			endpoint, _ = strings.CutPrefix(endpoint, "http://")
			secure = false
		} else {
			return nil, errors.Errorf("invalid endpoint: %s", endpoint)
		}

		minioClient, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(s3Conf.AccessKey, s3Conf.SecretKey, ""),
			Secure: secure,
			Region: zone.ToS3Region(),
		})
		logger.Debug("new minioClient", "s3Conf", s3Conf)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create minio client")
		}

		clients[zone] = minioClient
	}

	return &bucketService{
		clients: clients,
	}, nil
}

func (bs *bucketService) getClient(zone tcltypes.InstanceZone) (*minio.Client, error) {
	client, ok := bs.clients[zone]
	if !ok {
		return nil, errors.Errorf("invalid zone: %s", zone)
	}

	return client, nil
}

func (bs *bucketService) PutObject(
	ctx context.Context,
	zone tcltypes.InstanceZone,
	bucket string,
	path string,
	contents []byte,
) error {
	client, err := bs.getClient(zone)
	if err != nil {
		return err
	}

	_, err = client.PutObject(ctx, bucket, path, strings.NewReader(string(contents)), int64(len(contents)), minio.PutObjectOptions{})
	return errors.WithStack(err)
}

func (bs *bucketService) CreateBucket(
	ctx context.Context,
	zone tcltypes.InstanceZone,
	name string,
) error {
	client, err := bs.getClient(zone)
	if err != nil {
		return err
	}
	return errors.WithStack(client.MakeBucket(ctx, name, minio.MakeBucketOptions{
		Region: zone.ToS3Region(),
	}))
}

func (bs *bucketService) DeleteBucket(
	ctx context.Context,
	zone tcltypes.InstanceZone,
	name string,
) error {
	client, err := bs.getClient(zone)
	if err != nil {
		return err
	}
	return errors.WithStack(client.RemoveBucketWithOptions(ctx, name, minio.RemoveBucketOptions{
		ForceDelete: true,
	}))
}

func (bs *bucketService) IsBucketExists(
	ctx context.Context,
	zone tcltypes.InstanceZone,
	name string,
) (bool, error) {
	client, err := bs.getClient(zone)
	if err != nil {
		return false, err
	}
	exists, err := client.BucketExists(ctx, name)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check bucket")
	}

	return exists, nil
}

func (bs *bucketService) GetTotalSize(
	ctx context.Context,
	zone tcltypes.InstanceZone,
	name string,
) (totalSize int64, err error) {
	client, err := bs.getClient(zone)
	if err != nil {
		return 0, err
	}
	objectsCh := client.ListObjects(ctx, name, minio.ListObjectsOptions{
		Recursive: true,
	})

	for {
		select {
		case <-ctx.Done():
			err = errors.Wrapf(ctx.Err(), "context done")
			return
		case object, ok := <-objectsCh:
			if !ok {
				return
			}
			totalSize += object.Size
		}
	}
}

const ServiceKeyBucketService digo.ObjectKey = "bucketService"

func init() {
	digo.ProvideService(ServiceKeyBucketService, func(ctx *digo.Container) (any, error) {
		switch ctx.Env {
		case digo.EnvTest:
			return NewBucketService(config.S3Config{
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				Seoul: config.RegionalS3Config{
					Endpoint: "http://minio.local.shaple.io",
				},
				Singapore: config.RegionalS3Config{
					Endpoint: "http://minio.local.shaple.io",
				},
			})
		case digo.EnvProd:
			return NewBucketService(ctx.Config.S3)
		default:
			return nil, errors.Errorf("unsupported env: %s", ctx.Env)
		}
	})
}
