package stack_test

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/mokiat/gog"
	"time"
)

func (s *StackServiceTestSuite) enableStorage(ctx context.Context, st *domain.Stack) {
	s.T().Log("-- enable storage")
	s.Require().NoError(s.stackService.EnableOrUpdateStorage(ctx, st.ID, stack.EnableOrUpdateStorageInput{
		StorageInput: stack.StorageInput{
			TenantID: gog.PtrOf("root"),
		},
	}, true))
	time.Sleep(250 * time.Millisecond)
}

func (s *StackServiceTestSuite) disableStorage(ctx context.Context, stack *domain.Stack) {
	s.T().Log("-- disable storage")
	s.NoError(s.stackService.DisableStorage(ctx, stack.ID))
}

func (s *StackServiceTestSuite) TestEnableStorage1() {
	ctx := s.Context

	s.Run("given not enabled auth, when try to enable storage, then should return error", func() {
		// when
		err := s.stackService.EnableOrUpdateStorage(ctx, s.stack.ID, stack.EnableOrUpdateStorageInput{
			StorageInput: stack.StorageInput{
				TenantID: gog.PtrOf("root"),
			},
		}, true)

		// then
		s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
	})

}

func (s *StackServiceTestSuite) TestEnableStorage2() {
	ctx := s.Context

	s.Run("given enabled auth and storage, when try to enable storage, then should return error", func() {
		// given
		s.installAuth(ctx, s.stack)
		defer s.uninstallAuth(ctx, s.stack)

		s.enableStorage(ctx, s.stack)
		defer s.disableStorage(ctx, s.stack)

		// when
		err := s.stackService.EnableOrUpdateStorage(ctx, s.stack.ID, stack.EnableOrUpdateStorageInput{
			StorageInput: stack.StorageInput{
				TenantID: gog.PtrOf("root"),
			},
		}, true)

		// then
		s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
	})
}

func (s *StackServiceTestSuite) TestEnableStorage3() {
	ctx := s.Context

	s.Run("given not enabled storage, when try to patch storage, then should return error", func() {
		// given
		s.installAuth(ctx, s.stack)
		defer s.uninstallAuth(ctx, s.stack)

		// when
		err := s.stackService.EnableOrUpdateStorage(ctx, s.stack.ID, stack.EnableOrUpdateStorageInput{
			StorageInput: stack.StorageInput{
				TenantID: gog.PtrOf("root"),
			},
		}, false)

		// then
		s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
	})
}

func (s *StackServiceTestSuite) TestDisableStorage1() {
	ctx := s.Context
	s.Run("given enabled auth and storage, when try to disable storage, then should be ok", func() {
		s.installAuth(ctx, s.stack)
		defer s.uninstallAuth(ctx, s.stack)

		s.enableStorage(ctx, s.stack)

		// when
		err := s.stackService.DisableStorage(ctx, s.stack.ID)

		// then
		s.NoError(err)
	})
}

func (s *StackServiceTestSuite) TestDisableStorage2() {

	s.Run("given no condition, when try to disable postgrest, then should return error", func() {
		// when
		err := s.stackService.DisableStorage(s.Context, s.stack.ID)

		// then
		s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
	})
}
