package instance_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/functions"
	"github.com/habiliai/apidepot/pkg/internal/instance"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/habiliai/apidepot/pkg/internal/util"
	vapitest "github.com/habiliai/apidepot/pkg/internal/vapi/test"
	"github.com/jackc/pgx/v5"
	"github.com/mitchellh/mapstructure"
	"github.com/mokiat/gog"
	"github.com/supabase-community/postgrest-go"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func (s *InstanceServiceTestSuite) TestGivenInstalledAuthStackWhenInstanceServiceDeployStackShouldBeOKToSignUp() {
	st := s.stack

	inst, err := s.instances.CreateInstance(s, instance.CreateInstanceInput{
		StackID: st.ID,
		Zone:    tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)
	s.Require().NotNil(inst)
	defer func() {
		s.NoError(s.instances.DeleteInstance(s, inst.ID, true))
	}()

	s.T().Log("-- enable auth")
	st = s.installAuth(s, st)
	defer s.uninstallAuth(s, st)

	s.Require().NoError(s.instances.DeployStack(s, inst.ID, instance.DeployStackInput{
		Timeout: gog.PtrOf("60s"),
	}))

	k8sClient, err := s.k8sClientPool.GetClient(inst.Zone)
	s.Require().NoError(err)

	object, err := k8sClient.GetResource(s, "deployment", "auth", st.Namespace())
	s.Require().NoError(err)

	s.Require().NotEmpty(object.GetName())
	s.Require().Equal("auth", object.GetName())
	s.Require().Equal(st.Namespace(), object.GetNamespace())

	inst, err = domain.FindInstanceById(s.db, inst.ID)
	s.Require().NoError(err)
	s.Require().Equal(domain.InstanceStateRunning, inst.State)

	s.T().Log("-- sign up by auth")
	s.signUp(s, st)
}

func (s *InstanceServiceTestSuite) installAuth(ctx context.Context, st *domain.Stack) *domain.Stack {
	require := s.Require()

	validUntilStr := "2024-01-23T23:00:00+09:00"
	_, err := time.Parse(time.RFC3339, validUntilStr)
	require.NoError(err, "failed to parse time. format: %s", time.Now().Format(time.RFC3339))

	s.Require().NoError(s.stacks.EnableOrUpdateAuth(ctx, st.ID, stack.EnableOrUpdateAuthInput{
		AuthInput: stack.AuthInput{
			SenderName:                   gog.PtrOf("Shaple"),
			AdminEmail:                   gog.PtrOf("shaple@shaple.io"),
			PhoneEnabled:                 gog.PtrOf(true),
			EmailEnabled:                 gog.PtrOf(true),
			Exp:                          gog.PtrOf("1h"),
			MailerAutoConfirm:            gog.PtrOf(true),
			MailerInviteSubject:          gog.PtrOf("Shaple Invite"),
			MailerInviteTemplate:         gog.PtrOf("Hello, {{.Email}}! Link: {{.SiteURL}}"),
			MailerConfirmationSubject:    gog.PtrOf("Shaple Confirmation"),
			MailerConfirmationTemplate:   gog.PtrOf("Hello, {{.Email}}! Please confirm your email."),
			MailerRecoverySubject:        gog.PtrOf("Shaple Recovery"),
			MailerRecoveryTemplate:       gog.PtrOf("Hello, {{.Email}}! Please recover your password."),
			MailerEmailChangeSubject:     gog.PtrOf("Shaple Email Change"),
			MailerEmailChangeTemplate:    gog.PtrOf("Hello, {{.Email}}! Please confirm your email change."),
			MailerMagicLinkSubject:       gog.PtrOf("Shaple Magic Link"),
			MailerMagicLinkTemplate:      gog.PtrOf("Hello, {{.Email}}! Please confirm your magic link."),
			SMSAutoConfirm:               gog.PtrOf(true),
			TestOTP:                      gog.PtrOf("8201012341234:123456,8201034563456:123123"),
			TestOTPValidUntil:            &validUntilStr,
			SecurityManualLinkingEnabled: gog.PtrOf(true),
		},
	}, true))
	time.Sleep(250 * time.Millisecond)

	st, err = domain.FindStackByID(s.db, st.ID)
	s.Require().NoError(err)

	return st
}

func (s *InstanceServiceTestSuite) uninstallAuth(ctx context.Context, stack *domain.Stack) {
	s.T().Log("-- disable auth")
	s.NoError(s.stacks.DisableAuth(ctx, stack.ID))
}

func (s *InstanceServiceTestSuite) signUp(
	ctx context.Context, stack *domain.Stack,
) (string, string, string) {
	k8sClient, err := s.k8sClientPool.GetClient(tcltypes.InstanceZoneDefault)
	s.Require().NoError(err)

	res, err := k8sClient.GetResource(ctx, "deployment", "auth", stack.Namespace())
	s.Require().NoError(err)
	s.T().Logf("object: %+v", res.Object)

	input := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "dennis.park@paust.io",
		Password: "1q2w#E$R",
	}
	reqBody, err := json.Marshal(input)
	s.Require().NoError(err)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/auth/v1/signup", stack.Endpoint()), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	s.Require().NoError(err)
	defer req.Body.Close()

	s.T().Logf("req: %+v", req)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	output := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID               string    `json:"id"`
			Email            string    `json:"email"`
			EmailConfirmedAt time.Time `json:"email_confirmed_at"`
			CreatedAt        time.Time `json:"created_at"`
			UpdatedAt        time.Time `json:"updated_at"`
		} `json:"user"`
	}{}
	respBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	s.T().Logf("output: %s", string(respBytes))
	s.Require().NoError(json.Unmarshal(respBytes, &output))

	userId := output.User.ID
	accessToken := output.AccessToken
	refreshToken := output.RefreshToken

	s.NotEmpty(userId)
	s.NotEmpty(accessToken)
	s.NotEmpty(refreshToken)
	s.Equal(input.Email, output.User.Email)
	s.Less(time.Now().UTC().Add(-24*time.Hour), output.User.EmailConfirmedAt)
	s.Less(time.Now().UTC().Add(-24*time.Hour), output.User.UpdatedAt)

	return userId, accessToken, refreshToken
}

