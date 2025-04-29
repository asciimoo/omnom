package oauth

import (
	"context"
	"encoding/json"
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
	AuthURL     AuthURL     `json:"authorization_endpoint"`
	TokenURL    TokenURL    `json:"token_endpoint"`
	UserInfoURL UserInfoURL `json:"userinfo_endpoint"`

	Scopes       []ScopeValue   `json:"scopes_supported"`
	ResponseType []ResponseType `json:"response_types_supported"`
	GrantType    []GrantType    `json:"grant_types_supported"`

	ConfigurationURL ConfigurationURL
}

func (o *OIDCOAuth) fetch(ctx context.Context) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.ConfigurationURL.String(), nil)
	if err != nil {
		return fmt.Errorf("oidc: failed to make request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("oidc: failed to fetch configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oidc: unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("oidc: failed to read OIDC configuration response: %w", err)
	}

	err = json.Unmarshal(body, &o)
	if err != nil {
		return fmt.Errorf("oidc: failed to parse OIDC configuration: %w", err)
	}

	return nil
}

func (o *OIDCOAuth) Prepare(ctx context.Context, cURL ConfigurationURL) error {
	if cURL == "" {
		return fmt.Errorf("oidc: missing configuration URL")
	}

	o.ConfigurationURL = cURL

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

func (o *OIDCOAuth) GetRedirectURL(clientID ClientID, redirectURI RedirectURI) string {
	params := &url.Values{}
	params.Add("client_id", clientID.String())
	params.Add("response_type", responseTypeCode.String())
	params.Add("redirect_uri", redirectURI.String())

	return o.AuthURL.String() + "?" + params.Encode()
}

func (o *OIDCOAuth) GetTokenRequest(
	ctx context.Context,
	clientID ClientID,
	clientSecret ClientSecret,
	code Code,
	redirectURI RedirectURI,
) (*http.Request, error) {
	data := &url.Values{}
	data.Set("grant_type", grantTypeAuthorizationCode.String())
	data.Set("code", code.String())
	data.Set("redirect_uri", redirectURI.String())
	data.Set("client_id", clientID.String())
	data.Set("client_secret", clientSecret.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.TokenURL.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (o *OIDCOAuth) GetUniqueUserID(ctx context.Context, token []byte) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.UserInfoURL.String(), nil)
	if err != nil {
		return "", err
	}

	var bearerToken tokenData

	err = json.Unmarshal(token, &bearerToken)
	if err != nil {
		return "", err
	}

	if len(bearerToken.AccessToken) < 1 {
		return "", fmt.Errorf("oidc: failed to get access token")
	}

	req.Header.Set("Authorization", "Bearer "+bearerToken.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var uData userData

	err = json.Unmarshal(body, &uData)
	if err != nil {
		return "", err
	}

	if len(uData.Email) < 1 {
		return "", fmt.Errorf("failed to extract user ID from UserInfo response")
	}

	return fmt.Sprintf("oidc-%s", uData.Email), nil
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
