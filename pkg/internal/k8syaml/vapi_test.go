package k8syaml_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
)

func (s *K8sYamlServiceTestSuite) TestK8sYamlService_RenderYamlWithVapi() {
	stack := domain.Stack{
		Model: domain.Model{
			ID: 1,
		},
		Hash: "iktjke1233",
		Name: "dev",
		Project: domain.Project{
			Model: domain.Model{
				ID: 1,
			},
			Name: "test123",
		},
		Vapis: []domain.StackVapi{
			{
				VapiID:  1,
				StackID: 1,
				Vapi: domain.VapiRelease{
					Model: domain.Model{
						ID: 1,
					},
					PackageID: 1,
					Package: domain.VapiPackage{
						Model: domain.Model{
							ID: 1,
						},
						Name: "test",
					},
					Version:     "1.0.0",
					Published:   true,
					TarFilePath: "prj/test/v1.0.0.tar",
				},
			},
		},
	}

	yamlFile, err := s.k8sYamlService.RenderYaml(
		[]string{
			"vapi/configmap.yaml",
			"vapi/deployment.yaml",
			"vapi/secret.yaml",
			"vapi/service.yaml",
			"common/ingress.yaml",
		},
		s.k8sYamlService.NewValuesFromStack(&stack).WithVapis([]k8syaml.VapiYamlValues{
			{
				VapiRelease: &stack.Vapis[0].Vapi,
				TarFileUrl:  "http://localhost:8080/prj/test/v1.0.0.tar",
				EnvVars: map[string]string{
					"TEST":     "test",
					"TEST_T":   "test",
					"TEST_T_T": "test",
				},
			},
			{
				VapiRelease: &stack.Vapis[0].Vapi,
				TarFileUrl:  "http://localhost:8080/prj/test/v1.0.0.tar",
				EnvVars: map[string]string{
					"TEST1":     "test1",
					"TEST1_T":   "test1",
					"TEST1_T_T": "test1",
				},
			},
		}),
	)
	s.Require().NoError(err)

	s.T().Logf("yamlFile: %s", yamlFile)
}

func (s *K8sYamlServiceTestSuite) TestGetVapiYamlValues() {
	s.Run("given valid input, when GetVapiYamlValues is called, should be OK", func() {
		// given
		vapiReleases := []domain.VapiRelease{
			{
				Model: domain.Model{ID: 1},
				Package: domain.VapiPackage{
					Model: domain.Model{ID: 1},
					Name:  "test",
				},
				Version:     "1.0.0",
				Published:   true,
				TarFilePath: "prj/test/v1.0.0.tar",
			},
		}
		vapiEnvVars := []domain.StackVapiEnvVar{
			{Name: "test.TEST_KEY", Value: "test_value"},
		}

		// when
		vapiYamlValues, err := s.k8sYamlService.GetVapiYamlValues(s, vapiReleases, vapiEnvVars)

		// then
		s.Require().NoError(err)
		s.Require().Len(vapiYamlValues, 1)
		s.Require().Equal("test_value", vapiYamlValues[0].EnvVars["TEST_KEY"])
	})

	s.Run("given invalid key name in env var, when GetVapiYamlValues is called, should return error", func() {
		// given
		vapiReleases := []domain.VapiRelease{
			{
				Model: domain.Model{ID: 1},
				Package: domain.VapiPackage{
					Model: domain.Model{ID: 1},
					Name:  "test",
				},
				Version:     "1.0.0",
				Published:   true,
				TarFilePath: "prj/test/v1.0.0.tar",
			},
		}
		vapiEnvVars := []domain.StackVapiEnvVar{
			{Name: "invalid_key", Value: "test_value"},
		}

		// when
		_, err := s.k8sYamlService.GetVapiYamlValues(s, vapiReleases, vapiEnvVars)

		// then
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "invalid vapi env var name")
	})
}
