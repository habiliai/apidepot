package stack

import (
	"context"
	"github.com/Masterminds/goutils"
	"github.com/golang-jwt/jwt/v5"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type AuthInput struct {
	// SMTP settings
	SenderName *string `json:"sender_name"`
	AdminEmail *string `json:"admin_email"`

	// External Auth settings
	EmailEnabled   *bool                              `json:"email_enabled"`
	PhoneEnabled   *bool                              `json:"phone_enabled"`
	IOSBundleID    *string                            `json:"ios_bundle_id"`
	RedirectURL    *string                            `json:"redirect_url"`
	OAuthProviders []domain.AuthExternalOAuthProvider `json:"oauth_providers"`

	// JWT settings
	Exp *string `json:"exp"`

	// Mailer settings
	MailerAutoConfirm          *bool   `json:"mailer_auto_confirm"`
	MailerConfirmationSubject  *string `json:"mailer_confirmation_subject"`
	MailerRecoverySubject      *string `json:"mailer_recovery_subject"`
	MailerInviteSubject        *string `json:"mailer_invite_subject"`
	MailerEmailChangeSubject   *string `json:"mailer_email_change_subject"`
	MailerMagicLinkSubject     *string `json:"mailer_magic_link_subject"`
	MailerRecoveryTemplate     *string `json:"mailer_recovery_template"`
	MailerInviteTemplate       *string `json:"mailer_invite_template"`
	MailerEmailChangeTemplate  *string `json:"mailer_email_change_template"`
	MailerConfirmationTemplate *string `json:"mailer_confirmation_template"`
	MailerMagicLinkTemplate    *string `json:"mailer_magic_link_template"`

	// SMS settings
	SMSAutoConfirm                *bool   `json:"sms_auto_confirm"`
	SMSMaxFrequency               *string `json:"sms_max_frequency"`
	OTPExp                        *string `json:"otp_exp"`
	OTPLength                     *int32  `json:"otp_length"`
	SMSProvider                   *string `json:"sms_provider"`
	TwilioAccountSID              *string `json:"twilio_account_sid"`
	TwilioAuthToken               *string `json:"twilio_auth_token"`
	TwilioMessageServiceSID       *string `json:"twilio_message_service_sid"`
	TwilioContentSID              *string `json:"twilio_content_sid"`
	TwilioVerifyAccountSID        *string `json:"twilio_verify_account_sid"`
	TwilioVerifyAuthToken         *string `json:"twilio_verify_auth_token"`
	TwilioVerifyMessageServiceSID *string `json:"twilio_verify_message_service_sid"`
	MessagebirdAccessKey          *string `json:"messagebird_access_key"`
	MessagebirdOrginator          *string `json:"messagebird_orginator"`
	VonageAPIKey                  *string `json:"vonage_api_key"`
	VonageAPISecret               *string `json:"vonage_api_secret"`
	VonageFrom                    *string `json:"vonage_from"`
	TestOTP                       *string `json:"test_otp"`
	TestOTPValidUntil             *string `json:"test_otp_valid_until"`

	// MFA settings
	MFAEnabled                  *bool    `json:"mfa_enabled"`
	ChallengeExpiryDuration     *string  `json:"challenge_expiry_duration"`
	RateLimitChallengeAndVerify *float64 `json:"rate_limit_challenge_and_verify"`
	MaxEnrolledFactors          *float64 `json:"max_enrolled_factors"`
	MaxVerifiedFactors          *int32   `json:"max_verified_factors"`

	// Captcha settings
	CaptchaEnabled  *bool   `json:"captcha_enabled"`
	CaptchaSecret   *string `json:"captcha_secret"`
	CaptchaProvider *string `json:"captcha_provider"`

	// Rate limits
	RateLimitEmailSent    *float64 `json:"rate_limit_email_sent"`
	RateLimitSMSSent      *float64 `json:"rate_limit_sms_sent"`
	RateLimitVerify       *float64 `json:"rate_limit_verify"`
	RateLimitTokenRefresh *float64 `json:"rate_limit_token_refresh"`
	RateLimitSSO          *float64 `json:"rate_limit_sso"`

	// Security settings
	SecurityManualLinkingEnabled *bool `json:"security_manual_linking_enabled"`
}

type EnableOrUpdateAuthInput struct {
	AuthInput
}

func (input EnableOrUpdateAuthInput) mergeToDomainAuth(auth *domain.Auth) error {
	var err error

	if auth.JWTSecret == "" {
		auth.JWTSecret, err = goutils.CryptoRandomAlphaNumeric(32)
		if err != nil {
			return errors.Wrapf(err, "failed to generate random hash")
		}
	}

	if input.SenderName != nil {
		auth.SMTPSenderName = *input.SenderName
	} else {
		auth.SMTPSenderName = "Admin"
	}

	if input.EmailEnabled != nil {
		auth.ExternalEmailEnabled = *input.EmailEnabled
	}
	if input.PhoneEnabled != nil {
		auth.ExternalPhoneEnabled = *input.PhoneEnabled
	}
	if input.IOSBundleID != nil {
		auth.ExternalIOSBundleID = *input.IOSBundleID
	}
	if len(input.OAuthProviders) > 0 {
		auth.ExternalOAuthProviders = input.OAuthProviders
	}
	if input.RedirectURL != nil {
		auth.ExternalRedirectURL = *input.RedirectURL
	}

	if input.MFAEnabled != nil {
		auth.MFAEnabled = *input.MFAEnabled
	}
	if input.ChallengeExpiryDuration != nil {
		ced := *input.ChallengeExpiryDuration
		if ced == "" {
			ced = "0s"
		}
		if auth.MFAChallengeExpiryDuration, err = time.ParseDuration(ced); err != nil {
			return errors.Wrapf(err, "failed to parse duration")
		}
	}
	if input.RateLimitChallengeAndVerify != nil {
		auth.MFARateLimitChallengeAndVerify = *input.RateLimitChallengeAndVerify
	}
	if input.MaxEnrolledFactors != nil {
		auth.MFAMaxEnrolledFactors = *input.MaxEnrolledFactors
	}
	if input.MaxVerifiedFactors != nil {
		auth.MFAMaxVerifiedFactors = int(*input.MaxVerifiedFactors)
	}

	if input.CaptchaEnabled != nil {
		auth.SecurityCaptchaEnabled = *input.CaptchaEnabled
	}
	if input.CaptchaSecret != nil {
		auth.SecurityCaptchaSecret = *input.CaptchaSecret
	}
	if input.CaptchaProvider != nil {
		auth.SecurityCaptchaProvider = *input.CaptchaProvider
	}

	if input.Exp != nil {
		exp := *input.Exp
		if exp == "" {
			exp = "0s"
		}
		if auth.JWTExp, err = time.ParseDuration(exp); err != nil {
			return errors.Wrapf(err, "failed to parse duration")
		}
	}

	if input.MailerAutoConfirm != nil {
		auth.MailerAutoConfirm = *input.MailerAutoConfirm
	}
	if input.MailerConfirmationSubject != nil {
		auth.MailerConfirmationSubject = *input.MailerConfirmationSubject
	}
	if input.MailerRecoverySubject != nil {
		auth.MailerRecoverySubject = *input.MailerRecoverySubject
	}
	if input.MailerInviteSubject != nil {
		auth.MailerInviteSubject = *input.MailerInviteSubject
	}
	if input.MailerEmailChangeSubject != nil {
		auth.MailerEmailChangeSubject = *input.MailerEmailChangeSubject
	}
	if input.MailerMagicLinkSubject != nil {
		auth.MailerMagicLinkSubject = *input.MailerMagicLinkSubject
	}
	if input.MailerRecoveryTemplate != nil {
		auth.MailerRecoveryTemplate = *input.MailerRecoveryTemplate
	}
	if input.MailerInviteTemplate != nil {
		auth.MailerInviteTemplate = *input.MailerInviteTemplate
	}
	if input.MailerEmailChangeTemplate != nil {
		auth.MailerEmailChangeTemplate = *input.MailerEmailChangeTemplate
	}
	if input.MailerConfirmationTemplate != nil {
		auth.MailerConfirmationTemplate = *input.MailerConfirmationTemplate
	}
	if input.MailerMagicLinkTemplate != nil {
		auth.MailerMagicLinkTemplate = *input.MailerMagicLinkTemplate
	}

	if input.SMSAutoConfirm != nil {
		auth.SMSAutoConfirm = *input.SMSAutoConfirm
	}

	if input.SMSMaxFrequency != nil {
		auth.SMSMaxFrequency = *input.SMSMaxFrequency
	} else {
		auth.SMSMaxFrequency = "60s"
	}
	if input.OTPExp != nil {
		otpExp := *input.OTPExp
		if otpExp == "" {
			otpExp = "0s"
		}
		if auth.SMSOTPExp, err = time.ParseDuration(otpExp); err != nil {
			return errors.Wrapf(err, "failed to parse duration")
		}
	}
	if input.OTPLength != nil {
		auth.SMSOTPLength = int(*input.OTPLength)
	}
	if input.SMSProvider != nil {
		auth.SMSProvider = *input.SMSProvider
	}
	if input.TwilioAccountSID != nil {
		auth.SMSTwilioAccountSID = *input.TwilioAccountSID
	}
	if input.TwilioAuthToken != nil {
		auth.SMSTwilioAuthToken = *input.TwilioAuthToken
	}
	if input.TwilioMessageServiceSID != nil {
		auth.SMSTwilioMessageServiceSID = *input.TwilioMessageServiceSID
	}
	if input.TwilioContentSID != nil {
		auth.SMSTwilioContentSID = *input.TwilioContentSID
	}
	if input.TwilioVerifyAccountSID != nil {
		auth.SMSTwilioVerifyAccountSID = *input.TwilioVerifyAccountSID
	}
	if input.TwilioVerifyAuthToken != nil {
		auth.SMSTwilioVerifyAuthToken = *input.TwilioVerifyAuthToken
	}
	if input.TwilioVerifyMessageServiceSID != nil {
		auth.SMSTwilioVerifyMessageServiceSID = *input.TwilioVerifyMessageServiceSID
	}
	if input.MessagebirdAccessKey != nil {
		auth.SMSMessagebirdAccessKey = *input.MessagebirdAccessKey
	}
	if input.MessagebirdOrginator != nil {
		auth.SMSMessagebirdOriginator = *input.MessagebirdOrginator
	}
	if input.VonageAPIKey != nil {
		auth.SMSVonageAPIKey = *input.VonageAPIKey
	}
	if input.VonageAPISecret != nil {
		auth.SMSVonageAPISecret = *input.VonageAPISecret
	}
	if input.VonageFrom != nil {
		auth.SMSVonageFrom = *input.VonageFrom
	}
	if input.TestOTP != nil {
		auth.SMSTestOTP = *input.TestOTP
	}
	if input.TestOTPValidUntil != nil {
		auth.SMSTestOTPValidUntil = *input.TestOTPValidUntil
	}

	if input.RateLimitEmailSent != nil {
		auth.RateLimitEmailSent = *input.RateLimitEmailSent
	}
	if input.RateLimitSMSSent != nil {
		auth.RateLimitSMSSent = *input.RateLimitSMSSent
	}
	if input.RateLimitVerify != nil {
		auth.RateLimitVerify = *input.RateLimitVerify
	}
	if input.RateLimitTokenRefresh != nil {
		auth.RateLimitTokenRefresh = *input.RateLimitTokenRefresh
	}
	if input.RateLimitSSO != nil {
		auth.RateLimitSSO = *input.RateLimitSSO
	}
	if input.SecurityManualLinkingEnabled != nil {
		auth.SecurityManualLinkingEnabled = *input.SecurityManualLinkingEnabled
	}

	return nil
}

func (ss *service) EnableOrUpdateAuth(ctx context.Context, stackID uint, input EnableOrUpdateAuthInput, isCreate bool) error {
	tx := helpers.GetTx(ctx)
	stack, err := ss.GetStack(ctx, stackID)
	if err != nil {
		return err
	}

	if err := ss.hasPermission(ctx, stack.ProjectID); err != nil {
		return err
	}

	if stack.AuthEnabled && isCreate {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "auth is already created")
	} else if !stack.AuthEnabled && !isCreate {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "auth is not created")
	}

	stack.AuthEnabled = true

	auth := stack.Auth.Data()
	if err := input.mergeToDomainAuth(&auth); err != nil {
		return err
	}

	if err := auth.Validate(); err != nil {
		return err
	}

	stack.Auth = datatypes.NewJSONType(auth)
	stack.AdminApiKey, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "service_role",
	}).SignedString([]byte(stack.Auth.Data().JWTSecret))
	if err != nil {
		return errors.Wrapf(err, "failed to sign jwt")
	}
	stack.AnonApiKey, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "anon",
	}).SignedString([]byte(stack.Auth.Data().JWTSecret))
	if err != nil {
		return errors.Wrapf(err, "failed to sign jwt")
	}

	if err := tx.Transaction(func(tx *gorm.DB) (err error) {
		if err = stack.Save(tx.Omit(clause.Associations)); err != nil {
			return
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("auth is enabling for stack")

	return nil
}

func (ss *service) DisableAuth(ctx context.Context, stackID uint) error {
	tx := helpers.GetTx(ctx)
	stack, err := ss.GetStack(ctx, stackID)
	if err != nil {
		return err
	}

	if user, err := ss.users.GetUser(ctx); err != nil {
		return err
	} else if stack.Project.OwnerID != user.ID && !user.IsSuperuser() {
		return errors.Wrapf(tclerrors.ErrForbidden, "you are not allowed to access this stack")
	}

	if !stack.AuthEnabled {
		return errors.Wrapf(tclerrors.ErrPreconditionRequired, "auth is not enabled")
	}

	stack.AuthEnabled = false
	return tx.Transaction(func(tx *gorm.DB) error {
		if err := stack.Save(tx.Omit(clause.Associations)); err != nil {
			return err
		}

		return nil
	})
}
