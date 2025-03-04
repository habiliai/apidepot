package k8syaml

import (
	pkgconfig "github.com/habiliai/apidepot/pkg/config"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"time"
)

type (
	AuthYamlValues struct {
		RateLimitEmailSent    float64
		RateLimitSMSSent      float64
		RateLimitVerify       float64
		RateLimitTokenRefresh float64
		RateLimitSSO          float64

		SiteURL string
		JWT     struct {
			Secret string
			Exp    time.Duration
		}
		SMTP struct {
			Host       string
			Port       int
			SenderName string
			AdminEmail string
			Username   string
			Password   string
		}
		Mailer struct {
			AutoConfirm bool
			Subjects    struct {
				Confirmation string
				Recovery     string
				Invite       string
				EmailChange  string
				MagicLink    string
			}
			Templates struct {
				Recovery     string
				Invite       string
				EmailChange  string
				Confirmation string
				MagicLink    string
			}
		}
		SMS struct {
			AutoConfirm  bool
			MaxFrequency string
			OTPExp       time.Duration
			OTPLength    int
			Provider     string
			Twilio       struct {
				AccountSID        string
				AuthToken         string
				MessageServiceSID string
				ContentSID        string
			}
			TwilioVerify struct {
				AccountSID        string
				AuthToken         string
				MessageServiceSID string
			}
			MessageBird struct {
				AccessKey  string
				Originator string
			}
			Vonage struct {
				APIKey    string
				APISecret string
				From      string
			}
			TestOTP           string
			TestOTPValidUntil string
		}
		External struct {
			EmailEnabled   bool
			PhoneEnabled   bool
			IOSBundleID    string
			OAuthProviders []struct {
				Enabled        bool
				Name           string
				Secret         string
				ClientID       string
				SkipNonceCheck bool
			}
			RedirectURL string
		}
		MFA struct {
			Enabled                     bool
			ChallengeExpiryDuration     time.Duration
			RateLimitChallengeAndVerify float64
			MaxEnrolledFactors          float64
			MaxVerifiedFactors          int
		}
		Security struct {
			Captcha struct {
				Enabled  bool
				Secret   string
				Provider string
			}
			ManualLinkingEnabled bool
		}
		Webhook struct {
			URL     string
			Retries int
			Timeout time.Duration
			Events  []string
			Secret  string
		}
	}

	StorageYamlValues struct {
		S3 struct {
			Bucket    string
			AccessKey string
			SecretKey string
			Endpoint  string
		}
		TenantID string
	}

	PostgrestYamlValues struct {
		Schemas []string
	}

	Values struct {
		ShapleEnv string

		Project *domain.Project
		Stack   *domain.Stack
		DB      domain.DB
		Paths   struct {
			Auth           string
			Postgrest      string
			Storage        string
			PostgrestReady string
			PostgrestLive  string
			Vapi           string
			CustomVapi     string
		}
		Auth        *AuthYamlValues
		Storage     *StorageYamlValues
		Postgrest   *PostgrestYamlValues
		Vapis       []VapiYamlValues
		CustomVapis []CustomVapiYamlValues
	}
)

func (s *Service) NewValuesFromStack(stack *domain.Stack) Values {
	values := Values{
		ShapleEnv: string(s.shapleEnv),
		Project:   &stack.Project,
		Stack:     stack,
		DB:        stack.DB.Data(),
	}
	values.Paths.Auth = constants.PathAuth
	values.Paths.Storage = constants.PathStorage
	values.Paths.Postgrest = constants.PathPostgrest
	values.Paths.PostgrestLive = constants.PathPostgrestLive
	values.Paths.PostgrestReady = constants.PathPostgrestReady
	values.Paths.Vapi = constants.PathVapis

	return values
}

func (v Values) WithPostgrest() Values {
	postgrest := v.Stack.Postgrest.Data()
	var values PostgrestYamlValues
	values.Schemas = postgrest.Schemas

	v.Postgrest = &values
	return v
}

