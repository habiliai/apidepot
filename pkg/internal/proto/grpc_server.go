package proto

import (
	"context"
	"fmt"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"strings"
)

const (
	ServiceKeyGrpcServer = "proto.grpcServer"
)

func handleErrorToGrpcStatus(err error) error {
	if err == nil {
		return nil
	}

	logger.Error(fmt.Sprintf("grpc route error: %+v", err))
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, tclerrors.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	} else if errors.Is(err, tclerrors.ErrBadRequest) {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if errors.Is(err, tclerrors.ErrForbidden) {
		return status.Error(codes.PermissionDenied, err.Error())
	} else if errors.Is(err, tclerrors.ErrUnauthorized) {
		return status.Error(codes.Unauthenticated, err.Error())
	} else if errors.Is(err, gorm.ErrDuplicatedKey) {
		return status.Error(codes.AlreadyExists, err.Error())
	} else if errors.Is(err, tclerrors.ErrPreconditionFailed) {
		return status.Error(codes.FailedPrecondition, err.Error())
	} else if errors.Is(err, tclerrors.ErrPreconditionRequired) {
		return status.Error(codes.FailedPrecondition, err.Error())
	} else {
		return status.Error(codes.Internal, err.Error())
	}
}

func createGrpcServer(
	db *gorm.DB,
	server ApiDepotServer,
) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (resp interface{}, rErr error) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}

			if values, ok := md["x-device-id"]; ok && len(values) > 0 {
				deviceId := strings.TrimSpace(values[0])
				ctx = helpers.WithDeviceId(ctx, deviceId)
				logger.Debug("metadata", "x-device-id", deviceId)
			}

			if values, ok := md["authorization"]; ok && len(values) > 0 {
				token, ok := strings.CutPrefix(values[0], "Bearer")
				if !ok {
					return nil, errors.Wrapf(tclerrors.ErrUnauthorized, "Invalid authorization header")
				}
				token = strings.TrimSpace(token)
				ctx = helpers.WithAuthToken(ctx, token)
				logger.Debug("metadata", "token", token)
			}

			if values, ok := md["x-github-token"]; ok && len(values) > 0 {
				token := strings.TrimSpace(values[0])
				ctx = helpers.WithGithubToken(ctx, token)
				logger.Debug("metadata", "x-github-token", token)
			}

			rErr = db.WithContext(ctx).Transaction(func(tx *gorm.DB) (err error) {
				ctx = helpers.WithTx(ctx, tx)

				logger.Info("call", "method", info.FullMethod)
				resp, err = handler(ctx, req)
				return
			})

			rErr = handleErrorToGrpcStatus(rErr)

			return
		}),
	)
	RegisterApiDepotServer(grpcServer, server)
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	return grpcServer, nil
}

func init() {
	digo.ProvideService(ServiceKeyGrpcServer, func(ctx *digo.Container) (interface{}, error) {
		db, err := digo.Get[*gorm.DB](ctx, services.ServiceKeyDB)
		if err != nil {
			return nil, err
		}

		apiDepotServer, err := digo.Get[ApiDepotServer](ctx, ServiceKey)
		if err != nil {
			return nil, err
		}

		return createGrpcServer(db, apiDepotServer)
	})
}
