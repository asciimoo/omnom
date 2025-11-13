package activitypub

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

const (
	apRequestTimeout = 10 * time.Second
	jsonNull         = "null"
)

type Outbox struct {
	Context      string        `json:"@context"`
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Summary      string        `json:"summary"`
	TotalItems   int64         `json:"totalItems"`
	OrderedItems []*OutboxItem `json:"orderedItems"`
}

type OutboxItem struct {
	Context   string       `json:"@context"`
	ID        string       `json:"id"`
	Type      string       `json:"type"`
	Actor     string       `json:"actor"`
	To        []string     `json:"to"`
	Cc        []string     `json:"cc"`
	Published string       `json:"published"`
	Object    OutboxObject `json:"object"`
	//Signature *Signature   `json:"signature,omitempty"`
}

type OutboxObject struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Content      string            `json:"content"`
	URL          string            `json:"url"`
	URI          string            `json:"uri"`
	Summary      string            `json:"summary"`
	AttributedTo string            `json:"attributedTo"`
	To           []string          `json:"to"`
	Cc           []string          `json:"cc"`
	Published    string            `json:"published"`
	Tag          []Tag             `json:"tag"`
	Replies      map[string]string `json:"replies"`
}

type Tag struct {
	Type string `json:"type"`
	Href string `json:"href"`
	Name string `json:"name"`
}

type Identity struct {
	Context           *Context `json:"@context"`
	ID                string   `json:"id"`
	Type              string   `json:"type"`
	Following         *string  `json:"following,omitempty"`
	Followers         *string  `json:"followers,omitempty"`
	Inbox             string   `json:"inbox"`
	Outbox            string   `json:"outbox"`
	PreferredUsername string   `json:"preferredUsername"`
	Name              string   `json:"name"`
	Summary           string   `json:"summary"`
	URL               string   `json:"url"`
	Discoverable      bool     `json:"discoverable"`
	Memorial          bool     `json:"memorial"`
	Icon              *Image   `json:"icon"`
	Image             *Image   `json:"image"`
	PubKey            PubKey   `json:"publicKey"`
}

type Image struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
}

type PubKey struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

