package services

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v60/github"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const ServiceKeyGithubClient = "githubClient"

type GithubClient interface {
	GetUser(
		ctx context.Context,
		token string,
	) (*github.User, error)
	FetchUserAccessToken(
		ctx context.Context,
		authorizationCode string,
	) (string, error)
	VerifyInstallation(
		ctx context.Context,
		accessToken string,
		installationId int64,
	) (bool, error)
	GenerateInstallationAccessToken(
		ctx context.Context,
		installationId int64,
	) (string, error)
	GetExistingInstallationId(
		ctx context.Context,
		accessToken string,
	) (int64, error)
	CreateRepository(
		ctx context.Context,
		accessToken string,
		org string,
		repoName string,
		description string,
	) error
}

type githubClient struct {
	appId         string
	clientId      string
	clientSecret  string
	b64PrivateKey string
}

func NewGithubClient(appId string, clientId, clientSecret string, b64PrivateKey string) GithubClient {
	return &githubClient{
		appId:         appId,
		clientId:      clientId,
		clientSecret:  clientSecret,
		b64PrivateKey: b64PrivateKey,
	}
}

func (g *githubClient) GetUser(
	ctx context.Context,
	token string,
) (*github.User, error) {
	githubClient := github.NewClient(nil).WithAuthToken(token)
	user, resp, err := githubClient.Users.Get(ctx, "")
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return nil, errors.Wrapf(err, "failed request to get user")
	}

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("failed to get user. statusCode=%d", resp.StatusCode)
	}

	logger.Debug("got user", "user", user)

	return user, nil
}

func (g *githubClient) FetchUserAccessToken(ctx context.Context, authorizationCode string) (string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"client_id":     g.clientId,
		"client_secret": g.clientSecret,
		"code":          authorizationCode,
	})

	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal request body")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to send request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != 200 {
		return "", errors.Errorf("failed to fetch access token. statusCode=%d", resp.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if bodyBytes, err := io.ReadAll(resp.Body); err != nil {
		return "", errors.Wrapf(err, "failed to read response body")
	} else if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return "", errors.Wrapf(err, "failed to decode response body=%s", string(bodyBytes))
	} else if body.AccessToken == "" {
		return "", errors.Errorf("failed to get access token. body=%s", string(bodyBytes))
	}

	return body.AccessToken, nil
}

func (g *githubClient) VerifyInstallation(ctx context.Context, accessToken string, installationId int64) (bool, error) {
	user, err := g.GetUser(ctx, accessToken)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get user")
	}

	existingId, err := g.GetExistingInstallationId(ctx, user.GetLogin())
	if err != nil {
		return false, errors.Wrapf(err, "failed to get existing installation id")
	}

	return existingId == installationId, nil
}

func (g *githubClient) GetExistingInstallationId(ctx context.Context, username string) (int64, error) {
	jwt, err := g.generateJWT()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to generate JWT")
	}
	githubClient := github.NewClient(nil).WithAuthToken(jwt)
	installation, resp, err := githubClient.Apps.FindUserInstallation(ctx, username)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return 0, errors.Wrapf(err, "failed request to get installations")
	}

	if resp.StatusCode != 200 {
		return 0, errors.Errorf("failed to get installations. statusCode=%d", resp.StatusCode)
	}

	return installation.GetID(), nil
}

func (g *githubClient) GenerateInstallationAccessToken(ctx context.Context, installationId int64) (string, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationId)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create request")
	}

	jwt, err := g.generateJWT()
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate JWT")
	}
	logger.Info("created jwt", "jwt", jwt)
	// Add JWT to the Authorization header
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.Errorf("failed to create access token: %s", string(body))
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", errors.Wrapf(err, "failed to decode response body")
	}

	accessToken, ok := body["token"]
	if !ok {
		return "", errors.New("failed to get access token")
	}

	return accessToken.(string), nil
}

// generateJWT generates a JWT for GitHub App authentication using a Base64-encoded private key
func (g *githubClient) generateJWT() (string, error) {
	// Decode Base64-encoded private key
	privateKeyData, err := base64.StdEncoding.DecodeString(g.b64PrivateKey)
	if err != nil {
		return "", errors.Wrapf(err, "failed to decode base64 private key")
	}

	// Decode PEM block
	block, _ := pem.Decode(privateKeyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return "", errors.Errorf("failed to decode PEM block containing private key")
	}

	// Parse the RSA private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse private key")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": time.Now().Unix(),                       // Issued at
		"exp": time.Now().Add(10 * time.Minute).Unix(), // Expiration time
		"iss": g.appId,                                 // GitHub App ID
	})

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.Wrapf(err, "failed to sign token")
	}

	return signedToken, nil
}

func (g *githubClient) CreateRepository(
	ctx context.Context,
	accessToken string,
	org string,
	repoName string,
	description string,
) error {
	user, err := g.GetUser(ctx, accessToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get user")
	}

	githubClient := github.NewClient(nil).WithAuthToken(accessToken)
	if strings.ToLower(org) == *user.Login {
		org = ""
	}

	repo, resp, err := githubClient.Repositories.Create(ctx, org, &github.Repository{
		Name:        github.String(repoName),
		Description: github.String(description),
		AutoInit:    github.Bool(true),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create repository")
	}

	if resp.StatusCode != 201 {
		return errors.Errorf("failed to create repository. statusCode=%d", resp.StatusCode)
	}

	logger.Debug("created repository", "repo", repo)

	return nil
}

func init() {
	digo.ProvideService(ServiceKeyGithubClient, func(ctx *digo.Container) (any, error) {
		switch ctx.Env {
		case digo.EnvTest:
			return NewGithubClient("", "", "", ""), nil
		case digo.EnvProd:
			return NewGithubClient(
				ctx.Config.Github.AppId,
				ctx.Config.Github.ClientId,
				ctx.Config.Github.ClientSecret,
				ctx.Config.Github.AppPrivateKey), nil
		default:
			return nil, errors.Errorf("unsupported environment: %s", ctx.Env)
		}
	})
}
