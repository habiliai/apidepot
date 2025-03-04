package proto

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/mokiat/gog"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

func newProjectPbFromDb(project *domain.Project) *Project {
	return &Project{
		Id:          int32(project.ID),
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   tspb.New(project.CreatedAt),
		UpdatedAt:   tspb.New(project.UpdatedAt),
		Stacks: gog.Map(project.Stacks, func(s domain.Stack) *Stack {
			return newStackPbFromDb(s)
		}),
	}
}

func newStackPbFromDb(stack domain.Stack) *Stack {
	result := &Stack{
		Id:               int32(stack.ID),
		Name:             stack.Name,
		Description:      stack.Description,
		CreatedAt:        tspb.New(stack.CreatedAt),
		UpdatedAt:        tspb.New(stack.UpdatedAt),
		PostgrestEnabled: stack.PostgrestEnabled,
		AuthEnabled:      stack.AuthEnabled,
		StorageEnabled:   stack.StorageEnabled,
		Vapis: gog.Map(stack.Vapis, func(v domain.StackVapi) *StackVapi {
			return newStackVapiPbFromDb(&v)
		}),
		Auth: newStackAuthPbFromDb(stack.Auth.Data()),
		Storage: &StackStorage{
			TenantId: stack.Storage.Data().TenantID,
			S3Bucket: stack.Storage.Data().S3Bucket,
		},
		Postgrest: &StackPostgrest{
			Schemas: stack.Postgrest.Data().Schemas,
		},
		ProjectId:   int32(stack.ProjectID),
		GitRepo:     stack.GitRepo,
		GitBranch:   stack.GitBranch,
		Domain:      stack.Domain,
		Scheme:      stack.Scheme,
		SiteUrl:     stack.SiteURL,
		AdminApiKey: stack.AdminApiKey,
		AnonApiKey:  stack.AnonApiKey,
		Db: &StackDB{
			Name:     stack.DB.Data().Name,
			Username: stack.DB.Data().Username,
			Password: stack.DB.Data().Password,
		},
		LogoImageUrl:  stack.LogoImageUrl,
		CustomVapis:   gog.Map(stack.CustomVapis, newCustomVapiPbFromDb),
		DefaultRegion: getInstanceZonePbFromDb(stack.DefaultRegion),
	}

	if stack.TelegramMiniappPromotion != nil {
		result.TelegramMiniappPromotion = newTelegramMiniappPromotionPbFromDb(*stack.TelegramMiniappPromotion, false)
	}

	if stack.ServiceTemplateID != nil {
		result.ServiceTemplateId = gog.PtrOf(int32(*stack.ServiceTemplateID))
	}

	return result
}

func newCustomVapiPbFromDb(v domain.CustomVapi) *CustomVapi {
	return &CustomVapi{
		StackId:     int32(v.StackID),
		Name:        v.Name,
		TarFilePath: v.TarFilePath,
	}
}

