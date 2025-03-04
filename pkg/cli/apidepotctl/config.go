package apidepotctl

import (
	"strings"
	"time"
)

type (
	CliArgs struct {
		CliConfig
		ConfigFile string

		Login struct {
			Email    string
			Password string
		}
		Save bool

		StackVapi struct {
			Name    string
			Version string
		}
	}

	CliConfig struct {
		Server  string        `yaml:"server,omitempty"`
		ApiKey  string        `yaml:"apiKey,omitempty"`
		Timeout time.Duration `yaml:"timeout,omitempty"`

		Session struct {
			AppId     string `yaml:"appId,omitempty"`
			AppSecret string `yaml:"appSecret,omitempty"`
		} `yaml:"session,omitempty"`
		Stack struct {
			Name        string `yaml:"name,omitempty"`
			SiteUrl     string `yaml:"siteUrl,omitempty"`
			Description string `yaml:"description,omitempty"`

			Auth struct {
				SMTP struct {
					SenderName string `yaml:"senderName,omitempty"`
					AdminEmail string `yaml:"adminEmail,omitempty"`
				} `yaml:"smtp,omitempty"`
				External struct {
					EmailEnabled bool                `yaml:"emailEnabled,omitempty"`
					PhoneEnabled bool                `yaml:"phoneEnabled,omitempty"`
					ISOBundleID  string              `yaml:"isoBundleID,omitempty"`
					RedirectURL  string              `yaml:"redirectURL,omitempty"`
					Apple        OAuthProviderConfig `yaml:"apple,omitempty"`
					Azure        OAuthProviderConfig `yaml:"azure,omitempty"`
					Bitbucket    OAuthProviderConfig `yaml:"bitbucket,omitempty"`
					Discord      OAuthProviderConfig `yaml:"discord,omitempty"`
					Facebook     OAuthProviderConfig `yaml:"facebook,omitempty"`
					Figma        OAuthProviderConfig `yaml:"figma,omitempty"`
					Fly          OAuthProviderConfig `yaml:"fly,omitempty"`
					Github       OAuthProviderConfig `yaml:"github,omitempty"`
					Gitlab       OAuthProviderConfig `yaml:"gitlab,omitempty"`
					Google       OAuthProviderConfig `yaml:"google,omitempty"`
					Kakao        OAuthProviderConfig `yaml:"kakao,omitempty"`
					Notion       OAuthProviderConfig `yaml:"notion,omitempty"`
					Keycloak     OAuthProviderConfig `yaml:"keycloak,omitempty"`
					Linkedin     OAuthProviderConfig `yaml:"linkedin,omitempty"`
					LinkedinOIDC OAuthProviderConfig `yaml:"linkedinOIDC,omitempty"`
					Spotify      OAuthProviderConfig `yaml:"spotify,omitempty"`
					Slack        OAuthProviderConfig `yaml:"slack,omitempty"`
					Twitter      OAuthProviderConfig `yaml:"twitter,omitempty"`
					Twitch       OAuthProviderConfig `yaml:"twitch,omitempty"`
					WorkOS       OAuthProviderConfig `yaml:"workOS,omitempty"`
					Zoom         OAuthProviderConfig `yaml:"zoom,omitempty"`
				} `yaml:"external,omitempty"`
				JWT struct {
					EXP string `yaml:"exp,omitempty"`
				} `yaml:"jwt,omitempty"`
				Mailer struct {
					AutoConfirm bool `yaml:"autoConfirm,omitempty"`
					Subjects    struct {
						Confirmation string `yaml:"confirmation,omitempty"`
						Recovery     string `yaml:"recovery,omitempty"`
						Invite       string `yaml:"invite,omitempty"`
						EmailChange  string `yaml:"emailChange,omitempty"`
						MagicLink    string `yaml:"magicLink,omitempty"`
					} `yaml:"subjects,omitempty"`
					Templates struct {
						Recovery     string `yaml:"recovery,omitempty"`
						Invite       string `yaml:"invite,omitempty"`
						EmailChange  string `yaml:"emailChange,omitempty"`
						Confirmation string `yaml:"confirmation,omitempty"`
						MagicLink    string `yaml:"magicLink,omitempty"`
					} `yaml:"templates,omitempty"`
				} `yaml:"mailer,omitempty"`
				SMS struct {
					AutoConfirm bool   `yaml:"autoConfirm,omitempty"`
					OTPExp      string `yaml:"otpExp,omitempty"`
					OTPLength   int32  `yaml:"otpLength,omitempty"`
					Provider    string `yaml:"provider,omitempty"`
					Twilio      struct {
						AccountSID        string `yaml:"accountSID,omitempty"`
						AuthToken         string `yaml:"authToken,omitempty"`
						MessageServiceSID string `yaml:"messageServiceSID,omitempty"`
						ContentSID        string `yaml:"contentSID,omitempty"`
					} `yaml:"twilio,omitempty"`
					TwilioVerify struct {
						AccountSID        string `yaml:"accountSID,omitempty"`
						AuthToken         string `yaml:"authToken,omitempty"`
						MessageServiceSID string `yaml:"messageServiceSID,omitempty"`
					} `yaml:"twilioVerify,omitempty"`
					Messagebird struct {
						AccessKey string `yaml:"accessKey,omitempty"`
						Orginator string `yaml:"orginator,omitempty"`
					} `yaml:"messagebird,omitempty"`
					Vonage struct {
						APIKey    string `yaml:"apiKey,omitempty"`
						APISecret string `yaml:"apiSecret,omitempty"`
						From      string `yaml:"from,omitempty"`
					} `yaml:"vonage,omitempty"`
					TestOTP           string `yaml:"testOTP,omitempty"`
					TestOTPValidUntil string `yaml:"testOTPValidUntil,omitempty"`
					MaxFrequency      string `yaml:"maxFrequency,omitempty"`
				} `yaml:"sms,omitempty"`
				MFA struct {
					Enabled                     bool    `yaml:"enabled,omitempty"`
					ChallengeExpiryDuration     string  `yaml:"challengeExpiryDuration,omitempty"`
					RateLimitChallengeAndVerify float64 `yaml:"rateLimitChallengeAndVerify,omitempty"`
					MaxEnrolledFactors          float64 `yaml:"maxEnrolledFactors,omitempty"`
					MaxVerifiedFactors          int32   `yaml:"maxVerifiedFactors,omitempty"`
				} `yaml:"mfa,omitempty"`
				Captcha struct {
					Enabled  bool   `yaml:"enabled,omitempty"`
					Secret   string `yaml:"secret,omitempty"`
					Provider string `yaml:"provider,omitempty"`
				} `yaml:"captcha,omitempty"`
				RateLimitEmailSent    float64 `yaml:"rateLimitEmailSent,omitempty"`
				RateLimitSMSSent      float64 `yaml:"rateLimitSMSSent,omitempty"`
				RateLimitVerify       float64 `yaml:"rateLimitVerify,omitempty"`
				RateLimitTokenRefresh float64 `yaml:"rateLimitTokenRefresh,omitempty"`
				Security              struct {
					ManualLinkingEnabled bool `yaml:"manualLinkingEnabled,omitempty"`
				} `yaml:"security,omitempty"`
			} `yaml:"auth,omitempty"`

			Storage struct {
				TenantID string `yaml:"tenantID,omitempty"`
			} `yaml:"storage,omitempty"`

			Database struct {
				Schemas []string `yaml:"schemas,omitempty"`
			} `yaml:"database,omitempty"`

			Env []struct {
				Name  string `yaml:"name,omitempty"`
				Value string `yaml:"value,omitempty"`
			} `yaml:"env,omitempty"`
		} `yaml:"stack,omitempty"`
	}

	OAuthProviderConfig struct {
		Enabled        bool   `yaml:"enabled,omitempty"`
		Secret         string `yaml:"secret,omitempty"`
		ClientID       string `yaml:"clientID,omitempty"`
		SkipNonceCheck bool   `yaml:"skipNonceCheck,omitempty"`
	}
)

func (c *Cli) getServerUrl() string {
	if c.testServerUrl != "" {
		return c.testServerUrl
	}

	var scheme string
	if strings.Contains(c.args.Server, "local.shaple.io") {
		scheme = "http"
	} else {
		scheme = "https"
	}

	return scheme + "://" + c.args.Server
}
