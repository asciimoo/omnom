package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	scopeReadUser  ScopeValue = "read:user"
	scopeUserEmail ScopeValue = "user:email"
)

// GitHubOAuth implements OAuth 2.0 authentication for GitHub.
type GitHubOAuth struct {
	AuthURL  string
	TokenURL string
}

// Prepare initializes the GitHub OAuth provider. No preparation needed for GitHub.
func (g GitHubOAuth) Prepare(_ context.Context, _ *PrepareRequest) error { return nil }

// GetRedirectURL constructs the GitHub authorization URL for redirecting users.
func (g GitHubOAuth) GetRedirectURL(req *RedirectURIRequest) string {
	params := &url.Values{}

	params.Add("client_id", req.clientID)
	params.Add("scope", scopeReadUser.String())
	params.Add("response_type", responseTypeCode.String())
	params.Add("redirect_uri", req.redirectURI)

	return g.AuthURL + "?" + params.Encode()
}

// GetToken exchanges an authorization code for an access token.
func (g GitHubOAuth) GetToken(ctx context.Context, req *TokenRequest) (*http.Response, error) {
	params := &url.Values{}

	params.Add("client_id", req.clientID)
	params.Add("client_secret", req.clientSecret)
	params.Add("code", req.code)
	params.Add("redirect_uri", req.redirectURI)

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodGet, g.TokenURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.New("github: failed to create token request")
	}

	return http.DefaultClient.Do(tokenReq)
}

// GetUserInfo fetches user information from GitHub using the access token.
func (g GitHubOAuth) GetUserInfo(ctx context.Context, response TokenResponse) (*UserInfoResponse, error) {
	v, err := url.ParseQuery(string(response))
	if err != nil {
		return nil, fmt.Errorf("github: failed to parse token response: %w", err)
	}

	accessToken := v.Get("access_token")
	if accessToken == "" {
		return nil, errors.New("github: no access token found")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("github: failed to create UserInfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: failed to execute UserInfo request: %w", err)
	}
	defer resp.Body.Close()

	uBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github: failed to read UserInfo response: %w", err)
	}

	var uData userData

	if err := json.Unmarshal(uBody, &uData); err != nil {
		return nil, fmt.Errorf("github: failed to parse UserInfo response: %w", err)
	}

	if len(uData.Login) < 1 {
		return nil, errors.New("github: failed to get user login from UserInfo")
	}

	return &UserInfoResponse{
		UID:      "gh-" + uData.Login,
		Email:    uData.Email,
		Username: uData.Login,
	}, nil
}

// GetScope returns the OAuth scopes required for GitHub authentication.
func (g GitHubOAuth) GetScope() (ScopeName, ScopeValue) {
	str := &strings.Builder{}

	str.WriteString(scopeReadUser.String())
	str.WriteRune(' ')
	str.WriteString(scopeUserEmail.String())

	return scopeName, ScopeValue(str.String())
}