func newStackAuthPbFromDb(auth domain.Auth) *StackAuth {
	return &StackAuth{
		JwtSecret:                        auth.JWTSecret,
		JwtExp:                           int64(auth.JWTExp),
		SmtpSenderName:                   auth.SMTPSenderName,
		MailerAutoConfirm:                auth.MailerAutoConfirm,
		MailerConfirmationSubject:        auth.MailerConfirmationSubject,
		MailerRecoverySubject:            auth.MailerRecoverySubject,
		MailerInviteSubject:              auth.MailerInviteSubject,
		MailerEmailChangeSubject:         auth.MailerEmailChangeSubject,
		MailerMagicLinkSubject:           auth.MailerMagicLinkSubject,
		MailerRecoveryTemplate:           auth.MailerRecoveryTemplate,
		MailerConfirmationTemplate:       auth.MailerConfirmationTemplate,
		MailerInviteTemplate:             auth.MailerInviteTemplate,
		MailerEmailChangeTemplate:        auth.MailerEmailChangeTemplate,
		MailerMagicLinkTemplate:          auth.MailerMagicLinkTemplate,
		SmsAutoConfirm:                   auth.SMSAutoConfirm,
		SmsOtpExp:                        int64(auth.SMSOTPExp),
		SmsOtpLength:                     int32(auth.SMSOTPLength),
		SmsProvider:                      auth.SMSProvider,
		SmsTwilioAccountSid:              auth.SMSTwilioAccountSID,
		SmsTwilioAuthToken:               auth.SMSTwilioAuthToken,
		SmsTwilioMessageServiceSid:       auth.SMSTwilioMessageServiceSID,
		SmsTwilioContentSid:              auth.SMSTwilioContentSID,
		SmsTwilioVerifyAccountSid:        auth.SMSTwilioVerifyAccountSID,
		SmsTwilioVerifyAuthToken:         auth.SMSTwilioVerifyAuthToken,
		SmsTwilioVerifyMessageServiceSid: auth.SMSTwilioVerifyMessageServiceSID,
		SmsMessagebirdAccessKey:          auth.SMSMessagebirdAccessKey,
		SmsMessagebirdOriginator:         auth.SMSMessagebirdOriginator,
		SmsVonageApiKey:                  auth.SMSVonageAPIKey,
		SmsVonageApiSecret:               auth.SMSVonageAPISecret,
		SmsVonageFrom:                    auth.SMSVonageFrom,
		SmsTestOtp:                       auth.SMSTestOTP,
		SmsTestOtpValidUntil:             auth.SMSTestOTPValidUntil,
		ExternalEmailEnabled:             auth.ExternalEmailEnabled,
		ExternalPhoneEnabled:             auth.ExternalPhoneEnabled,
		ExternalIosBundleId:              auth.ExternalIOSBundleID,
		ExternalOauthProviders: gog.Map(auth.ExternalOAuthProviders, func(p domain.AuthExternalOAuthProvider) *StackAuthExternalOAuthProvider {
			return &StackAuthExternalOAuthProvider{
				Enabled:        p.Enabled,
				Name:           p.Name,
				Secret:         p.Secret,
				ClientId:       p.ClientID,
				SkipNonceCheck: p.SkipNonceCheck,
			}
		}),
		MfaEnabled:                     auth.MFAEnabled,
		MfaChallengeExpiryDuration:     int64(auth.MFAChallengeExpiryDuration),
		MfaRateLimitChallengeAndVerify: float32(auth.MFARateLimitChallengeAndVerify),
		MfaMaxEnrolledFactors:          float32(auth.MFAMaxEnrolledFactors),
		MfaMaxVerifiedFactors:          int32(auth.MFAMaxVerifiedFactors),
		SecurityCaptchaEnabled:         auth.SecurityCaptchaEnabled,
		SecurityCaptchaSecret:          auth.SecurityCaptchaSecret,
		SecurityCaptchaProvider:        auth.SecurityCaptchaProvider,
		ExternalRedirectUrl:            auth.ExternalRedirectURL,
		RateLimitEmailSent:             float32(auth.RateLimitEmailSent),
		RateLimitSmsSent:               float32(auth.RateLimitSMSSent),
		RateLimitVerify:                float32(auth.RateLimitVerify),
		RateLimitTokenRefresh:          float32(auth.RateLimitTokenRefresh),
		RateLimitSso:                   float32(auth.RateLimitSSO),
	}
}

func newStackVapiPbFromDb(vapi *domain.StackVapi) *StackVapi {
	return &StackVapi{
		StackId: int32(vapi.StackID),
		VapiId:  int32(vapi.VapiID),
		Vapi:    newVapiReleasePbFromDb(&vapi.Vapi),
	}
}

func newVapiReleasePbFromDb(vapi *domain.VapiRelease) *VapiRelease {
	return &VapiRelease{
		Id:          int32(vapi.ID),
		CreatedAt:   tspb.New(vapi.CreatedAt),
		UpdatedAt:   tspb.New(vapi.UpdatedAt),
		Version:     vapi.Version,
		Deprecated:  vapi.Deprecated,
		Suspended:   vapi.Suspended,
		Published:   vapi.Published,
		TarFilePath: vapi.TarFilePath,
		GitHash:     vapi.GitHash,
		PackageId:   int32(vapi.PackageID),
	}
}

