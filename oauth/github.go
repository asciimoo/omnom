package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type GitHubOAuth struct {
	AuthURL  string
	TokenURL string
}

const scopeName = "scope"
const scopeVal = "read:user"

func (g GitHubOAuth) GetRedirectURL(clientID, handlerURL string) string {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("scope", "read:user")
	params.Add("response_type", "code")
	params.Add("redirect_uri", handlerURL)
	return fmt.Sprintf("%s?%s", g.AuthURL, params.Encode())
}

func (g GitHubOAuth) GetScope() (string, string) {
	return scopeName, scopeVal
}

func (g GitHubOAuth) GetTokenRequest(clientID, clientSecret, code, handlerURL string) (*http.Request, error) {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", handlerURL)
	u := fmt.Sprintf("%s?%s", g.TokenURL, params.Encode())
	return http.NewRequest("GET", u, nil)
}

func (g GitHubOAuth) GetUniqueUserID(body []byte) (string, error) {
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "", err
	}

	t := v.Get("access_token")
	if t == "" {
		return "", errors.New("no access token found")
	}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
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

	uBody, err := ioutil.ReadAll(resp.Body)
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
