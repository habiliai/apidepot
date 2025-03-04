package k8syaml_test

import (
	"github.com/goccy/go-yaml"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	"strings"
)

func (s *K8sYamlServiceTestSuite) TestParseK8sYaml() {
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

	yamlFile, err := s.k8sYamlService.RenderYaml([]string{"common/namespace.yaml"}, values)
	s.NoError(err)

	path, err := yaml.PathString("$.metadata.name")
	s.NoError(err)

	var name string
	s.NoError(path.Read(strings.NewReader(yamlFile), &name))

	s.Equal(stack.Namespace(), name)

	objects, err := k8syaml.ParseK8sYaml(yamlFile)
	s.Require().NoError(err)

	s.Len(objects, 1)
	s.Equal("Namespace", objects[0].GroupVersionKind().Kind)
}