func (s *InstanceServiceTestSuite) TestShapleServicePostgrest() {
	ctx := s.Context
	st := s.stack

	k8sClient, err := s.k8sClientPool.GetClient(tcltypes.InstanceZoneDefault)
	s.Require().NoError(err)

	inst, err := s.instances.CreateInstance(ctx, instance.CreateInstanceInput{
		StackID: st.ID,
		Zone:    tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)
	defer func() {
		s.NoError(s.instances.DeleteInstance(ctx, inst.ID, true))
	}()

	st = s.installAuth(ctx, st)
	defer s.uninstallAuth(ctx, st)
	s.Require().NoError(s.instances.DeployStack(ctx, inst.ID, instance.DeployStackInput{
		Timeout: gog.PtrOf("60s"),
	}))

	userId, accessToken, _ := s.signUp(ctx, st)

	{
		s.T().Log("-- enable postgrest")
		s.Require().NoError(s.stacks.EnableOrUpdatePostgrest(ctx, st.ID, stack.EnableOrUpdatePostgrestInput{
			PostgrestInput: stack.PostgrestInput{
				Schemas: []string{"api", "public"},
			},
		}, true))
		s.Require().NoError(s.instances.DeployStack(ctx, inst.ID, instance.DeployStackInput{
			Timeout: gog.PtrOf("60s"),
		}))

		object, err := k8sClient.GetResource(ctx, "deployment", "postgrest", st.Namespace())
		s.Require().NoError(err)

		s.Equal("postgrest", object.GetName())
	}

	{
		s.T().Log("-- migrate database")
		s.Require().NoError(s.stacks.MigrateDatabase(ctx, st.ID, stack.MigrateDatabaseInput{
			Migrations: []stack.Migration{
				{
					Version: time.Now(),
					Query: `
CREATE TABLE IF NOT EXISTS api.test_table (id SERIAL PRIMARY KEY, name VARCHAR(255), user_id UUID REFERENCES auth.users(id));
ALTER TABLE api.test_table ENABLE ROW LEVEL SECURITY;
CREATE POLICY test_table_policy ON api.test_table FOR ALL TO authenticated USING (
	user_id = auth.uid()
);
`,
				},
			},
		}))
	}

	{
		s.T().Log("using postgrest with rls")
		stack, err := domain.FindStackByID(s.db, st.ID)
		s.Require().NoError(err)
		data := struct {
			Name   string `json:"name"`
			UserId string `json:"user_id"`
		}{
			Name:   "test123",
			UserId: userId,
		}

		client := postgrest.NewClient(stack.Endpoint()+"/postgrest/v1", "api", nil).SetAuthToken(accessToken)
		s.Require().NoError(client.ClientError)
		if !s.True(client.Ping()) {
			s.FailNow("failed to ping")
		}

		{
			resp, _, err := client.From("test_table").Select("", "", false).ExecuteString()
			s.Require().NoError(err)
			s.T().Logf("select resp: %s", resp)
		}
		{
			resp, _, err := client.From("test_table").Insert(data, false, "", "", "").ExecuteString()
			s.Require().NoError(err)
			s.T().Logf("insert resp: %s", resp)
		}
		{
			resp, _, err := client.From("test_table").Select("", "", false).ExecuteString()
			s.Require().NoError(err)
			s.T().Logf("select resp: %s", resp)
			var output []struct {
				Name   string `json:"name"`
				UserId string `json:"user_id"`
			}
			s.NoError(json.Unmarshal([]byte(resp), &output))
			if s.Len(output, 1) {
				s.Equal(data.Name, output[0].Name)
				s.Equal(data.UserId, output[0].UserId)
			}
		}
	}

}

func (s *InstanceServiceTestSuite) TestShapleServiceStorage() {
	ctx := s.Context
	st := s.stack

	inst, err := s.instances.CreateInstance(ctx, instance.CreateInstanceInput{
		StackID: st.ID,
		Zone:    tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)
	defer func() {
		s.NoError(s.instances.DeleteInstance(ctx, inst.ID, true))
	}()

	st = s.installAuth(ctx, st)
	defer s.uninstallAuth(ctx, st)

	st = s.installStorage(ctx, st)
	defer s.uninstallStorage(ctx, st)

	err = s.instances.DeployStack(ctx, inst.ID, instance.DeployStackInput{
		Timeout: gog.PtrOf("60s"),
	})
	s.Require().NoError(err)

	k8sClient, err := s.k8sClientPool.GetClient(tcltypes.InstanceZoneDefault)
	s.Require().NoError(err)

	{
		s.T().Log("-- create bucket and upload object")
		require := s.Require()
		res, err := k8sClient.GetResource(ctx, "deployment", "storage", st.Namespace())
		require.NoError(err)
		s.T().Logf("object: %+v", res.Object)

		stack, err := domain.FindStackByID(s.db, st.ID)
		require.NoError(err)

		{
			s.T().Log("-- create bucket")
			inputs := struct {
				Name   string `json:"name"`
				Public bool   `json:"public"`
			}{
				Name:   "test-bucket",
				Public: false,
			}
			reqBody, err := json.Marshal(inputs)
			require.NoError(err)
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/storage/v1/bucket", stack.Endpoint()), bytes.NewReader(reqBody))
			require.NoError(err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+stack.AdminApiKey)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(err)
			defer resp.Body.Close()

			s.Equal(http.StatusOK, resp.StatusCode)
			respBody, err := io.ReadAll(resp.Body)
			if s.NoError(err) {
				s.T().Logf("respBody: %s", string(respBody))
			}
		}

		{
			s.T().Log("-- upload object")
			// get checksum of the file
			file, err := os.Open("testdata/images/159.jpg")
			require.NoError(err)
			defer file.Close()
			var fileData bytes.Buffer
			_, err = io.Copy(&fileData, file)
			checksum, err := util.GetChecksum(fileData.Bytes())
			require.NoError(err)

			var requestBody bytes.Buffer
			mw := multipart.NewWriter(&requestBody)
			fw, err := mw.CreateFormFile("file", "a.jpg")
			require.NoError(err)
			written, err := io.Copy(fw, &fileData)
			require.NoError(err)
			require.NotZero(written)
			require.NoError(mw.Close())

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/storage/v1/object/test-bucket/img/a.jpg", stack.Endpoint()),
				&requestBody,
			)
			require.NoError(err)
			req.Header.Set("Content-Type", mw.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+stack.AdminApiKey)
			defer req.Body.Close()

			resp, err := http.DefaultClient.Do(req)
			s.T().Logf("resp: %+v", resp)
			require.NoError(err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			s.Require().NoError(err)
			s.T().Logf("body: %s", body)

			require.Equal(http.StatusOK, resp.StatusCode)

			var respBody struct {
				Id  string `json:"id"`
				Key string `json:"key"`
			}
			s.NoError(json.Unmarshal(body, &respBody))

			s.NotEmpty(respBody.Id)
			s.NotEmpty(respBody.Key)
			s.T().Logf("respBody: %+v", respBody)

			// get object
			req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/storage/v1/object/test-bucket/img/a.jpg", stack.Endpoint()), nil)
			require.NoError(err)
			req.Header.Set("Authorization", "Bearer "+stack.AdminApiKey)

			resp, err = http.DefaultClient.Do(req)
			require.NoError(err)
			defer resp.Body.Close()

			require.Equal(http.StatusOK, resp.StatusCode)
			var respBytes []byte
			respBytes, err = io.ReadAll(resp.Body)
			actualChecksum, err := util.GetChecksum(respBytes)
			require.NoError(err)
			require.Equal(checksum, actualChecksum)
		}
	}
}

func (s *InstanceServiceTestSuite) installStorage(ctx context.Context, st *domain.Stack) *domain.Stack {
	s.T().Log("-- enable storage")
	s.Require().NoError(s.stacks.EnableOrUpdateStorage(ctx, st.ID, stack.EnableOrUpdateStorageInput{
		StorageInput: stack.StorageInput{
			TenantID: gog.PtrOf("root"),
		},
	}, true))

	time.Sleep(250 * time.Millisecond)

	st, err := s.stacks.GetStack(ctx, st.ID)
	s.Require().NoError(err)

	return st
}

func (s *InstanceServiceTestSuite) uninstallStorage(ctx context.Context, stack *domain.Stack) {
	s.T().Log("-- disable storage")
	s.NoError(s.stacks.DisableStorage(ctx, stack.ID))
}

func (s *InstanceServiceTestSuite) TestInstallVapi() {
	ctx := s.Context

	inst, err := s.instances.CreateInstance(ctx, instance.CreateInstanceInput{
		StackID: s.stack.ID,
		Zone:    tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)
	defer s.instances.DeleteInstance(ctx, inst.ID, true)

	s.stack = s.installVapis(ctx, s.stack.ID)
	defer s.uninstallVapis(ctx, s.stack.ID)

	s.Require().NoError(s.stacks.SetVapiEnv(ctx, s.stack.ID, map[string]string{
		"helloworld.HELLO": "world",
		"sns.THIS_ENV":     "test",
	}))
	defer s.stacks.UnsetVapiEnv(ctx, s.stack.ID, []string{"helloworld.HELLO", "sns.THIS_ENV"})

	s.Require().NoError(s.instances.DeployStack(ctx, inst.ID, instance.DeployStackInput{
		Timeout: gog.PtrOf("60s"),
	}))
	{
		require := s.Require()
		s.T().Log("-- invoke helloworld/hello function")

		client := functions.NewClient(s.stack.Endpoint()+constants.PathVapis, s.stack.AdminApiKey, nil)
		resp := client.Invoke("helloworld/hello", functions.FunctionInvokeOptions{
			Body:         bytes.NewBufferString(`{"name":"test"}`),
			ResponseType: "json",
		})
		require.NoErrorf(resp.Error, "resp: %+v, error: %+v", resp.Data, resp.Error)

		if !s.Equal(http.StatusOK, resp.Status) {
			s.FailNow("TODO: check response body")
		}

		var data struct {
			Message string `json:"message"`
		}
		require.NoError(mapstructure.Decode(resp.Data, &data))

		s.Equal("Hello test!", data.Message)
	}
	{
		require := s.Require()
		s.T().Log("-- invoke helloworld/world function")

		client := functions.NewClient(s.stack.Endpoint()+constants.PathVapis, s.stack.AdminApiKey, nil)
		resp := client.Invoke("helloworld/world", functions.FunctionInvokeOptions{
			Body:         bytes.NewBufferString(`{"name":"test"}`),
			ResponseType: "json",
		})
		require.NoError(resp.Error)

		if !s.Equal(http.StatusOK, resp.Status) {
			s.FailNow("TODO: check response body")
		}

		var data struct {
			Message string `json:"message"`
		}
		require.NoError(mapstructure.Decode(resp.Data, &data))

		s.Equal("World test!", data.Message)
	}
	{
		require := s.Require()
		s.T().Log("-- invoke sns/v1/get_feed function")

		_, err := domain.FindVapiReleaseByPackageNameAndVersion(s.db, "sns", "1.2.0")
		require.NoError(err)

		client := functions.NewClient(s.stack.Endpoint()+constants.PathVapis, s.stack.AdminApiKey, nil)
		resp := client.Invoke("sns/v1/get_feed", functions.FunctionInvokeOptions{
			Body:         bytes.NewBufferString(`{"name":"test"}`),
			ResponseType: "blob",
		})

		s.Require().NoError(resp.Error)

		if !s.Equal(http.StatusOK, resp.Status) {
			s.FailNow("TODO: check response body")
		}

		respBody, ok := resp.Data.([]byte)
		s.True(ok)

		var data struct {
			Message string `json:"message"`
		}
		s.Require().NoErrorf(json.Unmarshal(respBody, &data), "respBody: %s", respBody)

		s.Equal("Feed for test!", data.Message)
	}
}

func (s *InstanceServiceTestSuite) installVapis(
	ctx context.Context,
	stackId uint,
) *domain.Stack {
	s.T().Logf("-- enable vapi")
	var err error

	_, rel2 := vapitest.RegisterVapis(s.T(), s.Context, s.vapis)

	st, err := domain.FindStackByID(s.db, stackId)
	s.Require().NoError(err)

	_, err = s.stacks.EnableVapi(
		ctx,
		st.ID,
		stack.EnableVapiInput{
			VapiID: rel2.ID,
		},
	)
	s.Require().NoError(err)

	s.T().Logf("stack: %s", st.String())

	st, err = s.stacks.GetStack(ctx, st.ID)
	s.Require().NoError(err)

	return st
}

func (s *InstanceServiceTestSuite) uninstallVapis(
	ctx context.Context,
	stackId uint,
) {
	s.T().Log("-- disable vapi")

	stack, err := domain.FindStackByID(s.db, stackId)
	s.Require().NoError(err)

	for _, stackVapi := range stack.Vapis {
		s.NoError(s.stacks.DisableVapi(
			ctx,
			stackId,
			stackVapi.VapiID,
		))
	}
}

func (s *InstanceServiceTestSuite) TestMigrateVapiDatabase() {
	ctx, cancel := context.WithCancel(s)
	defer cancel()

	// Given
	inst, err := s.instances.CreateInstance(ctx, instance.CreateInstanceInput{
		StackID: s.stack.ID,
		Zone:    tcltypes.InstanceZoneDefault,
	})
	s.Require().NoError(err)
	defer func() {
		s.NoError(s.instances.DeleteInstance(ctx, inst.ID, true))
	}()
	s.installVapis(ctx, s.stack.ID)
	defer s.uninstallVapis(ctx, s.stack.ID)

	// When
	s.Require().NoError(s.instances.DeployStack(ctx, inst.ID, instance.DeployStackInput{
		Timeout: gog.PtrOf("60s"),
	}))
	conn, err := pgx.Connect(
		ctx,
		fmt.Sprintf(
			"postgres://postgres:postgres@localhost:6543/%s?search_path=helloworld&sslmode=disable",
			s.stack.DB.Data().Name,
		),
	)
	s.Require().NoError(err)
	defer conn.Close(ctx)

	s.Require().NoError(conn.Ping(ctx))

	_, err = conn.Exec(
		ctx,
		"INSERT INTO profiles(name) VALUES('test');",
	)

	// Then
	s.Require().NoError(err)
	var (
		id   uint
		name string
	)
	s.Require().NoError(conn.QueryRow(
		ctx,
		"SELECT id, name FROM profiles LIMIT 1",
	).Scan(&id, &name))

	s.Equal("test", name)
	s.NotEqual(uint(0), id)
}
