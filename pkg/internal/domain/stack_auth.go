package domain

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type AuthExternalOAuthProvider struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Name           string `json:"name,omitempty"`
	Secret         string `json:"secret,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	SkipNonceCheck bool   `json:"skip_nonce_check,omitempty"`
}

type Auth struct {
	JWTSecret                        string                      `json:"jwt_secret,omitempty"`
	JWTExp                           time.Duration               `json:"jwt_exp,omitempty"`
	SMTPSenderName                   string                      `json:"smtp_sender_name,omitempty"`
	MailerAutoConfirm                bool                        `json:"mailer_auto_confirm,omitempty"`
	MailerConfirmationSubject        string                      `json:"mailer_confirmation_subject,omitempty"`
	MailerRecoverySubject            string                      `json:"mailer_recovery_subject,omitempty"`
	MailerInviteSubject              string                      `json:"mailer_invite_subject,omitempty"`
	MailerEmailChangeSubject         string                      `json:"mailer_email_change_subject,omitempty"`
	MailerMagicLinkSubject           string                      `json:"mailer_magic_link_subject,omitempty"`
	MailerRecoveryTemplate           string                      `json:"mailer_recovery_template,omitempty"`
	MailerConfirmationTemplate       string                      `json:"mailer_confirmation_template,omitempty"`
	MailerInviteTemplate             string                      `json:"mailer_invite_template,omitempty"`
	MailerEmailChangeTemplate        string                      `json:"mailer_email_change_template,omitempty"`
	MailerMagicLinkTemplate          string                      `json:"mailer_magic_link_template,omitempty"`
	SMSAutoConfirm                   bool                        `json:"sms_auto_confirm,omitempty"`
	SMSMaxFrequency                  string                      `json:"sms_max_frequency,omitempty"`
	SMSOTPExp                        time.Duration               `json:"sms_otp_exp,omitempty"`
	SMSOTPLength                     int                         `json:"sms_otp_length,omitempty"`
	SMSProvider                      string                      `json:"sms_provider,omitempty"`
	SMSTwilioAccountSID              string                      `json:"sms_twilio_account_sid,omitempty"`
	SMSTwilioAuthToken               string                      `json:"sms_twilio_auth_token,omitempty"`
	SMSTwilioMessageServiceSID       string                      `json:"sms_twilio_message_service_sid,omitempty"`
	SMSTwilioContentSID              string                      `json:"sms_twilio_content_sid,omitempty"`
	SMSTwilioVerifyAccountSID        string                      `json:"sms_twilio_verify_account_sid,omitempty"`
	SMSTwilioVerifyAuthToken         string                      `json:"sms_twilio_verify_auth_token,omitempty"`
	SMSTwilioVerifyMessageServiceSID string                      `json:"sms_twilio_verify_message_service_sid,omitempty"`
	SMSMessagebirdAccessKey          string                      `json:"sms_messagebird_access_key,omitempty"`
	SMSMessagebirdOriginator         string                      `json:"sms_messagebird_originator,omitempty"`
	SMSVonageAPIKey                  string                      `json:"sms_vonage_api_key,omitempty"`
	SMSVonageAPISecret               string                      `json:"sms_vonage_api_secret,omitempty"`
	SMSVonageFrom                    string                      `json:"sms_vonage_from,omitempty"`
	SMSTestOTP                       string                      `json:"sms_test_otp,omitempty"`
	SMSTestOTPValidUntil             string                      `json:"sms_test_otp_valid_until,omitempty"`
	ExternalEmailEnabled             bool                        `json:"external_email_enabled,omitempty"`
	ExternalPhoneEnabled             bool                        `json:"external_phone_enabled,omitempty"`
	ExternalIOSBundleID              string                      `json:"external_ios_bundle_id,omitempty"`
	ExternalOAuthProviders           []AuthExternalOAuthProvider `json:"external_oauth_providers,omitempty"`
	MFAEnabled                       bool                        `json:"mfa_enabled,omitempty"`
	MFAChallengeExpiryDuration       time.Duration               `json:"mfa_challenge_expiry_duration,omitempty"`
	MFARateLimitChallengeAndVerify   float64                     `json:"mfa_rate_limit_challenge_and_verify,omitempty"`
	MFAMaxEnrolledFactors            float64                     `json:"mfa_enrolled_factors,omitempty"`
	MFAMaxVerifiedFactors            int                         `json:"mfa_verified_factors,omitempty"`
	SecurityCaptchaEnabled           bool                        `json:"security_captcha_enabled,omitempty"`
	SecurityCaptchaSecret            string                      `json:"security_captcha_secret,omitempty"`
	SecurityCaptchaProvider          string                      `json:"security_captcha_provider,omitempty"`
	SecurityManualLinkingEnabled     bool                        `json:"security_manual_linking_enabled,omitempty"`
	ExternalRedirectURL              string                      `json:"external_redirect_url,omitempty"`
	RateLimitEmailSent               float64                     `json:"rate_limit_email_sent,omitempty"`
	RateLimitSMSSent                 float64                     `json:"rate_limit_sms_sent,omitempty"`
	RateLimitVerify                  float64                     `json:"rate_limit_verify,omitempty"`
	RateLimitTokenRefresh            float64                     `json:"rate_limit_token_refresh,omitempty"`
	RateLimitSSO                     float64                     `json:"rate_limit_sso,omitempty"`
}

func (a Auth) Validate() error {
	switch strings.ToLower(a.SMSProvider) {
	case "":
		break
	case "twillio":
		if a.SMSTwilioAccountSID == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "twillio account sid is required")
		}
		if a.SMSTwilioAuthToken == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "twillio auth token is required")
		}
		if a.SMSTwilioMessageServiceSID == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "twillio message service sid is required")
		}
		if a.SMSTwilioContentSID == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "twilio content sid is required")
		}
	case "messagebird":
		if a.SMSMessagebirdAccessKey == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "messagebird access key is required")
		}
		if a.SMSMessagebirdOriginator == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "messagebird originator is required")
		}
	case "vonage":
		if a.SMSVonageAPIKey == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "vonage api key is required")
		}

		if a.SMSVonageAPISecret == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "vonage api secret is required")
		}

		if a.SMSVonageFrom == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "vonage from is required")
		}
	case "twilio_verify":
		if a.SMSTwilioVerifyAccountSID == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "twilio verify account sid is required")
		}
		if a.SMSTwilioVerifyAuthToken == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "twilio verify auth token is required")
		}
		if a.SMSTwilioVerifyMessageServiceSID == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "twilio verify message service sid is required")
		}
	default:
		return errors.Wrapf(tclerrors.ErrValidation, "sms provider=%s is not unsupported", a.SMSProvider)
	}

	for _, oauthProvider := range a.ExternalOAuthProviders {
		if oauthProvider.Name == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "oauth provider name is required")
		}
		if oauthProvider.Secret == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "%s oauth provider secret is required", oauthProvider.Name)
		}
		if oauthProvider.ClientID == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "%s oauth provider client id is required", oauthProvider.Name)
		}
	}

	if a.SecurityCaptchaEnabled {
		if a.SecurityCaptchaProvider == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "captcha provider is required")
		}
		if a.SecurityCaptchaSecret == "" {
			return errors.Wrapf(tclerrors.ErrValidation, "captcha secret is required")
		}
	}

	return nil
}
