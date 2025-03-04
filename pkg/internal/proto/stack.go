package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	"github.com/mokiat/gog"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) UpdateStack(ctx context.Context, req *UpdateStackRequest) (*emptypb.Empty, error) {
	if err := s.stackService.PatchStack(ctx, uint(req.StackId), stack.PatchStackInput{
		SiteURL:      req.SiteUrl,
		Name:         req.Name,
		Description:  req.Description,
		LogoImageUrl: req.LogoImageUrl,
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetStacks(ctx context.Context, request *GetStacksRequest) (*GetStacksResponse, error) {
	stacks, err := s.stackService.GetStacks(
		ctx,
		uint(request.ProjectId),
		request.Name,
		uint(request.Cursor),
		int(request.Limit),
	)
	if err != nil {
		return nil, err
	}

	var resp GetStacksResponse
	for _, stack := range stacks {
		resp.Stacks = append(resp.Stacks, newStackPbFromDb(stack))
		resp.NextCursor = int32(stack.ID)
	}
	return &resp, nil
}

func (s *apiDepotServer) CreateStack(ctx context.Context, req *CreateStackRequest) (*Stack, error) {
	stack, err := s.stackService.CreateStack(ctx, stack.CreateStackInput{
		Name:          req.Name,
		ProjectID:     uint(req.ProjectId),
		SiteURL:       req.SiteUrl,
		Description:   req.Description,
		DefaultRegion: req.DefaultRegion.ToDomain(),
	})
	if err != nil {
		return nil, err
	}

	return newStackPbFromDb(*stack), nil
}

func (s *apiDepotServer) GetStackById(ctx context.Context, id *StackId) (*Stack, error) {
	stack, err := s.stackService.GetStack(ctx, uint(id.GetId()))
	if err != nil {
		return nil, err
	}

	return newStackPbFromDb(*stack), nil
}

func (s *apiDepotServer) DeleteStack(ctx context.Context, id *StackId) (*emptypb.Empty, error) {
	if err := s.stackService.DeleteStack(ctx, uint(id.GetId())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) InstallAuth(ctx context.Context, req *InstallAuthRequest) (*emptypb.Empty, error) {
	input := stack.AuthInput{
		SenderName:   req.SmtpSenderName,
		AdminEmail:   req.SmtpAdminEmail,
		EmailEnabled: req.ExternalEmailEnabled,
		PhoneEnabled: req.ExternalPhoneEnabled,
		IOSBundleID:  req.ExternalIosBundleId,
		RedirectURL:  req.ExternalRedirectUrl,
		OAuthProviders: gog.Map(req.ExternalOauthProviders, func(p *AuthExternalOAuthProvider) domain.AuthExternalOAuthProvider {
			var provider domain.AuthExternalOAuthProvider
			if p == nil {
				return provider
			}

			if p.Name != nil {
				provider.Name = *p.Name
			}
			if p.ClientId != nil {
				provider.ClientID = *p.ClientId
			}
			if p.Secret != nil {
				provider.Secret = *p.Secret
			}
			if p.Enabled != nil {
				provider.Enabled = *p.Enabled
			}
			if p.SkipNonceCheck != nil {
				provider.SkipNonceCheck = *p.SkipNonceCheck
			}

			return provider
		}),
		Exp:                           req.JwtExp,
		MailerAutoConfirm:             req.MailerAutoConfirm,
		MailerConfirmationSubject:     req.MailerConfirmationSubject,
		MailerRecoverySubject:         req.MailerRecoverySubject,
		MailerInviteSubject:           req.MailerInviteSubject,
		MailerEmailChangeSubject:      req.MailerEmailChangeSubject,
		MailerMagicLinkSubject:        req.MailerMagicLinkSubject,
		MailerRecoveryTemplate:        req.MailerRecoveryTemplate,
		MailerInviteTemplate:          req.MailerInviteTemplate,
		MailerEmailChangeTemplate:     req.MailerEmailChangeTemplate,
		MailerConfirmationTemplate:    req.MailerConfirmationTemplate,
		MailerMagicLinkTemplate:       req.MailerMagicLinkTemplate,
		SMSAutoConfirm:                req.SmsAutoConfirm,
		OTPExp:                        req.SmsOtpExp,
		OTPLength:                     req.SmsOtpLength,
		SMSProvider:                   req.SmsProvider,
		TwilioAccountSID:              req.SmsTwilioAccountSid,
		TwilioAuthToken:               req.SmsTwilioAuthToken,
		TwilioMessageServiceSID:       req.SmsTwilioMessageServiceSid,
		TwilioContentSID:              req.SmsTwilioContentSid,
		TwilioVerifyAccountSID:        req.SmsTwilioVerifyAccountSid,
		TwilioVerifyAuthToken:         req.SmsTwilioVerifyAuthToken,
		TwilioVerifyMessageServiceSID: req.SmsTwilioVerifyMessageServiceSid,
		MessagebirdAccessKey:          req.SmsMessagebirdAccessKey,
		MessagebirdOrginator:          req.SmsMessagebirdOrginator,
		VonageAPIKey:                  req.SmsVonageApiKey,
		VonageAPISecret:               req.SmsVonageApiSecret,
		VonageFrom:                    req.SmsVonageFrom,
		TestOTP:                       req.SmsTestOtp,
		TestOTPValidUntil:             req.SmsTestOtpValidUntil,
		MFAEnabled:                    req.MfaEnabled,
		ChallengeExpiryDuration:       req.MfaChallengeExpiryDuration,
		RateLimitChallengeAndVerify:   req.MfaRateLimitChallengeAndVerify,
		MaxEnrolledFactors:            req.MfaMaxEnrolledFactors,
		MaxVerifiedFactors:            req.MfaMaxVerifiedFactors,
		CaptchaEnabled:                req.CaptchaEnabled,
		CaptchaSecret:                 req.CaptchaSecret,
		CaptchaProvider:               req.CaptchaProvider,
		RateLimitEmailSent:            req.RateLimitEmailSent,
		RateLimitSMSSent:              req.RateLimitSmsSent,
		RateLimitVerify:               req.RateLimitVerify,
		RateLimitTokenRefresh:         req.RateLimitTokenRefresh,
		RateLimitSSO:                  req.RateLimitSso,
		SMSMaxFrequency:               req.SmsMaxFrequency,
		SecurityManualLinkingEnabled:  req.SecurityManualLinkingEnabled,
	}

	if err := s.stackService.EnableOrUpdateAuth(ctx, uint(req.Id), stack.EnableOrUpdateAuthInput{
		AuthInput: input,
	}, !req.IsUpdate); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) InstallPostgrest(ctx context.Context, req *InstallPostgrestRequest) (*emptypb.Empty, error) {
	input := stack.PostgrestInput{
		Schemas: req.Schemas,
	}

	if err := s.stackService.EnableOrUpdatePostgrest(ctx, uint(req.Id), stack.EnableOrUpdatePostgrestInput{
		PostgrestInput: input,
	}, !req.IsUpdate); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) InstallStorage(ctx context.Context, request *InstallStorageRequest) (*emptypb.Empty, error) {
	input := stack.StorageInput{
		TenantID: request.TenantId,
	}

	if err := s.stackService.EnableOrUpdateStorage(ctx, uint(request.Id), stack.EnableOrUpdateStorageInput{
		StorageInput: input,
	}, !request.IsUpdate); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) MigrateDatabase(ctx context.Context, request *MigrateDatabaseRequest) (*emptypb.Empty, error) {
	if err := s.stackService.MigrateDatabase(ctx, uint(request.StackId), stack.MigrateDatabaseInput{
		Migrations: gog.Map(request.Migrations, func(s *Migration) stack.Migration {
			return stack.Migration{
				Query:   s.Query,
				Version: s.Version.AsTime(),
			}
		}),
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetStackInstances(ctx context.Context, id *StackId) (*GetStackInstancesResponse, error) {
	instances, err := s.instanceService.GetInstancesInStack(ctx, uint(id.Id))
	if err != nil {
		return nil, err
	}

	return &GetStackInstancesResponse{
		Instances: gog.Map(instances, func(i domain.Instance) *Instance {
			return newInstancePbFromDb(&i)
		}),
	}, nil
}

func (s *apiDepotServer) UninstallAuth(ctx context.Context, id *StackId) (*emptypb.Empty, error) {
	if err := s.stackService.DisableAuth(ctx, uint(id.Id)); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) UninstallPostgrest(ctx context.Context, id *StackId) (*emptypb.Empty, error) {
	if err := s.stackService.DisablePostgrest(ctx, uint(id.Id)); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) UninstallStorage(ctx context.Context, id *StackId) (*emptypb.Empty, error) {
	if err := s.stackService.DisableStorage(ctx, uint(id.Id)); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetMyStorageUsage(ctx context.Context, req *emptypb.Empty) (*GetMyStorageUsageResponse, error) {
	numUsed, err := s.stackService.GetMyTotalStorageUsage(ctx)
	if err != nil {
		return nil, err
	}

	return &GetMyStorageUsageResponse{
		NumUsed: numUsed,
	}, nil
}
