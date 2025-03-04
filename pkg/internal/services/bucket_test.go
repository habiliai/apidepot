package services_test

import (
	"context"
	"fmt"
	"github.com/Masterminds/goutils"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/services"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"strings"
	"time"
)

func (s *ServicesTestSuite) TestBucketService() {
	randomName, err := goutils.CryptoRandomAlphabetic(12)
	s.Require().NoError(err)

	randomName = strings.ToLower(randomName)

	zone := tcltypes.InstanceZoneDefault
	bucketName := fmt.Sprintf("test-%s", randomName)
	ctx := context.TODO()
	container := digo.NewContainer(ctx, digo.EnvTest, nil)
	bs, err := digo.Get[services.BucketService](container, services.ServiceKeyBucketService)
	s.NoError(err)

	s.T().Logf("bucketName: %s", bucketName)

	s.Run("create bucket", func() {
		s.NoError(bs.CreateBucket(ctx, zone, bucketName))

		resp, err := bs.IsBucketExists(ctx, zone, bucketName)
		s.NoError(err)

		s.True(resp)
	})

	// wait for bucket to be created
	time.Sleep(5 * time.Second)

	s.Run("delete bucket", func() {
		s.NoError(bs.DeleteBucket(ctx, zone, bucketName))

		resp, err := bs.IsBucketExists(ctx, zone, bucketName)
		s.NoError(err)
		s.False(resp)
	})
}

func (s *ServicesTestSuite) TestBucketServiceGetTotalSize() {
	randomName, err := goutils.CryptoRandomAlphabetic(12)
	s.Require().NoError(err)

	randomName = strings.ToLower(randomName)

	zone := tcltypes.InstanceZoneDefault
	bucketName := fmt.Sprintf("test-%s", randomName)
	ctx := context.TODO()
	container := digo.NewContainer(ctx, digo.EnvTest, nil)
	bs, err := digo.Get[services.BucketService](container, services.ServiceKeyBucketService)
	s.NoError(err)

	s.Require().NoError(bs.CreateBucket(ctx, zone, bucketName))

	minioClient, err := minio.New("minio.local.shaple.io", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
		Region: "ap-seoul-1",
	})
	s.Require().NoError(err)

	info, err := minioClient.PutObject(ctx, bucketName, "/README.txt", strings.NewReader("I am a habili.ai which is future of dev AI"), -1, minio.PutObjectOptions{})
	s.Require().NoError(err)
	info1, err := minioClient.PutObject(ctx, bucketName, "/test/README.txt", strings.NewReader("I am a habili.ai which is future of dev AI"), -1, minio.PutObjectOptions{})
	s.Require().NoError(err)
	info2, err := minioClient.PutObject(ctx, bucketName, "/t/README.txt", strings.NewReader("I am a habili.ai which is future of dev AI"), -1, minio.PutObjectOptions{})
	s.Require().NoError(err)
	info3, err := minioClient.PutObject(ctx, bucketName, "/tt/tt/README.txt", strings.NewReader("I am a habili.ai which is future of dev AI"), -1, minio.PutObjectOptions{})
	s.Require().NoError(err)
	info4, err := minioClient.PutObject(ctx, bucketName, "/dd/dd/README.txt", strings.NewReader("I am a habili.ai which is future of dev AI"), -1, minio.PutObjectOptions{})
	s.Require().NoError(err)

	totalSize, err := bs.GetTotalSize(ctx, zone, bucketName)
	s.Require().NoError(err)

	s.Equal(info.Size+info1.Size+info2.Size+info3.Size+info4.Size, totalSize)
}