type Webfinger struct {
	Subject string   `json:"subject"`
	Aliases []string `json:"aliases"`
	Links   []Link   `json:"links"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Type string `json:"type"`
}

type InboxRequest struct {
	Context      *Context            `json:"@context,omitempty"`
	ID           string              `json:"id,omitempty"`
	Type         string              `json:"type,omitempty"`
	Actor        string              `json:"actor,omitempty"`
	AttributedTo string              `json:"attributedTo,omitempty"`
	Object       *InboxRequestObject `json:"object,omitempty"`
	To           []string            `json:"to,omitempty"`
	Cc           []string            `json:"cc,omitempty"`
	Published    string              `json:"published,omitempty"`
	Tag          []Tag               `json:"tag,omitempty"`
	Replies      map[string]string   `json:"replies,omitempty"`
}

type InboxRequestObject struct {
	ID           string `json:"id,omitempty"`
	Type         string `json:"type,omitempty"`
	Actor        string `json:"actor,omitempty"`
	URL          string `json:"url,omitempty"`
	Object       string `json:"object,omitempty"`
	Content      string `json:"content,omitempty"`
	AttributedTo string `json:"attributedTo,omitempty"`
	inlineID     bool
}

type Context struct {
	ID    string
	Parts []any
}

type FollowResponseItem struct {
	Context string               `json:"@context"`
	ID      string               `json:"id"`
	Type    string               `json:"type"`
	Actor   string               `json:"actor"`
	Object  FollowResponseObject `json:"object"`
}

type FollowResponseObject struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Actor  string `json:"actor"`
	Object string `json:"object"`
}

// https://docs.joinmastodon.org/spec/security/#ld
//type Signature struct {
//	Type    string `json:"type"`
//	Creator string `json:"creator"`
//	Created string `json:"created"`
//	Sig     string `json:"signatureValue"`
//}

func SendSignedPostRequest(us, keyID string, data []byte, key *rsa.PrivateKey) error {
	u, err := url.Parse(us)
	if err != nil {
		return err
	}
	d := time.Now().UTC().Format(http.TimeFormat)
	hash := sha256.Sum256(data)
	digest := fmt.Sprintf("SHA-256=%s", base64.StdEncoding.EncodeToString(hash[:]))
	sigData := fmt.Appendf(nil, "(request-target): post %s\nhost: %s\ndate: %s\ndigest: %s", u.Path, u.Host, d, digest)
	sigHash := sha256.Sum256(sigData)
	sig, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, sigHash[:])
	if err != nil {
		return err
	}
	sigHeader := fmt.Sprintf(`keyId="%s",headers="(request-target) host date digest",signature="%s",algorithm="rsa-sha256"`, keyID, base64.StdEncoding.EncodeToString(sig))
	cli := &http.Client{Timeout: apRequestTimeout}
	req, err := http.NewRequest("POST", us, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Host", u.Host)
	req.Header.Set("Date", d)
	req.Header.Set("Digest", digest)
	req.Header.Set("Signature", sigHeader)
	req.Header.Set("Content-Type", "application/activity+json; charset=utf-8")
	req.Header.Set("Accept", "application/activity+json")
	r, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	rb, _ := io.ReadAll(r.Body)
	if bytes.Contains(rb, []byte("error")) {
		return errors.New("invalid response: " + string(rb))
	}
	return nil
}

func SendFollowRequest(inURL, actorURL, userURL string, key *rsa.PrivateKey) error {
	id := userURL + "#" + uuid.NewString()
	r := InboxRequest{
		Context: &Context{ID: "https://www.w3.org/ns/activitystreams"},
		ID:      id,
		Type:    "Follow",
		Actor:   userURL,
		Object: &InboxRequestObject{
			ID:       actorURL,
			inlineID: true,
		},
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return SendSignedPostRequest(inURL, userURL+"#key", data, key)
}

func SendUnfollowRequest(us, userURL string, key *rsa.PrivateKey) error {
	id := userURL + "#" + uuid.NewString()
	r := InboxRequest{
		Context: &Context{ID: "https://www.w3.org/ns/activitystreams"},
		ID:      id,
		Type:    "Undo",
		Actor:   userURL,
		Object: &InboxRequestObject{
			ID:     us + "#" + uuid.NewString(),
			Type:   "Follow",
			Actor:  userURL,
			Object: us,
		},
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	return SendSignedPostRequest(us, userURL+"#key", data, key)
}

func FetchObject(u string) (*InboxRequestObject, error) {
	c := &http.Client{Timeout: apRequestTimeout}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/activity+json; charset=utf-8")
	req.Header.Set("Accept", "application/activity+json")
	r, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	o := &InboxRequestObject{}
	err = json.NewDecoder(r.Body).Decode(o)
	return o, err
}

func FetchActor(us string, keyID string, key *rsa.PrivateKey) (*Identity, error) {
	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}
	c := &http.Client{Timeout: apRequestTimeout}
	req, err := http.NewRequest("GET", us, nil)
	if err != nil {
		return nil, err
	}
	d := time.Now().UTC().Format(http.TimeFormat)
	sigData := fmt.Appendf(nil, "(request-target): get %s\nhost: %s\ndate: %s", u.Path, u.Host, d)
	sigHash := sha256.Sum256(sigData)
	sig, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, sigHash[:])
	if err != nil {
		return nil, errors.New("failed to sign data")
	}
	sigHeader := fmt.Sprintf(`keyId="%s",headers="(request-target) host date",signature="%s",algorithm="rsa-sha256"`, keyID, base64.StdEncoding.EncodeToString(sig))
	req.Header.Set("Host", u.Host)
	req.Header.Set("Date", d)
	req.Header.Set("Signature", sigHeader)
	req.Header.Set("Content-Type", "application/activity+json; charset=utf-8")
	req.Header.Set("Accept", "application/activity+json")
	r, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	i := &Identity{}
	err = json.NewDecoder(r.Body).Decode(i)
	if i.ID == "" || i.Inbox == "" || i.Outbox == "" {
		return nil, errors.New("mandatory actor data is missing")
	}
	return i, err
}

func (c *Context) UnmarshalJSON(data []byte) error {
	if data[0] == '"' && data[len(data)-1] == '"' {
		return json.Unmarshal(data, &c.ID)
	}
	if data[0] == '[' && data[len(data)-1] == ']' {
		d := []any{}
		err := json.Unmarshal(data, &d)
		c.Parts = d
		return err
	}
	return nil
}

func (c *Context) MarshalJSON() ([]byte, error) {
	if c.ID != "" {
		return json.Marshal(c.ID)
	}
	return json.Marshal(c.Parts)
}

func (i *Image) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == jsonNull {
		return nil
	}
	if data[0] == '"' && data[len(data)-1] == '"' {
		return json.Unmarshal(data, &i.URL)
	}
	if data[0] == '{' && data[len(data)-1] == '}' {
		type T struct {
			Type      string `json:"type"`
			MediaType string `json:"mediaType"`
			URL       string `json:"url"`
		}
		return json.Unmarshal(data, (*T)(i))
	}
	return nil
}

// required for json conversion
type shadowInboxRequestObject InboxRequestObject

func (i *InboxRequestObject) MarshalJSON() ([]byte, error) {
	if i.inlineID {
		return json.Marshal(i.ID)
	}
	return json.Marshal((*shadowInboxRequestObject)(i))
}

func (i *InboxRequestObject) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == jsonNull {
		return nil
	}
	if data[0] == '"' && data[len(data)-1] == '"' {
		return json.Unmarshal(data, &i.ID)
	}
	if data[0] == '{' && data[len(data)-1] == '}' {
		return json.Unmarshal(data, (*shadowInboxRequestObject)(i))
	}
	return nil
}
