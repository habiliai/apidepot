package stack_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/jackc/pgx/v5"
)

func (s *StackServiceTestSuite) TestMigrateVapiDatabaseGivenDuplicated() {
	ctx, cancel := context.WithCancel(s.Context)
	defer cancel()

	// Given already exists migration same version
	{
		conn, err := pgx.Connect(
			ctx,
			fmt.Sprintf(
				"postgres://postgres:postgres@localhost:6543/%s?search_path=stack&sslmode=disable",
				s.stack.DB.Data().Name,
			),
		)
		s.Require().NoError(err)
		defer conn.Close(ctx)
		s.Require().NoError(pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "INSERT INTO stack.vapi_schema_migrations(version, vapi_package_id) VALUES ('2024-03-26 13:52:00', 1)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "INSERT INTO stack.vapi_schema_migrations(version, vapi_package_id) VALUES ('2024-03-26 13:52:00', 2)"); err != nil {
				return err
			}

			return nil
		}))
	}

	// when enable vapis to migrate db same version
	s.enableVapis(ctx, s.stack.ID)
	defer s.disableVapis(ctx, s.stack.ID)

	conn, err := pgx.Connect(
		ctx,
		fmt.Sprintf(
			"postgres://postgres:postgres@localhost:6543/%s?search_path=helloworld&sslmode=disable",
			s.stack.DB.Data().Name,
		),
	)
	s.Require().NoError(err)
	defer conn.Close(ctx)
	s.Require().NoError(conn.Ping(ctx))

	_, err = conn.Exec(
		ctx,
		"INSERT INTO profiles(name) VALUES('test');",
	)

	// Then not ok
	s.Require().Error(err)
}

func (s *StackServiceTestSuite) enableVapis(
	ctx context.Context,
	stackId uint,
) {
	s.T().Logf("-- enable vapi")

	st, err := domain.FindStackByID(s.db, stackId)
	s.Require().NoError(err)

	s.T().Logf("stack: %s", st.String())

	_, err = s.vapis.Register(
		ctx,
		"habiliai/vapi-helloworld-sns",
		"main",
		"sns",
		"test",
		[]string{},
		uuid.NewString(),
		"https://habili.ai",
	)
	s.Require().NoError(err)

	_, err = s.vapis.Register(
		ctx,
		"habiliai/vapi-helloworld",
		"main",
		"helloworld",
		"test",
		[]string{},
		uuid.NewString(),
		"https://habili.ai",
	)
	s.Require().NoError(err)

	vapiRelease, err := domain.FindVapiReleaseByPackageNameAndVersion(s.db, "helloworld", "0.1.0")
	s.Require().NoError(err)

	_, err = s.stackService.EnableVapi(
		ctx,
		st.ID,
		stack.EnableVapiInput{
			VapiID: vapiRelease.ID,
		},
	)
	s.Require().NoError(err)

	s.T().Logf("wait for available all vapis")
}

func (s *StackServiceTestSuite) disableVapis(
	ctx context.Context,
	stackId uint,
) {
	s.T().Log("-- disable vapi")

	st, err := domain.FindStackByID(s.db, stackId)
	s.Require().NoError(err)

	for _, stackVapi := range st.Vapis {
		s.NoError(s.stackService.DisableVapi(
			ctx,
			stackId,
			stackVapi.VapiID,
		))
	}
}
