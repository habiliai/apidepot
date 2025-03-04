package usertest

import (
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/gotrue-go"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
	"testing"
)

var session *gotruetypes.Session

func SignIn(t *testing.T, gotrueClient gotrue.Client) *gotruetypes.TokenResponse {
	if session != nil {
		SignOut(t, gotrueClient)
	}
	_, err := gotrueClient.Signup(gotruetypes.SignupRequest{
		Email:    "test1@cloudtest.com",
		Password: "qwer1234",
	})
	require.NoError(t, err)

	signinResp, err := gotrueClient.SignInWithEmailPassword(
		"test1@cloudtest.com",
		"qwer1234",
	)
	require.NoError(t, err)

	session = &signinResp.Session

	return signinResp
}

func SignOut(t *testing.T, gotrueClient gotrue.Client) {
	if session == nil {
		return
	}
	require.NoError(t, gotrueClient.WithToken(session.AccessToken).Logout())
	session = nil
}
