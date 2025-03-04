package apidepotctl

import (
	"context"
	"fmt"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

func (c *Cli) newStackCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:     "stack",
		Short:   "Manage stacks",
		Aliases: []string{"stacks", "st"},
	}

	f := cmd.PersistentFlags()
	f.StringP("stack.name", "n", "", "Stack name")

	cmd.AddCommand(
		c.newGetStackCmd(),
		c.newUpdateStackCmd(),
		c.newStackVapiCmd(),
		c.newStackEnvCmd(),
		c.newStackCustomVapiCmd(),
	)

	return &cmd
}

func (c *Cli) newGetStackCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "get",
		Short: "Get stack",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			defer c.close()

			if err = c.readConfig(cmd.Flags()); err != nil {
				return errors.Wrapf(err, "failed to unmarshal stack")
			}

			if err = c.connectApiDepot(); err != nil {
				return
			}

			ctx, err = c.verifyCli(ctx)
			if err != nil {
				return
			}

			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("Stack: %s\n", st.Name)
			fmt.Printf("Stack ID: %d\n", st.Id)
			fmt.Printf("Stack URL: %s\n", st.Domain)
			fmt.Printf("Stack SiteURL: %s\n", st.SiteUrl)

			return nil
		},
	}

	return &cmd
}

func (c *Cli) newUpdateStackCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "update",
		Short: "Update stack",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer c.close()
			ctx := cmd.Context()

			if err = c.readConfig(cmd.Flags()); err != nil {
				return errors.Wrapf(err, "failed to unmarshal stack")
			}

			if err = c.connectApiDepot(); err != nil {
				return
			}

			ctx, err = c.verifyCli(ctx)
			if err != nil {
				return
			}

			st, err := c.getStack(ctx)
			if err != nil {
				return err
			}
			authConfig := c.args.Stack.Auth
			tcc := proto.NewApiDepotClient(c.conn)
			req := proto.InstallAuthRequest{
				Id:                               st.Id,
				IsUpdate:                         true,
				SmtpSenderName:                   &authConfig.SMTP.SenderName,
				SmtpAdminEmail:                   &authConfig.SMTP.AdminEmail,
				ExternalEmailEnabled:             &authConfig.External.EmailEnabled,
				ExternalPhoneEnabled:             &authConfig.External.PhoneEnabled,
				ExternalRedirectUrl:              &authConfig.External.RedirectURL,
				JwtExp:                           &authConfig.JWT.EXP,
				MailerAutoConfirm:                &authConfig.Mailer.AutoConfirm,
				MailerConfirmationSubject:        &authConfig.Mailer.Subjects.Confirmation,
				MailerRecoverySubject:            &authConfig.Mailer.Subjects.Recovery,
				MailerInviteSubject:              &authConfig.Mailer.Subjects.Invite,
				MailerEmailChangeSubject:         &authConfig.Mailer.Subjects.EmailChange,
				MailerMagicLinkSubject:           &authConfig.Mailer.Subjects.MagicLink,
				MailerRecoveryTemplate:           &authConfig.Mailer.Templates.Recovery,
				MailerInviteTemplate:             &authConfig.Mailer.Templates.Invite,
				MailerEmailChangeTemplate:        &authConfig.Mailer.Templates.EmailChange,
				MailerConfirmationTemplate:       &authConfig.Mailer.Templates.Confirmation,
				MailerMagicLinkTemplate:          &authConfig.Mailer.Templates.MagicLink,
				SmsAutoConfirm:                   &authConfig.SMS.AutoConfirm,
				SmsOtpExp:                        &authConfig.SMS.OTPExp,
				SmsOtpLength:                     &authConfig.SMS.OTPLength,
				SmsProvider:                      &authConfig.SMS.Provider,
				SmsTwilioAccountSid:              &authConfig.SMS.Twilio.AccountSID,
				SmsTwilioAuthToken:               &authConfig.SMS.Twilio.AuthToken,
				SmsTwilioMessageServiceSid:       &authConfig.SMS.Twilio.MessageServiceSID,
				SmsTwilioContentSid:              &authConfig.SMS.Twilio.ContentSID,
				SmsTwilioVerifyAccountSid:        &authConfig.SMS.TwilioVerify.AccountSID,
				SmsTwilioVerifyAuthToken:         &authConfig.SMS.TwilioVerify.AuthToken,
				SmsTwilioVerifyMessageServiceSid: &authConfig.SMS.TwilioVerify.MessageServiceSID,
				SmsMessagebirdAccessKey:          &authConfig.SMS.Messagebird.AccessKey,
				SmsMessagebirdOrginator:          &authConfig.SMS.Messagebird.Orginator,
				SmsVonageApiKey:                  &authConfig.SMS.Vonage.APIKey,
				SmsVonageApiSecret:               &authConfig.SMS.Vonage.APISecret,
				SmsVonageFrom:                    &authConfig.SMS.Vonage.From,
				SmsTestOtp:                       &authConfig.SMS.TestOTP,
				SmsTestOtpValidUntil:             &authConfig.SMS.TestOTPValidUntil,
				SmsMaxFrequency:                  &authConfig.SMS.MaxFrequency,
				ExternalIosBundleId:              &authConfig.External.ISOBundleID,
				MfaEnabled:                       &authConfig.MFA.Enabled,
				MfaChallengeExpiryDuration:       &authConfig.MFA.ChallengeExpiryDuration,
				MfaRateLimitChallengeAndVerify:   &authConfig.MFA.RateLimitChallengeAndVerify,
				MfaMaxEnrolledFactors:            &authConfig.MFA.MaxEnrolledFactors,
				MfaMaxVerifiedFactors:            &authConfig.MFA.MaxVerifiedFactors,
				CaptchaEnabled:                   &authConfig.Captcha.Enabled,
				CaptchaSecret:                    &authConfig.Captcha.Secret,
				CaptchaProvider:                  &authConfig.Captcha.Provider,
				RateLimitEmailSent:               &authConfig.RateLimitEmailSent,
				RateLimitSmsSent:                 &authConfig.RateLimitSMSSent,
				RateLimitVerify:                  &authConfig.RateLimitVerify,
				RateLimitTokenRefresh:            &authConfig.RateLimitTokenRefresh,
				SecurityManualLinkingEnabled:     &authConfig.Security.ManualLinkingEnabled,
			}
			oauthProviderArgs := map[string]*OAuthProviderConfig{
				"apple":         &authConfig.External.Apple,
				"azure":         &authConfig.External.Azure,
				"bitbucket":     &authConfig.External.Bitbucket,
				"discord":       &authConfig.External.Discord,
				"facebook":      &authConfig.External.Facebook,
				"figma":         &authConfig.External.Figma,
				"fly":           &authConfig.External.Fly,
				"github":        &authConfig.External.Github,
				"gitlab":        &authConfig.External.Gitlab,
				"google":        &authConfig.External.Google,
				"kakao":         &authConfig.External.Kakao,
				"notion":        &authConfig.External.Notion,
				"keycloak":      &authConfig.External.Keycloak,
				"linkedin":      &authConfig.External.Linkedin,
				"linkedin_oidc": &authConfig.External.LinkedinOIDC,
				"spotify":       &authConfig.External.Spotify,
				"slack":         &authConfig.External.Slack,
				"twitter":       &authConfig.External.Twitter,
				"twitch":        &authConfig.External.Twitch,
				"workOS":        &authConfig.External.WorkOS,
				"zoom":          &authConfig.External.Zoom,
			}
			req.ExternalOauthProviders = make([]*proto.AuthExternalOAuthProvider, 0, len(oauthProviderArgs))
			for key, arg := range oauthProviderArgs {
				if arg.Enabled {
					req.ExternalOauthProviders = append(req.ExternalOauthProviders, &proto.AuthExternalOAuthProvider{
						Secret:         &arg.Secret,
						ClientId:       &arg.ClientID,
						Enabled:        &arg.Enabled,
						Name:           &key,
						SkipNonceCheck: &arg.SkipNonceCheck,
					})
				}
			}
			if _, err = tcc.InstallAuth(ctx, &req); err != nil {
				return errors.Wrapf(err, "failed to install auth")
			}

			if _, err = tcc.InstallStorage(ctx, &proto.InstallStorageRequest{
				Id:       st.Id,
				TenantId: &c.args.Stack.Storage.TenantID,
				IsUpdate: true,
			}); err != nil {
				return errors.Wrapf(err, "failed to install storage")
			}

			if _, err = tcc.InstallPostgrest(ctx, &proto.InstallPostgrestRequest{
				Id:       st.Id,
				IsUpdate: true,
				Schemas:  c.args.Stack.Database.Schemas,
			}); err != nil {
				return errors.Wrapf(err, "failed to install postgrest")
			}

			if _, err = tcc.UpdateStack(ctx, &proto.UpdateStackRequest{
				StackId:     st.Id,
				SiteUrl:     &c.args.Stack.SiteUrl,
				Description: &c.args.Stack.Description,
			}); err != nil {
				return errors.Wrapf(err, "failed to update stack")
			}

			if len(c.args.Stack.Env) > 0 {
				var envVars []*proto.SetStackEnvRequest_EnvVar
				for _, env := range c.args.Stack.Env {
					envVars = append(envVars, &proto.SetStackEnvRequest_EnvVar{
						Name:  env.Name,
						Value: env.Value,
					})
				}
				if _, err = tcc.SetStackEnv(ctx, &proto.SetStackEnvRequest{
					StackId: st.Id,
					EnvVars: envVars,
				}); err != nil {
					return errors.Wrapf(err, "failed to set vapi env")
				}
			}

			return c.writeConfig()
		},
	}

	f := cmd.Flags()
	f.String("stack.siteUrl", "", "Change siteUrl for stack")
	f.String("stack.description", "", "Change siteUrl for stack")

	f.String("stack.auth.smtp.senderName", "Shaple", "Specify sender name for auth")
	f.Bool("stack.auth.external.emailEnabled", false, "Enable email for auth")
	f.Bool("stack.auth.external.phoneEnabled", false, "Enable phone for auth")
	f.String("stack.auth.external.iosBundleId", "", "Specify ios bundle id for auth")

	oauthServiceNames := []string{
		"Apple",
		"Azure",
		"Bitbucket",
		"Discord",
		"Facebook",
		"Figma",
		"Github",
		"Gitlab",
		"Google",
		"Kakao",
		"Notion",
		"Keycloak",
		"LinkedIn",
		"LinkedInOIDC",
		"Spotify",
		"Slack",
		"Twitter",
		"Twitch",
		"WorkOS",
		"Zoom",
	}
	for _, serviceName := range oauthServiceNames {
		keyName := strings.ToLower(serviceName)
		f.Bool("stack.auth.external."+keyName+".enabled", false, "Enable "+serviceName+" oauth for auth")
		f.String("stack.auth.external."+keyName+".secret", "", "Specify "+serviceName+" secret for auth")
		f.String("stack.auth.external."+keyName+".clientID", "", "Specify "+serviceName+" client id for auth")
		f.Bool("stack.auth.external."+keyName+".skipNonceCheck", false, "Enable "+serviceName+" to skip nonce check for auth")
	}
	f.String("stack.auth.external.redirectURL", "", "Specify redirect url for auth")
	f.String("stack.auth.jwt.exp", "", "Specify jwt exp(seconds) for auth")
	f.Bool("stack.auth.mailer.autoConfirm", false, "Specify auto confirm for mailer")
	f.Bool("stack.auth.sms.autoConfirm", false, "Specify auto confirm for SMS")
	f.String("stack.auth.sms.maxFrequency", "60s", "Specify max frequency for SMS")
	f.String("stack.auth.sms.otpExp", "", "Specify otp exp in duration format for SMS")
	f.Int("stack.auth.sms.otpLength", 0, "Specify otp length for SMS")
	f.String("stack.auth.sms.provider", "", "Specify provider for SMS")
	f.String("stack.auth.sms.twilio.AccountSID", "", "Specify Twilio account sid for SMS")
	f.String("stack.auth.sms.twilio.AuthToken", "", "Specify Twilio auth token for SMS")
	f.String("stack.auth.sms.twilio.MessageServiceSID", "", "Specify Twilio message service sid for SMS")
	f.String("stack.auth.sms.twilio.ContentSID", "", "Specify Twilio content sid for SMS")
	f.String("stack.auth.sms.twilioVerify.AccountSID", "", "Specify Twilio Verify account sid for SMS")
	f.String("stack.auth.sms.twilioVerify.AuthToken", "", "Specify Twilio Verify auth token for SMS")
	f.String("stack.auth.sms.twilioVerify.MessageServiceSID", "", "Specify Twilio Verify message service sid for SMS")
	f.String("stack.auth.sms.messagebird.accessKey", "", "Specify MessageBird access key for SMS")
	f.String("stack.auth.sms.messagebird.originator", "", "Specify MessageBird originator for SMS")
	f.String("stack.auth.sms.vonage.apiKey", "", "Specify Vonage api key for SMS")
	f.String("stack.auth.sms.vonage.apiSecret", "", "Specify Vonage api secret for SMS")
	f.String("stack.auth.sms.vonage.from", "", "Specify Vonage from for SMS")
	f.String("stack.auth.sms.testOTP", "", "Specify test OTP for SMS. Format: <phone>:<otp>,<phone>:<otp>,...")
	f.String("stack.auth.sms.testOTPValidUntil", "", "Specify test OTP valid until for SMS. Format: 2006-01-02T15:04:05+07:00")
	f.Bool("stack.auth.mfa.enabled", false, "Specify enabled for MFA")
	f.String("stack.auth.mfa.challengeExpiryDuration", "", "Specify challenge expiry duration for MFA")
	f.Float64("stack.auth.mfa.rateLimitChallengeAndVerify", 0.0, "Specify rate limit challenge and verify for MFA")
	f.Float64("stack.auth.mfa.maxEnrolledFactors", 0.0, "Specify max enrolled factors for MFA")
	f.Int("stack.auth.mfa.maxVerifiedFactors", 0, "Specify max verified factors for MFA")
	f.Bool("stack.auth.captcha.enabled", false, "Specify enabled for captcha")
	f.String("stack.auth.captcha.secret", "", "Specify secret for captcha")
	f.String("stack.auth.captcha.provider", "", "Specify provider for captcha")
	f.Float64("stack.auth.rateLimitEmailSent", 30.0, "Specify rate limit email sent for auth")
	f.Float64("stack.auth.rateLimitSMSSent", 30.0, "Specify rate limit sms sent for auth")
	f.Float64("stack.auth.rateLimitVerify", 30.0, "Specify rate limit verify for auth in seconds")
	f.Float64("stack.auth.rateLimitTokenRefresh", 150.0, "Specify rate limit token refresh for auth in seconds")
	f.Bool("stack.auth.security.manualLinkingEnabled", false, "Specify manual linking enabled for security")
	f.StringToString("stack.vapi.env", nil, "Specify env var for vapi")
	f.String("stack.storage.tenantID", "", "Specify tenant id for storage")
	f.StringSlice("stack.database.schemas", nil, "Specify comma seperated schemas for database")

	return &cmd
}

func (c *Cli) getStack(ctx context.Context) (*proto.Stack, error) {
	tcc := proto.NewApiDepotClient(c.conn)

	var project *proto.Project
	if resp, err := tcc.GetProjects(ctx, &proto.GetProjectsRequest{}); err != nil {
		return nil, errors.WithStack(err)
	} else if len(resp.Projects) == 0 {
		return nil, errors.New("no stack found")
	} else {
		project = resp.Projects[0]
	}

	var stack *proto.Stack
	if resp, err := tcc.GetStacks(ctx, &proto.GetStacksRequest{
		Name:      &c.args.Stack.Name,
		ProjectId: project.Id,
	}); err != nil {
		return nil, errors.WithStack(err)
	} else if len(resp.Stacks) == 0 {
		return nil, errors.New("no stack found")
	} else {
		stack = resp.Stacks[0]
	}

	return stack, nil
}
