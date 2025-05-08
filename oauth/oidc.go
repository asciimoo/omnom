package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const (
	scopeOpenID  ScopeValue = "openid"
	scopeProfile ScopeValue = "profile"
	scopeEmail   ScopeValue = "email"
)

type OIDCOAuth struct {
	AuthURL     string `json:"authorization_endpoint"`
	TokenURL    string `json:"token_endpoint"`
	UserInfoURL string `json:"userinfo_endpoint"`

	Scopes       []ScopeValue   `json:"scopes_supported"`
	ResponseType []ResponseType `json:"response_types_supported"`
	GrantType    []GrantType    `json:"grant_types_supported"`

	ConfigurationURL string
}

func (o *OIDCOAuth) fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.ConfigurationURL, nil)
	if err != nil {
		return fmt.Errorf("oidc: failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("oidc: failed to fetch configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oidc: unexpected configuration response status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("oidc: failed to read configuration response: %w", err)
	}

	err = json.Unmarshal(body, &o)
	if err != nil {
		return fmt.Errorf("oidc: failed to parse configuration: %w", err)
	}

	return nil
}

func (o *OIDCOAuth) Prepare(ctx context.Context, req *PrepareRequest) error {
	if req.configurationURL == "" {
		return fmt.Errorf("oidc: missing configuration URL")
	}

	o.ConfigurationURL = req.configurationURL

	if err := o.fetch(ctx); err != nil {
		return err
	}

	if o.AuthURL == "" || o.TokenURL == "" {
		return fmt.Errorf("oidc: empty AuthURL or TokenURL")
	}

	if len(o.Scopes) < 1 || slices.Index(o.Scopes, scopeOpenID) == -1 {
		return fmt.Errorf("oidc: must have \"%s\" in scopes_supported", scopeOpenID)
	}

	if len(o.ResponseType) < 1 || slices.Index(o.ResponseType, responseTypeCode) == -1 {
		return fmt.Errorf("oidc: must have \"%s\" in response_types_supported", responseTypeCode)
	}

	if len(o.GrantType) < 1 || slices.Index(o.GrantType, grantTypeAuthorizationCode) == -1 {
		return fmt.Errorf("oidc: must have \"%s\" in grant_types_supported", grantTypeAuthorizationCode)
	}

	return nil
}

func (o *OIDCOAuth) GetRedirectURL(req *RedirectURIRequest) string {
	params := &url.Values{}

	params.Add("client_id", req.clientID)
	params.Add("response_type", responseTypeCode.String())
	params.Add("redirect_uri", req.redirectURI)

	return o.AuthURL + "?" + params.Encode()
}

func (o *OIDCOAuth) GetToken(ctx context.Context, req *TokenRequest) (*http.Response, error) {
	params := &url.Values{}

	params.Set("grant_type", grantTypeAuthorizationCode.String())
	params.Set("code", req.code)
	params.Set("redirect_uri", req.redirectURI)
	params.Set("client_id", req.clientID)
	params.Set("client_secret", req.clientSecret)

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.TokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, errors.New("oidc: failed to create token request")
	}

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return http.DefaultClient.Do(tokenReq)
}

func (o *OIDCOAuth) GetUserInfo(ctx context.Context, response TokenResponse) (*UserInfoResponse, error) {
	var bearer tokenData

	if err := json.Unmarshal(response, &bearer); err != nil {
		return nil, err
	}

	if len(bearer.AccessToken) < 1 {
		return nil, fmt.Errorf("oidc: failed to get access token")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to create UserInfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+bearer.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to fetch UserInfo response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc: unexpected UserInfo response status code: %d", resp.StatusCode)
	}

	uBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to read UserInfo response: %w", err)
	}

	var uData userData

	if err := json.Unmarshal(uBody, &uData); err != nil {
		return nil, fmt.Errorf("oidc: failed to parse UserInfo response: %w", err)
	}

	if len(uData.Email) < 1 {
		return nil, fmt.Errorf("oidc: failed to parse email from UserInfo response")
	}

	return &UserInfoResponse{
		UID:      "oidc-" + uData.Email,
		Email:    uData.Email,
		Username: uData.PreferredUsername,
	}, nil
}

func (o *OIDCOAuth) GetScope() (ScopeName, ScopeValue) {
	str := &strings.Builder{}

	str.WriteString(scopeOpenID.String())
	str.WriteRune(' ')
	str.WriteString(scopeEmail.String())
	str.WriteRune(' ')
	str.WriteString(scopeProfile.String())

	return scopeName, ScopeValue(str.String())
}