func newVapiPackagePbFromDb(v *domain.VapiPackage) *VapiPackage {
	return &VapiPackage{
		Id:          int32(v.ID),
		Name:        v.Name,
		CreatedAt:   tspb.New(v.CreatedAt),
		UpdatedAt:   tspb.New(v.UpdatedAt),
		OwnerId:     int32(v.OwnerId),
		GitRepo:     v.GitRepo,
		GitBranch:   v.GitBranch,
		Releases:    nil,
		Description: v.Description,
		Domains:     v.Domains,
	}
}

func newOrganizationPbFromDb(org domain.Organization) *Organization {
	return &Organization{
		Id:        int32(org.ID),
		Name:      org.Name,
		CreatedAt: tspb.New(org.CreatedAt),
		UpdatedAt: tspb.New(org.UpdatedAt),
	}
}

func newUserPbFromDb(user domain.User) *User {
	return &User{
		Id:     int32(user.ID),
		AuthId: user.AuthUserId,
		Profile: &UserProfile{
			Name:           user.Name,
			Description:    user.Description,
			GithubEmail:    user.GithubEmail,
			GithubUsername: user.GithubUsername,
			MediumLink:     user.MediumLink,
			AvatarUrl:      user.AvatarUrl,
		},
		GithubInstallationId: user.GithubInstallationId,
		GithubAccessToken:    user.GithubAccessToken,
	}
}

func newInstancePbFromDb(instance *domain.Instance) *Instance {
	return &Instance{
		Id:          int32(instance.ID),
		Name:        instance.Name,
		StackId:     int32(instance.StackID),
		NumReplicas: int32(instance.NumReplicas),
		MaxReplicas: int32(instance.MaxReplicas),
		State:       getInstanceStatePbFromDb(instance.State),
		CreatedAt:   tspb.New(instance.CreatedAt),
		UpdatedAt:   tspb.New(instance.UpdatedAt),
		Zone:        getInstanceZonePbFromDb(instance.Zone),
	}
}

func getInstanceStatePbFromDb(state domain.InstanceState) Instance_InstanceState {
	return Instance_InstanceState(state)
}

func newServiceTemplatePbFromDb(st domain.ServiceTemplate) *ServiceTemplate {
	return &ServiceTemplate{
		Id:              int32(st.ID),
		Name:            st.Name,
		ConcentImageUrl: st.ConceptImageUrl,
		Detail:          st.Detail,
		GitRepo:         st.GitRepo,
		GitHash:         st.GitHash,
		Description:     st.Description,
		PrimaryImageUrl: st.PrimaryImageUrl,
	}
}

func (z Instance_InstanceZone) ToDomain() tcltypes.InstanceZone {
	switch z {
	case Instance_InstanceZoneOciApSeoul:
		return tcltypes.InstanceZoneOciApSeoul
	case Instance_InstanceZoneDefault:
		return tcltypes.InstanceZoneDefault
	case Instance_InstanceZoneOciSingapore:
		return tcltypes.InstanceZoneOciSingapore
	default:
		return ""
	}
}

func getInstanceZonePbFromDb(z tcltypes.InstanceZone) Instance_InstanceZone {
	switch z {
	case tcltypes.InstanceZoneOciApSeoul:
		return Instance_InstanceZoneOciApSeoul
	case tcltypes.InstanceZoneOciSingapore:
		return Instance_InstanceZoneOciSingapore
	default:
		return Instance_InstanceZoneNone
	}
}

func newTelegramMiniappPromotionPbFromDb(promo domain.TelegramMiniappPromotion, includeStack bool) *TelegramMiniappPromotion {
	logger.Debug("fields", "num views", promo.GetNumUniqueViews())
	result := &TelegramMiniappPromotion{
		LinkUrl:                promo.Link,
		AppTitle:               promo.AppTitle,
		AppIconImageUrl:        promo.AppIconImageUrl,
		AppDescription:         promo.AppDescription,
		AppScreenshotImageUrls: promo.AppScreenshotImageUrls,
		AppBannerImageUrl:      promo.AppBannerImageUrl,
		NumUniqueViews:         int32(promo.GetNumUniqueViews()),
		Public:                 promo.Public,
	}

	if promo.Stack.ID == promo.StackID && includeStack {
		logger.Debug("fields", "stack.serviceTemplateId", promo.Stack.ServiceTemplateID)
		result.Stack = newStackPbFromDb(promo.Stack)
	}

	return result
}
