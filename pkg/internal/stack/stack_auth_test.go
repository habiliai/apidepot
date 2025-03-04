package stack_test

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/mokiat/gog"
	"time"
)

func (s *StackServiceTestSuite) TestGivenAuthDisabledStackWhenDisableAuthShouldBeError() {
	// when
	err := s.stackService.DisableAuth(s.Context, s.stack.ID)

	// then
	s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
}

func (s *StackServiceTestSuite) TestGivenAuthDisabledStackWhenEnableAuthThenAuthShouldBeEnabled() {
	// when
	s.installAuth(s.Context, s.stack)
	defer s.uninstallAuth(s.Context, s.stack)
}

func (s *StackServiceTestSuite) TestGivenAuthEnabledStackWhenEnableAuthShouldBeError() {
	// given
	s.installAuth(s.Context, s.stack)
	defer s.uninstallAuth(s.Context, s.stack)

	// when
	err := s.stackService.EnableOrUpdateAuth(
		s.Context,
		s.stack.ID,
		stack.EnableOrUpdateAuthInput{},
		true,
	)

	// then
	s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
}

func (s *StackServiceTestSuite) TestGivenNotEnabledAuthWhenPatchAuthShouldBeError() {
	ctx := s.Context

	// when
	err := s.stackService.EnableOrUpdateAuth(
		ctx,
		s.stack.ID,
		stack.EnableOrUpdateAuthInput{},
		false,
	)

	// then
	s.ErrorAs(err, &tclerrors.ErrPreconditionRequired)
}

func (s *StackServiceTestSuite) installAuth(ctx context.Context, st *domain.Stack) {
	require := s.Require()

	validUntilStr := "2024-01-23T23:00:00+09:00"
	_, err := time.Parse(time.RFC3339, validUntilStr)
	require.NoError(err, "failed to parse time. format: %s", time.Now().Format(time.RFC3339))

	s.Require().NoError(s.stackService.EnableOrUpdateAuth(ctx, st.ID, stack.EnableOrUpdateAuthInput{
		AuthInput: stack.AuthInput{
			SenderName:   gog.PtrOf("Shaple"),
			AdminEmail:   gog.PtrOf("shaple@shaple.io"),
			PhoneEnabled: gog.PtrOf(true),
			EmailEnabled: gog.PtrOf(true),

			Exp: gog.PtrOf("1h"),

			MailerAutoConfirm:          gog.PtrOf(true),
			MailerInviteSubject:        gog.PtrOf("Shaple Invite"),
			MailerInviteTemplate:       gog.PtrOf("Hello, {{.Email}}! Link: {{.SiteURL}}"),
			MailerConfirmationSubject:  gog.PtrOf("Shaple Confirmation"),
			MailerConfirmationTemplate: gog.PtrOf("Hello, {{.Email}}! Please confirm your email."),
			MailerRecoverySubject:      gog.PtrOf("Shaple Recovery"),
			MailerRecoveryTemplate:     gog.PtrOf("Hello, {{.Email}}! Please recover your password."),
			MailerEmailChangeSubject:   gog.PtrOf("Shaple Email Change"),
			MailerEmailChangeTemplate:  gog.PtrOf("Hello, {{.Email}}! Please confirm your email change."),
			MailerMagicLinkSubject:     gog.PtrOf("Shaple Magic Link"),
			MailerMagicLinkTemplate:    gog.PtrOf("Hello, {{.Email}}! Please confirm your magic link."),

			SMSAutoConfirm:    gog.PtrOf(true),
			TestOTP:           gog.PtrOf("8201012341234:123456,8201034563456:123123"),
			TestOTPValidUntil: &validUntilStr,
		},
	}, true))

	{
		output, err := s.stackService.GetStack(ctx, st.ID)
		s.Require().NoError(err)

		s.Require().Equal(st.ID, output.ID)

		auth := output.Auth.Data()
		s.Require().Equal(auth.MailerInviteSubject, "Shaple Invite")
		s.Require().Equal(auth.MailerInviteTemplate, "Hello, {{.Email}}! Link: {{.SiteURL}}")
		s.Require().Equal(auth.MailerConfirmationSubject, "Shaple Confirmation")
		s.Require().Equal(auth.MailerConfirmationTemplate, "Hello, {{.Email}}! Please confirm your email.")
		s.Require().Equal(auth.MailerRecoverySubject, "Shaple Recovery")
		s.Require().Equal(auth.MailerRecoveryTemplate, "Hello, {{.Email}}! Please recover your password.")
		s.Require().Equal(auth.MailerEmailChangeSubject, "Shaple Email Change")
		s.Require().Equal(auth.MailerEmailChangeTemplate, "Hello, {{.Email}}! Please confirm your email change.")
		s.Require().Equal(auth.MailerMagicLinkSubject, "Shaple Magic Link")
		s.Require().Equal(auth.MailerMagicLinkTemplate, "Hello, {{.Email}}! Please confirm your magic link.")
	}
}

func (s *StackServiceTestSuite) uninstallAuth(ctx context.Context, stack *domain.Stack) {
	s.NoError(s.stackService.DisableAuth(ctx, stack.ID))
}
