package stack_test

import (
	pkgconfig "github.com/habiliai/apidepot/pkg/config"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/mokiat/gog"
	"io"
	"net/http"
	"os"
	"strings"
)

func (s *StackServiceTestSuite) TestCreateCustomVapi() {
	githubToken := os.Getenv("YOUR_GITHUB_TOKEN")
	if githubToken == "" {
		s.T().Skip("github token is missing")
	}

	ctx := helpers.WithGithubToken(s, githubToken)

	st, err := s.stackService.CreateStack(ctx, stack.CreateStackInput{
		ProjectID:     s.project.ID,
		Name:          "test-stack",
		SiteURL:       "localhost:3000",
		DefaultRegion: tcltypes.InstanceZoneDefault,
		GitRepo:       "habiliai/service-template-example",
		GitBranch:     "main",
	})
	s.Require().NoError(err)

	s.Require().NoError(s.stackService.EnableOrUpdateAuth(ctx, st.ID, stack.EnableOrUpdateAuthInput{
		AuthInput: stack.AuthInput{
			MailerAutoConfirm: gog.PtrOf(true),
		},
	}, true))
	defer s.stackService.DisableAuth(ctx, st.ID)

	customVapi, err := s.stackService.EnableCustomVapi(
		ctx,
		st.ID,
		stack.CreateCustomVapiInput{
			Name: "my-custom-vapi",
		},
	)
	s.Require().NoError(err)
	s.Require().NotNil(customVapi)

	{
		url := s.storageClient.GetPublicUrl(constants.CustomVapiBucketId, customVapi.TarFilePath).SignedURL
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		s.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()
		contents, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)

		s.Greater(len(contents), 0)
	}

	s.T().Logf("CustomVapi: %+v", customVapi)

	{
		st, err = s.stackService.GetStack(ctx, st.ID)
		s.Require().NoError(err)

		values := s.k8syamlService.NewValuesFromStack(st).WithAuth(pkgconfig.SMTPConfig{
			Host:       "smtp.gmail.com",
			Port:       587,
			Username:   "test",
			Password:   "test",
			AdminEmail: "dennis@habili.ai",
		})
		customVapiYamlValues, err := s.k8syamlService.GetCustomVapiYamlValues(ctx, st.CustomVapis, nil)
		s.Require().NoError(err)
		values = values.WithCustomVapis(customVapiYamlValues)

		s.T().Logf("values: %+v", values)

		k8sYamlFiles := []string{
			"common/network-policy.yaml",
			"common/ingress.yaml",
			"database/configmap.yaml",
			"database/secret.yaml",
			"auth/configmap.yaml", "auth/secret.yaml", "auth/deployment.yaml", "auth/service.yaml",
			"custom-vapi/configmap.yaml",
			"custom-vapi/secret.yaml",
			"custom-vapi/service.yaml",
			"custom-vapi/deployment.yaml",
		}

		yamlResult, err := s.k8syamlService.RenderYaml(k8sYamlFiles, values)
		s.Require().NoError(err)
		os.WriteFile("custom_vapi_test.yaml", []byte(yamlResult), 0644)

		s.T().Logf("YAML: %s", yamlResult)
		s.True(strings.Contains(yamlResult, "shaple.io/component: custom-vapi"))
	}
}
