package instance

import (
	"context"
	"github.com/habiliai/apidepot/pkg/config"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/k8s"
	"github.com/habiliai/apidepot/pkg/internal/k8syaml"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/pkg/errors"
)

var (
	logger = tclog.GetLogger()
)

const (
	ServiceKey = "instanceService"
)

type (
	Service interface {
		CreateInstance(
			ctx context.Context,
			input CreateInstanceInput,
		) (*domain.Instance, error)
		DeployStack(
			ctx context.Context,
			instanceId uint,
			input DeployStackInput,
		) error
		LaunchInstance(
			ctx context.Context,
			instanceId uint,
		) error
		StopInstance(
			ctx context.Context,
			instanceId uint,
		) error
		DeleteInstance(
			ctx context.Context,
			instanceId uint,
			force bool,
		) error
		RestartInstance(
			ctx context.Context,
			instanceId uint,
		) error
		EditInstance(
			ctx context.Context,
			instanceId uint,
			input EditInstanceInput,
		) error
		GetInstance(
			ctx context.Context,
			instanceId uint,
		) (*domain.Instance, error)
		GetInstancesInStack(
			ctx context.Context,
			stackId uint,
		) ([]domain.Instance, error)
	}

	service struct {
		k8sClientPool  *k8s.ClientPool
		k8sYamlService *k8syaml.Service
		smtpConfig     config.SMTPConfig
		s3Config       config.S3Config
		vapis          vapi.Service
		stackConfig    config.StackConfig
		dbConfig       config.DBConfig
		users          user.Service
		stacks         stack.Service
	}
)

func NewService(
	k8sClientPool *k8s.ClientPool,
	k8sYamlService *k8syaml.Service,
	smtpConfig config.SMTPConfig,
	s3Config config.S3Config,
	vapiService vapi.Service,
	stackConfig config.StackConfig,
	dbConfig config.DBConfig,
	userService user.Service,
	stackService stack.Service,
) Service {
	return &service{
		k8sClientPool:  k8sClientPool,
		k8sYamlService: k8sYamlService,
		smtpConfig:     smtpConfig,
		s3Config:       s3Config,
		vapis:          vapiService,
		stackConfig:    stackConfig,
		dbConfig:       dbConfig,
		users:          userService,
		stacks:         stackService,
	}
}

func init() {
	digo.ProvideService(ServiceKey, func(ctx *digo.Container) (any, error) {
		k8sClientPool, err := digo.Get[*k8s.ClientPool](ctx, k8s.ServiceKeyK8sClientPool)
		if err != nil {
			return nil, err
		}

		k8sYamlService, err := digo.Get[*k8syaml.Service](ctx, k8syaml.ServiceKey)
		if err != nil {
			return nil, err
		}

		vapiService, err := digo.Get[vapi.Service](ctx, vapi.ServiceKey)
		if err != nil {
			return nil, err
		}

		users, err := digo.Get[user.Service](ctx, user.ServiceKey)
		if err != nil {
			return nil, err
		}

		stacks, err := digo.Get[stack.Service](ctx, stack.ServiceKey)
		if err != nil {
			return nil, err
		}

		switch ctx.Env {
		case digo.EnvTest:
			regionalStackConfig := config.RegionalStackConfig{
				Scheme: "http",
				Domain: "local.shaple.io",
			}
			regionalDbConfig := config.RegionalDBConfig{
				Host:     "localhost",
				User:     "postgres",
				Password: "postgres",
				Name:     "test",
				Port:     6543,
			}
			return NewService(
				k8sClientPool,
				k8sYamlService,
				config.SMTPConfig{
					Host:       "smtp.gmail.com",
					Port:       587,
					Username:   "apidepot",
					Password:   "apidepot",
					AdminEmail: "noreply@habili.ai",
				},
				config.S3Config{
					AccessKey: "minioadmin",
					SecretKey: "minioadmin",
					Seoul: config.RegionalS3Config{
						Endpoint: "http://minio.local.shaple.io",
					},
					Singapore: config.RegionalS3Config{
						Endpoint: "http://minio.local.shaple.io",
					},
				},
				vapiService,
				config.StackConfig{
					Seoul:           regionalStackConfig,
					Singapore:       regionalStackConfig,
					ForceDelete:     true,
					SkipHealthCheck: false,
				},
				config.DBConfig{
					Seoul:           regionalDbConfig,
					Singapore:       regionalDbConfig,
					AutoMigration:   true,
					MaxIdleConns:    10,
					MaxOpenConns:    10,
					ConnMaxLifetime: "1h",
				},
				users,
				stacks,
			), nil
		case digo.EnvProd:
			return NewService(
				k8sClientPool,
				k8sYamlService,
				ctx.Config.SMTP,
				ctx.Config.S3,
				vapiService,
				ctx.Config.Stack,
				ctx.Config.DB,
				users,
				stacks,
			), nil
		}

		return nil, errors.New("unknown environment")
	})
}
