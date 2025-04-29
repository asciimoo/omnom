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

type GoogleOAuth struct {
	AuthURL  AuthURL
	TokenURL TokenURL
}

func (g GoogleOAuth) Prepare(_ context.Context, _ ConfigurationURL) error { return nil }

func (g GoogleOAuth) GetRedirectURL(clientID ClientID, redirectURI RedirectURI) string {
	params := &url.Values{}
	params.Add("client_id", clientID.String())
	params.Add("response_type", responseTypeCode.String())
	params.Add("redirect_uri", redirectURI.String())

	return g.AuthURL.String() + "?" + params.Encode()
}

func (g GoogleOAuth) GetScope() (ScopeName, ScopeValue) {
	return scopeName, "https://www.googleapis.com/auth/userinfo.email"
}

func (g GoogleOAuth) GetTokenRequest(ctx context.Context, clientID ClientID, clientSecret ClientSecret, code Code, redirectURI RedirectURI) (*http.Request, error) {
	params := &url.Values{}
	params.Add("client_id", clientID.String())
	params.Add("client_secret", clientSecret.String())
	params.Add("code", code.String())
	params.Add("grant_type", grantTypeAuthorizationCode.String())
	params.Add("redirect_uri", redirectURI.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.TokenURL.String(), strings.NewReader(params.Encode()))
	if err != nil {
		return req, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, err
}

func (g GoogleOAuth) GetUniqueUserID(ctx context.Context, body []byte) (string, error) {
	var tData tokenData

	if err := json.Unmarshal(body, &tData); err != nil {
		return "", err
	}

	if len(tData.AccessToken) < 1 {
		return "", errors.New("no access token found")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo?access_token="+tData.AccessToken, nil)
	if err != nil {
		return "", errors.New("invalid token")
	}

	req.Header.Set("Authorization", "Bearer "+tData.AccessToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	uBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var uData userData
	err = json.Unmarshal(uBody, &uData)
	if err != nil {
		return "", err
	}

	if len(uData.Email) < 1 {
		return "", errors.New("invalid email")
	}

	return fmt.Sprintf("goog-%s", uData.Email), nil
}
