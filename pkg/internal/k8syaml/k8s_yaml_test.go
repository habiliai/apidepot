package k8syaml_test

import (
	"context"
	"github.com/goccy/go-yaml"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/k8s"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type K8sYamlServiceTestSuite struct {
	suite.Suite
	context.Context

	k8sYamlService *k8syaml.Service
	k8sClientPool  *k8s.ClientPool
}

func TestK8sYamlService(t *testing.T) {
	suite.Run(t, new(K8sYamlServiceTestSuite))
}

func (s *K8sYamlServiceTestSuite) SetupTest() {
	s.Context = context.TODO()
	container := digo.NewContainer(
		s,
		digo.EnvTest,
		nil,
	)

	s.k8sYamlService = digo.MustGet[*k8syaml.Service](container, k8syaml.ServiceKey)
	s.k8sClientPool = digo.MustGet[*k8s.ClientPool](container, k8s.ServiceKeyK8sClientPool)
}

func (s *K8sYamlServiceTestSuite) TestK8sYamlService_RenderYaml() {
	stack := domain.Stack{
		Hash: "iktjke1233",
		Name: "dev",
		Project: domain.Project{
			Name: "test123",
		},
	}

	values := struct {
		domain.Stack
	}{stack}

	object, err := s.k8sYamlService.RenderYaml([]string{"common/namespace.yaml"}, values)
	s.NoError(err)

	path, err := yaml.PathString("$.metadata.name")
	s.NoError(err)

	var name string
	s.NoError(path.Read(strings.NewReader(object), &name))

	s.Equal(stack.Namespace(), name)
}
