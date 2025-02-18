package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type GoogleOAuth struct {
	AuthURL  string
	TokenURL string
}

func (g GoogleOAuth) GetRedirectURL(clientID, handlerURL string) string {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", handlerURL)
	//params.Add("scope", "https://www.googleapis.com/auth/userinfo.email")
	return fmt.Sprintf("%s?%s", g.AuthURL, params.Encode())
}

func (g GoogleOAuth) GetScope() (string, string) {
	return "scope", "https://www.googleapis.com/auth/userinfo.email"
}

func (g GoogleOAuth) GetTokenRequest(clientID, clientSecret, code, handlerURL string) (*http.Request, error) {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("code", code)
	params.Add("grant_type", "authorization_code")
	params.Add("redirect_uri", handlerURL)
	r, err := http.NewRequest("POST", g.TokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return r, err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r, err
}

func (g GoogleOAuth) GetUniqueUserID(body []byte) (string, error) {
	var j struct {
		Tok string `json:"access_token"`
	}
	err := json.Unmarshal(body, &j)
	if err != nil {
		return "", err
	}

	if j.Tok == "" {
		return "", errors.New("no access token found")
	}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo?access_token="+j.Tok, nil)
	if err != nil {
		return "", errors.New("invalid token")
	}
	req.Header.Set("Authorization", "bearer "+j.Tok)

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
	var u struct {
		Email string `json:"email"`
	}
	err = json.Unmarshal(uBody, &u)
	if err != nil {
		return "", err
	}
	if u.Email == "" {
		return "", errors.New("invalid email")
	}
	return fmt.Sprintf("goog-%s", u.Email), nil
}