func (v Values) WithAuth(smtpConfig pkgconfig.SMTPConfig) Values {
	stack := v.Stack
	auth := stack.Auth.Data()

	var values AuthYamlValues
	values.SiteURL = stack.SiteURL
	values.JWT.Secret = auth.JWTSecret
	values.JWT.Exp = auth.JWTExp
	values.SMTP.Host = smtpConfig.Host
	values.SMTP.Port = smtpConfig.Port
	values.SMTP.Username = smtpConfig.Username
	values.SMTP.Password = smtpConfig.Password
	values.SMTP.SenderName = auth.SMTPSenderName
	values.SMTP.AdminEmail = smtpConfig.AdminEmail
	values.Mailer.AutoConfirm = auth.MailerAutoConfirm
	values.Mailer.Subjects.Invite = auth.MailerInviteSubject
	values.Mailer.Subjects.Confirmation = auth.MailerConfirmationSubject
	values.Mailer.Subjects.Recovery = auth.MailerRecoverySubject
	values.Mailer.Subjects.EmailChange = auth.MailerEmailChangeSubject
	values.Mailer.Subjects.MagicLink = auth.MailerMagicLinkSubject
	values.Mailer.Templates.Invite = auth.MailerInviteTemplate
	values.Mailer.Templates.Confirmation = auth.MailerConfirmationTemplate
	values.Mailer.Templates.Recovery = auth.MailerRecoveryTemplate
	values.Mailer.Templates.EmailChange = auth.MailerEmailChangeTemplate
	values.Mailer.Templates.MagicLink = auth.MailerMagicLinkTemplate
	values.SMS.AutoConfirm = auth.SMSAutoConfirm
	values.SMS.OTPExp = auth.SMSOTPExp
	values.SMS.OTPLength = auth.SMSOTPLength
	values.SMS.Provider = auth.SMSProvider
	values.SMS.Twilio.AccountSID = auth.SMSTwilioAccountSID
	values.SMS.Twilio.AuthToken = auth.SMSTwilioAuthToken
	values.SMS.Twilio.MessageServiceSID = auth.SMSTwilioMessageServiceSID
	values.SMS.Twilio.ContentSID = auth.SMSTwilioContentSID
	values.SMS.TwilioVerify.AccountSID = auth.SMSTwilioVerifyAccountSID
	values.SMS.TwilioVerify.AuthToken = auth.SMSTwilioVerifyAuthToken
	values.SMS.TwilioVerify.MessageServiceSID = auth.SMSTwilioVerifyMessageServiceSID
	values.SMS.MessageBird.AccessKey = auth.SMSMessagebirdAccessKey
	values.SMS.MessageBird.Originator = auth.SMSMessagebirdOriginator
	values.SMS.Vonage.APIKey = auth.SMSVonageAPIKey
	values.SMS.Vonage.APISecret = auth.SMSVonageAPISecret
	values.SMS.Vonage.From = auth.SMSVonageFrom
	values.SMS.TestOTP = auth.SMSTestOTP
	values.SMS.TestOTPValidUntil = auth.SMSTestOTPValidUntil
	values.SMS.MaxFrequency = auth.SMSMaxFrequency
	values.External.IOSBundleID = auth.ExternalIOSBundleID
	values.External.EmailEnabled = auth.ExternalEmailEnabled
	values.External.PhoneEnabled = auth.ExternalPhoneEnabled
	for _, oauthProvider := range auth.ExternalOAuthProviders {
		values.External.OAuthProviders = append(values.External.OAuthProviders, struct {
			Enabled        bool
			Name           string
			Secret         string
			ClientID       string
			SkipNonceCheck bool
		}{
			Enabled:        oauthProvider.Enabled,
			Name:           oauthProvider.Name,
			Secret:         oauthProvider.Secret,
			ClientID:       oauthProvider.ClientID,
			SkipNonceCheck: oauthProvider.SkipNonceCheck,
		})
	}
	values.External.RedirectURL = auth.ExternalRedirectURL
	values.MFA.Enabled = auth.MFAEnabled
	values.MFA.ChallengeExpiryDuration = auth.MFAChallengeExpiryDuration
	values.MFA.RateLimitChallengeAndVerify = auth.MFARateLimitChallengeAndVerify
	values.MFA.MaxEnrolledFactors = auth.MFAMaxEnrolledFactors
	values.MFA.MaxVerifiedFactors = auth.MFAMaxVerifiedFactors
	values.Security.Captcha.Enabled = auth.SecurityCaptchaEnabled
	values.Security.Captcha.Provider = auth.SecurityCaptchaProvider
	values.Security.Captcha.Secret = auth.SecurityCaptchaSecret
	values.Security.ManualLinkingEnabled = auth.SecurityManualLinkingEnabled
	values.RateLimitTokenRefresh = auth.RateLimitTokenRefresh
	values.RateLimitSSO = auth.RateLimitSSO
	values.RateLimitVerify = auth.RateLimitVerify
	values.RateLimitSMSSent = auth.RateLimitSMSSent
	values.RateLimitEmailSent = auth.RateLimitEmailSent

	v.Auth = &values

	return v
}

func (v Values) WithStorage(s3Config pkgconfig.S3Config, region tcltypes.InstanceZone) Values {
	stack := v.Stack
	storage := stack.Storage.Data()

	regionalConfig := s3Config.GetRegionalConfig(region)

	var values StorageYamlValues
	values.S3.Bucket = storage.S3Bucket
	values.S3.AccessKey = s3Config.AccessKey
	values.S3.SecretKey = s3Config.SecretKey
	values.S3.Endpoint = regionalConfig.Endpoint
	values.TenantID = storage.TenantID
	v.Storage = &values

	return v
}

func (v Values) WithVapis(values []VapiYamlValues) Values {
	v.Vapis = values

	return v
}

func (v Values) WithCustomVapis(values []CustomVapiYamlValues) Values {
	v.CustomVapis = values

	return v
}
