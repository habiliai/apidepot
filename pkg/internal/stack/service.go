package stack

import (
	"context"
	pkgconfig "github.com/habiliai/apidepot/pkg/config"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/habiliai/apidepot/pkg/internal/storage"
	"github.com/habiliai/apidepot/pkg/internal/user"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/pkg/errors"
)

var logger = tclog.GetLogger()

type Service interface {
	GetStack(context.Context, uint) (*domain.Stack, error)
	GetStacks(
		ctx context.Context,
		projectId uint,
		name *string,
		cursor uint,
		limit int,
	) ([]domain.Stack, error)
	CreateStack(
		ctx context.Context,
		input CreateStackInput,
	) (*domain.Stack, error)
	DeleteStack(context.Context, uint) error
	PatchStack(
		ctx context.Context,
		id uint,
		input PatchStackInput,
	) error
	MigrateDatabase(context.Context, uint, MigrateDatabaseInput) error

	EnableOrUpdateAuth(ctx context.Context, stackId uint, input EnableOrUpdateAuthInput, isCreate bool) error
	DisableAuth(ctx context.Context, stackId uint) error

	EnableOrUpdateStorage(ctx context.Context, stackId uint, input EnableOrUpdateStorageInput, isCreate bool) error
	DisableStorage(ctx context.Context, stackId uint) error

	EnableOrUpdatePostgrest(ctx context.Context, stackId uint, input EnableOrUpdatePostgrestInput, isCreate bool) error
	DisablePostgrest(ctx context.Context, stackId uint) error

	EnableVapi(
		ctx context.Context,
		stackId uint,
		input EnableVapiInput,
	) (*domain.StackVapi, error)
	UpdateVapi(
		ctx context.Context,
		stackId uint,
		vapidId uint,
		input UpdateVapiInput,
	) (*domain.StackVapi, error)
	DisableVapi(
		ctx context.Context,
		stackId uint,
		vapiId uint,
	) error
	GetStorageUsage(ctx context.Context, stackId uint) (int64, error)
	GetMyTotalStorageUsage(
		ctx context.Context,
	) (int64, error)
	SetVapiEnv(
		ctx context.Context,
		stackId uint,
		env map[string]string,
	) error
	UnsetVapiEnv(ctx context.Context, stackId uint, names []string) error
	EnableCustomVapi(
		ctx context.Context,
		stackId uint,
		input CreateCustomVapiInput,
	) (*domain.CustomVapi, error)
	UpdateCustomVapi(
		ctx context.Context,
		stackId uint,
		name string,
		input UpdateCustomVapiInput,
	) error
	DisableCustomVapi(ctx context.Context, stackId uint, name string) error
	GetCustomVapiByName(
		ctx context.Context,
		stackId uint,
		name string,
	) (*domain.CustomVapi, error)
	GetCustomVapis(
		ctx context.Context,
		stackId uint,
	) ([]domain.CustomVapi, error)
}

type service struct {
	stackConfig   pkgconfig.StackConfig
	bucketService services.BucketService
	dbConfig      pkgconfig.DBConfig
	runtimeSchema *services.RuntimeSchema
	vapis         vapi.Service
	users         user.Service
	storageClient *storage.Client
	git           services.GitService
}

func NewService(
	bucketService services.BucketService,
	runtimeSchema *services.RuntimeSchema,
	stackConfig pkgconfig.StackConfig,
	dbConfig pkgconfig.DBConfig,
	vapiService vapi.Service,
	userService user.Service,
	storageClient *storage.Client,
	git services.GitService,
) Service {
	return &service{
		stackConfig:   stackConfig,
		bucketService: bucketService,
		runtimeSchema: runtimeSchema,
		dbConfig:      dbConfig,
		vapis:         vapiService,
		users:         userService,
		storageClient: storageClient,
		git:           git,
	}
}

const ServiceKey digo.ObjectKey = "stackService"

func init() {
	digo.ProvideService(ServiceKey, func(ctx *digo.Container) (any, error) {
		bs, err := digo.Get[services.BucketService](ctx, services.ServiceKeyBucketService)
		if err != nil {
			return nil, err
		}

		rs, err := digo.Get[*services.RuntimeSchema](ctx, services.ServiceKeyRuntimeSchema)
		if err != nil {
			return nil, err
		}

		vapis, err := digo.Get[vapi.Service](ctx, vapi.ServiceKey)
		if err != nil {
			return nil, err
		}

		users, err := digo.Get[user.Service](ctx, user.ServiceKey)
		if err != nil {
			return nil, err
		}

		storageClient, err := digo.Get[*storage.Client](ctx, services.ServiceKeyStorageClient)
		if err != nil {
			return nil, err
		}

		gitClient, err := digo.Get[services.GitService](ctx, services.ServiceKeyGitService)
		if err != nil {
			return nil, err
		}

		switch ctx.Env {
		case digo.EnvProd:
			return NewService(
				bs,
				rs,
				ctx.Config.Stack,
				ctx.Config.DB,
				vapis,
				users,
				storageClient,
				gitClient,
			), nil
		case digo.EnvTest:
			{
				regionalConf := pkgconfig.RegionalStackConfig{
					Domain: "local.shaple.io",
					Scheme: "http",
				}
				regionalDbConf := pkgconfig.RegionalDBConfig{
					Host:     "localhost",
					User:     "postgres",
					Password: "postgres",
					Name:     "test",
					Port:     6543,
				}
				return NewService(
					bs,
					rs,
					pkgconfig.StackConfig{
						ForceDelete: true,
						Seoul:       regionalConf,
						Singapore:   regionalConf,
					},
					pkgconfig.DBConfig{
						Seoul:           regionalDbConf,
						Singapore:       regionalDbConf,
						AutoMigration:   true,
						MaxIdleConns:    10,
						MaxOpenConns:    10,
						ConnMaxLifetime: "1h",
					},
					vapis,
					users,
					storageClient,
					gitClient,
				), nil
			}
		default:
			return nil, errors.New("unknown env")
		}
	})
}
