package apidepotctl_test

import (
	"context"
	"fmt"
	"github.com/habiliai/apidepot/pkg/cli/apidepotctl"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	tclproto "github.com/habiliai/apidepot/pkg/internal/proto"
	prototest "github.com/habiliai/apidepot/pkg/internal/proto/test"
	"github.com/habiliai/apidepot/pkg/internal/services"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/supabase-community/gotrue-go"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

type ApiDepotCtlTestSuite struct {
	suite.Suite
	context.Context
	cancel context.CancelFunc

	cli         *apidepotctl.Cli
	cloudServer *prototest.ApiDepotServerMock
	gotrue      gotrue.Client

	grpcServer *grpc.Server
	grpcAddr   string

	session         gotruetypes.Session
	sessionEmail    string
	sessionPassword string

	eg errgroup.Group
}

func (s *ApiDepotCtlTestSuite) SetupTest() {
	ctx := context.TODO()
	s.cli = apidepotctl.NewCli("http://apidepot-test.local.shaple.io")

	container := digo.NewContainer(ctx, digo.EnvTest, nil)

	s.cloudServer = &prototest.ApiDepotServerMock{}
	s.cloudServer.Test(s.T())
	digo.Set(container, tclproto.ServiceKey, s.cloudServer)

	s.grpcServer = digo.MustGet[*grpc.Server](container, tclproto.ServiceKeyGrpcServer)
	port := 12390
	s.grpcAddr = fmt.Sprintf("127.0.0.1:%d", port)

	s.Context = ctx
	s.eg.Go(func() error {
		listener, err := new(net.ListenConfig).Listen(ctx, "tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return errors.WithStack(err)
		}
		defer listener.Close()

		s.T().Logf("serving grpc: %s", s.grpcAddr)
		return errors.WithStack(s.grpcServer.Serve(listener))
	})

	{
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		prototest.WaitForServing(s.T(), ctx, s.grpcAddr)
	}

	{
		s.sessionEmail = "test@test.com"
		s.sessionPassword = "qwer1234"

		s.gotrue = digo.MustGet[gotrue.Client](container, services.ServiceKeyGoTrueClient)
		signupResp, err := s.gotrue.Signup(gotruetypes.SignupRequest{
			Email:    s.sessionEmail,
			Password: s.sessionPassword,
		})
		s.Require().NoError(err)

		s.session = signupResp.Session
	}
}

func (s *ApiDepotCtlTestSuite) TearDownTest() {
	s.grpcServer.Stop()
	s.Require().NoError(s.eg.Wait())

	users, err := s.gotrue.AdminListUsers()
	s.Require().NoError(err)
	for _, user := range users.Users {
		s.Require().NoError(s.gotrue.AdminDeleteUser(gotruetypes.AdminDeleteUserRequest{UserID: user.ID}))
	}
}

func TestApiDepotCtl(t *testing.T) {
	suite.Run(t, new(ApiDepotCtlTestSuite))
}
