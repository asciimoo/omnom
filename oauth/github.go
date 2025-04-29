package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type GitHubOAuth struct {
	AuthURL  AuthURL
	TokenURL TokenURL
}

const scopeReadUser ScopeValue = "read:user"

func (g GitHubOAuth) Prepare(_ context.Context, _ ConfigurationURL) error { return nil }

func (g GitHubOAuth) GetRedirectURL(clientID ClientID, redirectURI RedirectURI) string {
	params := &url.Values{}
	params.Add("client_id", clientID.String())
	params.Add("scope", scopeReadUser.String())
	params.Add("response_type", responseTypeCode.String())
	params.Add("redirect_uri", redirectURI.String())

	return g.AuthURL.String() + "?" + params.Encode()
}

func (g GitHubOAuth) GetScope() (ScopeName, ScopeValue) {
	return scopeName, scopeReadUser
}

func (g GitHubOAuth) GetTokenRequest(ctx context.Context, clientID ClientID, clientSecret ClientSecret, code Code, redirectURI RedirectURI) (*http.Request, error) {
	params := &url.Values{}
	params.Add("client_id", clientID.String())
	params.Add("client_secret", clientSecret.String())
	params.Add("code", code.String())
	params.Add("redirect_uri", redirectURI.String())

	u := fmt.Sprintf("%s?%s", g.TokenURL, params.Encode())

	return http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
}

func (g GitHubOAuth) GetUniqueUserID(ctx context.Context, body []byte) (string, error) {
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "", err
	}

	t := v.Get("access_token")
	if t == "" {
		return "", errors.New("no access token found")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+t)

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

	log.Println("BODY:", string(uBody))
	var j map[string]interface{}

	err = json.Unmarshal(uBody, &j)
	if err != nil {
		return "", err
	}

	l, ok := j["login"].(string)
	if !ok {
		return "", errors.New("failed to get user login data")
	}

	return fmt.Sprintf("gh-%s", l), nil
}
