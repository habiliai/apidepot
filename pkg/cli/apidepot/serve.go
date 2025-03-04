package apidepot

import (
	"fmt"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func (c *Cli) newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve APIDepot",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("start server 'APIDepot'")
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
			defer cancel()

			cfg, err := c.getServerConfig(cmd.Flags())
			if err != nil {
				return err
			}

			container := digo.NewContainer(cmd.Context(), digo.EnvProd, cfg)
			grpcServer, err := digo.Get[*grpc.Server](container, proto.ServiceKeyGrpcServer)
			if err != nil {
				return err
			}
			defer grpcServer.GracefulStop()
			go func() {
				<-ctx.Done()
				grpcServer.GracefulStop()
			}()

			eg := errgroup.Group{}
			eg.Go(func() error {
				address := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
				listener, err := new(net.ListenConfig).Listen(ctx, "tcp", address)
				if err != nil {
					return errors.WithStack(err)
				}
				defer listener.Close()

				logger.Info("serving grpc", "address", address)
				return errors.WithStack(grpcServer.Serve(listener))
			})

			grpcWebServer := grpcweb.WrapServer(grpcServer, grpcweb.WithOriginFunc(func(origin string) bool {
				return true
			}))

			httpServer := http.Server{Handler: grpcWebServer}
			defer httpServer.Close()
			go func() {
				<-ctx.Done()
				httpServer.Close()
			}()

			eg.Go(func() error {
				address := fmt.Sprintf("%s:%d", cfg.Address, cfg.WebPort)
				listener, err := new(net.ListenConfig).Listen(ctx, "tcp", address)
				if err != nil {
					return errors.WithStack(err)
				}
				defer listener.Close()

				logger.Info("serving grpc web", "address", address)

				if err := httpServer.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
					return errors.WithStack(err)
				}
				return nil
			})

			return eg.Wait()
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	localAnonKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.GgzQrVUWAlI5UwMSCcjkOm7tDcjg8RmMBtOiSlOe9IM"
	localAdminKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0.eEgOdkCfkVwqEx0rsq1RT4LoSAX1cEZR3PzJ1Erm0oI"

	f := cmd.Flags()
	f.String("address", "", "Address to listen")
	f.Bool("includeDebug", false, "Include debug")
	f.Int("port", 8080, "Port to listen")
	f.Int("webPort", 8081, "Port to listen web server")
	f.String("k8s.kubeconfig", homeDir+"/.kube/config", "Path to kubeconfig file")
	f.String("k8s.seoul.context", "seoul", "Specify the kubeconfig context to use")
	f.String("k8s.singapore.context", "singapore", "Specify the kubeconfig context to use for Singapore")
	f.Float32("k8s.qps", 100, "Kubernetes QPS")
	f.Int("k8s.burst", 1000, "Kubernetes burst")
	f.String("db.seoul.host", "localhost", "Database host")
	f.Int("db.seoul.port", 6543, "Database port")
	f.String("db.seoul.user", "postgres", "Database user")
	f.String("db.seoul.password", "postgres", "Database password")
	f.String("db.seoul.name", "postgres", "Database name")
	f.String("db.singapore.host", "localhost", "Database host")
	f.Int("db.singapore.port", 6543, "Database port")
	f.String("db.singapore.user", "postgres", "Database user")
	f.String("db.singapore.password", "postgres", "Database password")
	f.String("db.singapore.name", "postgres", "Database name")
	f.String("db.pingTimeout", "5s", "Database ping timeout")
	f.Bool("db.autoMigration", true, "Auto migration")
	f.Int("db.maxIdleConns", 10, "Max idle connections")
	f.Int("db.maxOpenConns", 100, "Max open connections")
	f.String("db.connMaxLifetime", "1h", "Connection max lifetime")
	f.String("smtp.host", "email-smtp.us-east-1.amazonaws.com", "Host for smtp")
	f.Int("smtp.port", 587, "Port for smtp")
	f.String("smtp.username", "", "User for smtp")
	f.String("smtp.password", "", "Password for smtp")
	f.String("smtp.adminEmail", "noreply@habili.ai", "Admin email for smtp")
	f.String("stack.seoul.scheme", "http", "Scheme for stack in seoul")
	f.String("stack.seoul.domain", "local.shaple.io", "Domain for stack in seoul")
	f.String("stack.singapore.scheme", "http", "Scheme for stack in singapore")
	f.String("stack.singapore.domain", "local.shaple.io", "Domain for stack in singapore")
	f.Bool("stack.skipHealthCheck", false, "Skip health check for stack when checking service availability")
	f.String("stoa.url", "http://apidepot.local.shaple.io", "Stoacloud stack url")
	f.String("stoa.anonKey", localAnonKey, "Stoacloud stack anon key")
	f.String("stoa.adminKey", localAdminKey, "Stoacloud stack admin key")
	f.String("github.clientId", "Iv23liWx1gYoaKhVsahG", "Github client id")
	f.String("github.clientSecret", "", "Github client secret")
	f.String("github.appId", "1068010", "Github app id")
	f.String("github.appPrivateKey", "", "Github app private key (base64)")
	f.String("s3.accessKey", "minioadmin", "Access key for s3")
	f.String("s3.secretKey", "minioadmin", "Secret key for s3")
	f.String("s3.seoul.endpoint", "http://minio.local.shaple.io", "Regional endpoint for s3 in seoul")
	f.String("s3.singapore.endpoint", "http://minio.local.shaple.io", "Regional endpoint for s3 in singapore")

	return cmd
}
