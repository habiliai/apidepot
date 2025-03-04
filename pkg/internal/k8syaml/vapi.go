package k8syaml

import (
	"context"
	_ "embed"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/pkg/errors"
	"strings"
)

//go:embed data/vapi/_vapi_main.ts
var vapiMainFileContents string

type (
	VapiYamlValues struct {
		*domain.VapiRelease
		TarFileUrl string
		EnvVars    map[string]string
	}

	CustomVapiYamlValues struct {
		*domain.CustomVapi
		TarFileUrl string
		EnvVars    map[string]string
	}
)

func (s *Service) GetVapiYamlValues(
	ctx context.Context,
	vapiReleases []domain.VapiRelease,
	vapiEnvVars []domain.StackVapiEnvVar,
) ([]VapiYamlValues, error) {
	vapis := make([]VapiYamlValues, 0, len(vapiReleases))

	envVars := map[string][]domain.StackVapiEnvVar{}
	for _, vapiEnvVar := range vapiEnvVars {
		vapiName, rest := util.SplitStringToPair(vapiEnvVar.Name, ".")
		if rest == "" {
			return nil, errors.Errorf("invalid vapi env var name: %s", vapiEnvVar.Name)
		}
		envVars[vapiName] = append(envVars[vapiName], vapiEnvVar)
	}

	for _, vapiRelease := range vapiReleases {
		vapiEnvVars, ok := envVars[vapiRelease.Package.Name]
		if !ok {
			vapiEnvVars = []domain.StackVapiEnvVar{}
		}

		envVars := make(map[string]string, len(vapiEnvVars)+len(vapiRelease.EnvVars))
		for _, envVar := range vapiRelease.EnvVars {
			envVars[envVar.Name] = envVar.Default
		}

		for _, envVar := range vapiEnvVars {
			vapiName, name := util.SplitStringToPair(envVar.Name, ".")
			if vapiName != vapiRelease.Package.Name {
				continue
			}
			envVars[name] = envVar.Value
		}

		tarFileUrl, err := s.getPackageTarFileUrl(ctx, constants.VapiBucketId, vapiRelease.Published, vapiRelease.TarFilePath)
		if err != nil {
			return nil, err
		}

		vapis = append(vapis, VapiYamlValues{
			VapiRelease: &vapiRelease,
			TarFileUrl:  tarFileUrl,
			EnvVars:     envVars,
		})
	}

	return vapis, nil
}

func (s *Service) GetCustomVapiYamlValues(
	ctx context.Context,
	customVapis []domain.CustomVapi,
	vapiEnvVars []domain.StackVapiEnvVar,
) ([]CustomVapiYamlValues, error) {
	result := make([]CustomVapiYamlValues, 0, len(customVapis))

	envVars := map[string][]domain.StackVapiEnvVar{}
	for _, vapiEnvVar := range vapiEnvVars {
		vapiName, rest := util.SplitStringToPair(vapiEnvVar.Name, ".")
		if rest == "" {
			return nil, errors.Errorf("invalid vapi env var name: %s", vapiEnvVar.Name)
		}
		envVars[vapiName] = append(envVars[vapiName], vapiEnvVar)
	}

	for _, customVapi := range customVapis {
		vapiEnvVars, ok := envVars[customVapi.Name]
		if !ok {
			vapiEnvVars = []domain.StackVapiEnvVar{}
		}

		envVars := make(map[string]string, len(vapiEnvVars))
		for _, envVar := range vapiEnvVars {
			vapiName, name := util.SplitStringToPair(envVar.Name, ".")
			if vapiName != customVapi.Name {
				continue
			}
			envVars[name] = envVar.Value
		}

		tarFileUrl, err := s.getPackageTarFileUrl(ctx, constants.CustomVapiBucketId, false, customVapi.TarFilePath)
		if err != nil {
			return nil, err
		}

		result = append(result, CustomVapiYamlValues{
			CustomVapi: &customVapi,
			TarFileUrl: tarFileUrl,
			EnvVars:    envVars,
		})
	}

	return result, nil
}

func (s *Service) getPackageTarFileUrl(
	ctx context.Context,
	bucketId string,
	public bool,
	tarFilePath string,
) (string, error) {
	if public {
		resp := s.storage.GetPublicUrl(bucketId, tarFilePath)
		return resp.SignedURL, nil
	}

	resp, err := s.storage.CreateSignedUrl(ctx, bucketId, tarFilePath, 120)
	logger.Debug("create signed url", "bucketId", bucketId, "tarFilePath", tarFilePath, "public", public)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create signed url")
	}

	url := resp.SignedURL
	if s.shapleEnv == digo.EnvTest {
		url = strings.Replace(url, s.storageUrl, "http://storage-test.default.svc.cluster.local:5000", 1)
	}

	return url, nil
}
