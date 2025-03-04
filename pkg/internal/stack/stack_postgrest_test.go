package stack_test

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/stack"
)

func (s *StackServiceTestSuite) TestGivenNotEnabledAuthWhenTryToEnablePostgrestThenShouldReturnError() {
	// when
	err := s.stackService.EnableOrUpdatePostgrest(s, s.stack.ID, stack.EnableOrUpdatePostgrestInput{
		PostgrestInput: stack.PostgrestInput{
			Schemas: []string{"api", "public"},
		},
	}, true)
	defer s.stackService.DisablePostgrest(s, s.stack.ID)

	// then
	s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
}

func (s *StackServiceTestSuite) TestGivenEnabledPostgrestWhenTryToEnablePostgrestThenShouldReturnError() {
	// given
	s.installAuth(s, s.stack)
	defer s.uninstallAuth(s, s.stack)

	err := s.stackService.EnableOrUpdatePostgrest(s, s.stack.ID, stack.EnableOrUpdatePostgrestInput{
		PostgrestInput: stack.PostgrestInput{
			Schemas: []string{"api", "public"},
		},
	}, true)
	s.Require().NoError(err)
	defer s.stackService.DisablePostgrest(s, s.stack.ID)

	// when
	err = s.stackService.EnableOrUpdatePostgrest(s, s.stack.ID, stack.EnableOrUpdatePostgrestInput{
		PostgrestInput: stack.PostgrestInput{
			Schemas: []string{"api", "public"},
		},
	}, true)
	defer s.stackService.DisablePostgrest(s, s.stack.ID)

	// then
	s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
}

func (s *StackServiceTestSuite) TestGivenNotEnabledPostgrestWhenTryToUpdatePostgrestThenShouldReturnError() {
	// given
	ctx := s.Context
	s.installAuth(ctx, s.stack)
	defer s.uninstallAuth(ctx, s.stack)

	// when
	err := s.stackService.EnableOrUpdatePostgrest(ctx, s.stack.ID, stack.EnableOrUpdatePostgrestInput{
		PostgrestInput: stack.PostgrestInput{
			Schemas: []string{"api", "public"},
		},
	}, false)
	defer s.stackService.DisablePostgrest(ctx, s.stack.ID)

	// then
	s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
}

func (s *StackServiceTestSuite) TestDisablePostgrest1() {
	ctx := s.Context
	s.Run("given enabled postgrest, when try to disable postgrest, then should disable postgrest", func() {
		// given
		s.installAuth(ctx, s.stack)
		defer s.uninstallAuth(ctx, s.stack)

		s.Require().NoError(
			s.stackService.EnableOrUpdatePostgrest(ctx, s.stack.ID, stack.EnableOrUpdatePostgrestInput{
				PostgrestInput: stack.PostgrestInput{
					Schemas: []string{"api", "public"},
				},
			}, true),
		)

		// when
		err := s.stackService.DisablePostgrest(ctx, s.stack.ID)

		// then
		s.NoError(err)
	})
}

func (s *StackServiceTestSuite) TestDisablePostgrest2() {
	s.Run("when try to disable postgrest, then should return error", func() {
		// when
		err := s.stackService.DisablePostgrest(s.Context, s.stack.ID)

		// then
		s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
	})
}
