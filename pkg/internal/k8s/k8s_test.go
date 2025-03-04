package k8s_test

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/k8s"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"testing"
	"time"
)

type K8sClientTestSuite struct {
	suite.Suite

	k k8s.Client
}

func (s *K8sClientTestSuite) SetupTest() {
	initCtx := digo.NewContainer(context.Background(), digo.EnvTest, nil)
	k8sPool := digo.MustGet[*k8s.ClientPool](initCtx, k8s.ServiceKeyK8sClientPool)
	var err error
	s.k, err = k8sPool.GetClient(tcltypes.InstanceZoneDefault)
	s.Require().NoError(err)
}

func TestK8sClient(t *testing.T) {
	suite.Run(t, new(K8sClientTestSuite))
}

func (s *K8sClientTestSuite) TestWait() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	contents, err := os.ReadFile("testdata/busy_deployment.yaml")
	s.Require().NoError(err)

	s.NoError(s.k.ApplyYamlFile(ctx, string(contents)))
	defer s.k.DeleteYamlFile(ctx, string(contents), true, k8s.WithForce(true))
	time.Sleep(1 * time.Second)

	s.NoError(s.k.Wait(ctx,
		"pod",
		"default",
		"app=busy",
		"ready",
	))
}

func (s *K8sClientTestSuite) TestGivenNoSetBurstAndQPS_WhenManyCallK8sIn15Seconds_ShouldBeError() {
	if os.Getenv("CI") != "" {
		s.T().Skipf("This test is skipped because it affects the local k8s cluster which has limited resources when running whole of tests.")
	}

	k := s.k

	contents, err := os.ReadFile("testdata/burst_deployment.yaml")
	s.Require().NoError(err)

	defer k.DeleteYamlFile(context.Background(), string(contents), true, k8s.WithForce(true))

	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, tclerrors.ErrTimeout)
	defer cancel()

	calls := func() error {
		for i := 0; i < 100; i++ {
			if err := k.ApplyYamlFile(ctx, string(contents)); err != nil {
				return err
			}
			if err := k.DeleteYamlFile(ctx, string(contents), true, k8s.WithForce(true)); err != nil {
				return err
			}
		}

		return nil
	}

	s.ErrorAs(calls(), &tclerrors.ErrTimeout, "should be error on timeout")
}

func (s *K8sClientTestSuite) TestGivenCustomBurstAndQPS_WhenManyCallK8sIn15Seconds_ShouldBeOk() {
	if os.Getenv("CI") != "" {
		s.T().Skipf("This test is skipped because it affects the local k8s cluster which has limited resources when running whole of tests.")
	}

	contents, err := os.ReadFile("testdata/burst_deployment.yaml")
	s.Require().NoError(err)

	defer s.k.DeleteYamlFile(context.Background(), string(contents), true, k8s.WithForce(true))

	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, tclerrors.ErrTimeout)
	defer cancel()

	startTime := time.Now()
	for i := 0; i < 15; i++ {
		s.Require().NoError(s.k.ApplyYamlFile(ctx, string(contents)))
		s.Require().NoError(s.k.DeleteYamlFile(ctx, string(contents), true, k8s.WithForce(true)))
		duration := time.Now().Sub(startTime)
		s.T().Logf("accumulated time=%s", duration.String())
	}
}

func (s *K8sClientTestSuite) TestGivenOldAndNewK8sYaml_WhenDiffK8sObjectsAndUpgrade_ShouldBeOk() {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	// given
	oldYaml, err := os.ReadFile("testdata/old.yaml")
	s.Require().NoError(err)

	newYaml, err := os.ReadFile("testdata/new.yaml")
	s.Require().NoError(err)

	oldObjects, err := k8syaml.ParseK8sYaml(string(oldYaml))
	s.Require().NoError(err)

	newObjects, err := k8syaml.ParseK8sYaml(string(newYaml))
	s.Require().NoError(err)

	// when
	_, _, markings, err := k8syaml.DiffK8sObjects(oldObjects, newObjects)

	s.Require().NoError(err)
	s.Require().Len(markings, 4)

	defer s.k.Delete(ctx, oldObjects, true, k8s.WithForce(true))
	s.Require().NoError(s.k.Apply(ctx, oldObjects))
	time.Sleep(250 * time.Millisecond)
	s.Require().NoError(s.k.Wait(ctx, "pod", "default", "app=testdata", "ready"))

	defer s.k.Delete(ctx, newObjects, true, k8s.WithForce(true))
	err = s.k.Upgrade(ctx, oldObjects, newObjects, k8s.WithApplyCheckFn(func(ctx context.Context) error {
		ctx, cancel := context.WithTimeoutCause(ctx, 3*time.Second, errors.Errorf("timeout when upgrade in k8s test"))
		defer cancel()

		return s.k.Wait(ctx, "pod", "default", "app=testdata", "ready")
	}))

	// then
	s.Require().NoError(err)

	{
		s.T().Logf("check old configmap are deleted")
		_, err := s.k.GetResource(ctx, "configmap", "testdata", "default")
		s.Require().Error(err)
	}
	{
		s.T().Logf("check new configmap are created")
		obj, err := s.k.GetResource(ctx, "configmap", "testdata-abc", "default")
		s.Require().NoError(err)
		s.Require().NotNil(obj)
		value, ok, err := unstructured.NestedString(obj.Object, "data", "username")
		s.Require().NoError(err)
		s.Require().True(ok)
		s.Require().Equal("elon", value)
	}
	{
		s.T().Logf("check new serviec are changed instead of old service")
		obj, err := s.k.GetResource(ctx, "service", "testdata", "default")
		s.Require().NoError(err)
		s.Require().NotNil(obj)
		ports, ok, err := unstructured.NestedSlice(obj.Object, "spec", "ports")
		s.Require().NoError(err)
		s.Require().True(ok)
		s.Require().Len(ports, 1)

		port, ok, err := unstructured.NestedInt64(ports[0].(map[string]any), "port")
		s.Require().NoError(err)
		s.Require().True(ok)
		s.Require().Equal(int64(12323), port)
	}
}
