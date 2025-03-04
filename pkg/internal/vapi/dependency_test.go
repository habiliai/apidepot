package vapi_test

import (
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"gopkg.in/yaml.v3"
	"strings"
)

func (s *VapiTestSuite) TestParseDependencies() {
	ymlData := `
dependencies:
  dep1: "1.0.0"
  dep2:
    version: "1.2.0"
`

	var vapiYaml map[string]interface{}
	s.Require().NoError(yaml.Unmarshal([]byte(strings.TrimSpace(ymlData)), &vapiYaml))

	deps, ok := vapiYaml["dependencies"].(map[string]interface{})
	s.Require().True(ok)

	dependencies, err := vapi.ParseDependencies(deps)
	s.Require().NoError(err)

	s.Require().Len(dependencies, 2)
	visited := map[string]bool{
		"dep1": false,
		"dep2": false,
	}
	for _, dep := range dependencies {
		visited[dep.Name] = true
		if dep.Name == "dep1" {
			s.Equal("1.0.0", dep.Version)
		} else if dep.Name == "dep2" {
			s.Equal("1.2.0", dep.Version)
		} else {
			s.Failf("unexpected dependency", "dep name: %s", dep.Name)
		}
	}

	for dep, v := range visited {
		s.Truef(v, "%s is not visited", dep)
	}
}
