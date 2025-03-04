package k8syaml

import (
	"github.com/habiliai/apidepot/pkg/internal/digo"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/modern-go/reflect2"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

var logger = tclog.GetLogger()

type Service struct {
	tmpl *template.Template

	storage    *storage.Client
	shapleEnv  digo.Env
	storageUrl string
}

func NewK8sYamlService(
	shapleEnv digo.Env,
	storageClient *storage.Client,
	storageUrl string,
) (*Service, error) {
	tmpl, err := NewTemplate()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create templates")
	}

	return &Service{
		tmpl:       tmpl,
		shapleEnv:  shapleEnv,
		storage:    storageClient,
		storageUrl: storageUrl,
	}, nil
}

func (s *Service) RenderYaml(templateNames []string, data interface{}) (string, error) {
	logger.Debug("render yaml", "templateNames", templateNames)

	if reflect2.IsNil(data) {
		return "", errors.New("data is nil")
	}

	var sb strings.Builder
	for i, templateName := range templateNames {
		var yb strings.Builder
		logger.Debug("begin of execution template", "name", templateName)
		if err := s.tmpl.ExecuteTemplate(&yb, templateName, data); err != nil {
			return "", errors.Wrapf(err, "failed to execute template")
		}

		yaml := strings.TrimSpace(yb.String())
		logger.Debug("result execution", "yaml", yaml)
		if yaml == "" || yaml == "---" {
			continue
		}

		if i != 0 {
			sb.WriteString("\n---\n")
		}
		sb.WriteString(yaml)

		logger.Debug("end of execution template", "name", templateName)
	}

	return sb.String(), nil
}

const ServiceKey digo.ObjectKey = "k8sYamlService"

func init() {
	digo.ProvideService(ServiceKey, func(ctx *digo.Container) (any, error) {
		storageClient, err := digo.Get[*storage.Client](ctx, services.ServiceKeyStorageClient)
		if err != nil {
			return nil, err
		}

		switch ctx.Env {
		case digo.EnvProd:
			storageUrl := ctx.Config.Stoa.URL + "/storage/v1"
			return NewK8sYamlService(
				ctx.Env,
				storageClient,
				storageUrl,
			)
		case digo.EnvTest:
			return NewK8sYamlService(
				ctx.Env,
				storageClient,
				"http://apidepot-test.local.shaple.io/storage/v1",
			)
		default:
			return nil, errors.New("unknown env")
		}
	})
}
