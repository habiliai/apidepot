package k8syaml_test

import (
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewTemplate(t *testing.T) {
	var require = require.New(t)

	tmpl, err := k8syaml.NewTemplate()
	require.NoError(err)
	require.NotNil(tmpl)

	require.NotNil(tmpl.Lookup("common/namespace.yaml"))
}

func (s *K8sYamlServiceTestSuite) TestGivenRenderYamlFromProjectHasEmptyNameWhenApplyToK8sShouldBeOK() {
	type Stack struct {
		ID        uint
		Namespace string
		Name      string
	}
	type Project struct {
		ID   uint
		Name string
	}

	values := struct {
		Stack   Stack
		Project Project
	}{
		Stack: Stack{
			ID:        1,
			Namespace: "test",
			Name:      "sample test",
		},
		Project: Project{
			ID:   2,
			Name: "sample project 1",
		},
	}

	yamlFile, err := s.k8sYamlService.RenderYaml([]string{"common/namespace.yaml"}, values)
	s.Require().NoError(err)

	k, err := s.k8sClientPool.GetClient(tcltypes.InstanceZoneDefault)
	s.Require().NoError(err)

	err = k.ApplyYamlFile(s, yamlFile)
	s.Require().NoError(err)
}
