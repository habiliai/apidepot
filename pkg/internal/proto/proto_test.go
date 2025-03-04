package proto_test

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/habiliai/apidepot/pkg/internal/instance"
	instancetest "github.com/habiliai/apidepot/pkg/internal/instance/test"
	"github.com/habiliai/apidepot/pkg/internal/project"
	projecttest "github.com/habiliai/apidepot/pkg/internal/project/test"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	prototest "github.com/habiliai/apidepot/pkg/internal/proto/test"
	"github.com/habiliai/apidepot/pkg/internal/services"
	servicestest "github.com/habiliai/apidepot/pkg/internal/services/test"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	stacktest "github.com/habiliai/apidepot/pkg/internal/stack/test"
	"github.com/habiliai/apidepot/pkg/internal/user"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	vapitest "github.com/habiliai/apidepot/pkg/internal/vapi/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/supabase-community/gotrue-go"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
	"net"
	"testing"
	"time"
)

type ProtoTestSuite struct {
	suite.Suite

	db         *gorm.DB
	grpcServer *grpc.Server

	vapis     *vapitest.ServiceMock
	stacks    *stacktest.ServiceMock
	projects  *projecttest.ServiceMock
	instances *instancetest.ServiceMock
	users     *usertest.Service
	git       *servicestest.MockGitService

	gotrueClient gotrue.Client

	eg errgroup.Group

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *ProtoTestSuite) Context() context.Context {
	return s.ctx
}

func (s *ProtoTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.Require().NotNil(s.ctx)

	container := digo.NewContainer(
		s.Context(),
		digo.EnvTest,
		nil,
	)
	s.vapis = &vapitest.ServiceMock{}
	digo.Set(container, vapi.ServiceKey, s.vapis)

	s.stacks = &stacktest.ServiceMock{}
	digo.Set(container, stack.ServiceKey, s.stacks)

	s.projects = &projecttest.ServiceMock{}
	digo.Set(container, project.ServiceKey, s.projects)

	s.instances = &instancetest.ServiceMock{}
	digo.Set(container, instance.ServiceKey, s.instances)

	s.users = usertest.NewService()
	digo.Set(container, user.ServiceKey, s.users)

	s.git = servicestest.NewTestGitService()
	digo.Set(container, services.ServiceKeyGitService, s.git)

	s.gotrueClient = digo.MustGet[gotrue.Client](container, services.ServiceKeyGoTrueClient)

	var err error
	s.db, err = digo.Get[*gorm.DB](container, services.ServiceKeyDB)
	s.Require().NoError(err)

	s.grpcServer, err = digo.Get[*grpc.Server](container, proto.ServiceKeyGrpcServer)
	s.Require().NoError(err)

	s.eg.Go(func() error {
		lc := net.ListenConfig{}
		listener, err := lc.Listen(s.ctx, "tcp", "0.0.0.0:15543")
		if err != nil {
			return err
		}
		return s.grpcServer.Serve(listener)
	})

	{
		ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
		defer cancel()
		prototest.WaitForServing(s.T(), ctx, "127.0.0.1:15543")
	}
}

func (s *ProtoTestSuite) TearDownTest() {
	defer s.cancel()

	s.grpcServer.Stop()
	s.Require().NoError(s.eg.Wait())
	s.resetGotrue()
}

func (s *ProtoTestSuite) resetGotrue() {
	listUsers, err := s.gotrueClient.AdminListUsers()
	s.Require().NoError(err)
	for _, user := range listUsers.Users {
		s.Require().NoError(s.gotrueClient.AdminDeleteUser(gotruetypes.AdminDeleteUserRequest{
			UserID: user.ID,
		}))
	}
}

func (s *ProtoTestSuite) newClient() (proto.ApiDepotClient, func() error) {
	conn, err := grpc.NewClient("127.0.0.1:15543", grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	return proto.NewApiDepotClient(conn), func() error {
		return errors.WithStack(conn.Close())
	}
}

func TestProto(t *testing.T) {
	suite.Run(t, new(ProtoTestSuite))
}
