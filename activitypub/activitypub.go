// Package activitypub implements ActivityPub federation protocol for Omnom.
//
// This package enables Omnom to act as an ActivityPub server, allowing users to:
//   - Follow ActivityPub actors (Mastodon, Pleroma, etc.)
//   - Receive posts from followed actors in their feed
//   - Be discovered and followed by other ActivityPub servers
//   - Serve user profiles via WebFinger
//
// The implementation follows the ActivityPub specification (W3C Recommendation)
// and supports core activities:
//   - Follow/Unfollow: Subscribe to actor updates
//   - Accept: Confirm follow requests
//   - Create: Receive new posts
//   - Announce: Receive boosts/reblogs
//
// HTTP signatures are used to authenticate requests between servers. The package
// handles signing outgoing requests and verifying incoming requests using RSA keys
// configured in the application.
//
// Key types:
//   - Actor: Represents a user or service on the federation
//   - InboxRequest: Incoming activity from another server
//   - OutboxItem: Outgoing activity to send to followers
//
// Example usage:
//
//	// Follow an actor
//	actor, err := activitypub.FetchActor(actorURL, userKey, privateKey)
//	err = activitypub.SendFollowRequest(actor.Inbox, actorURL, userURL, privateKey)
//
//	// Process inbox delivery
//	err = activitypub.ProcessInboxRequest(request, config, user)
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

	"github.com/asciimoo/omnom/storage"
	"github.com/asciimoo/omnom/utils"

	"github.com/google/uuid"
)

const (
	apRequestTimeout = 10 * time.Second
	jsonNull         = "null"
	imageType        = "Image"
)

// Outbox represents an ActivityPub outbox containing published activities.
type Outbox struct {
	Context      string        `json:"@context"`
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Summary      string        `json:"summary"`
	TotalItems   int64         `json:"totalItems"`
	OrderedItems []*OutboxItem `json:"orderedItems"`
}

// OutboxItem represents a single activity in an outbox.
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

// OutboxObject represents the object being published in an activity.
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

// Tag represents a tag or mention in ActivityPub content.
type Tag struct {
	Type string `json:"type"`
	Href string `json:"href"`
	Name string `json:"name"`
}

// Identity represents an ActivityPub actor (user or service).
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

// Image represents an image attachment in ActivityPub.
type Image struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
}

// PubKey represents a public key for HTTP signature verification.
type PubKey struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

// Webfinger represents a WebFinger response for actor discovery.
type Webfinger struct {
	Subject string   `json:"subject"`
	Aliases []string `json:"aliases"`
	Links   []Link   `json:"links"`
}

// Link represents a link in a WebFinger response.
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Type string `json:"type"`
}

// InboxRequest represents an incoming ActivityPub activity delivered to the inbox.
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

// InboxRequestObject represents the object within an inbox activity.
type InboxRequestObject struct {
	ID           string        `json:"id,omitempty"`
	Type         string        `json:"type,omitempty"`
	Actor        string        `json:"actor,omitempty"`
	URL          string        `json:"url,omitempty"`
	Object       string        `json:"object,omitempty"`
	Content      string        `json:"content,omitempty"`
	InReplyTo    string        `json:"inReplyTo,omitempty"`
	Context      string        `json:"context,omitempty"`
	AttributedTo string        `json:"attributedTo,omitempty"`
	Attachments  []*Attachment `json:"attachment,omitempty"`
	inlineID     bool
}

// Context represents the JSON-LD context for ActivityPub messages.
type Context struct {
	ID    string
	Parts []any
}

// FollowResponseItem represents a response to a follow request (Accept activity).
type FollowResponseItem struct {
	Context string               `json:"@context"`
	ID      string               `json:"id"`
	Type    string               `json:"type"`
	Actor   string               `json:"actor"`
	Object  FollowResponseObject `json:"object"`
}

// FollowResponseObject represents the object in a follow response.
type FollowResponseObject struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Actor  string `json:"actor"`
	Object string `json:"object"`
}

// Attachment represents a media attachment in an ActivityPub post.
type Attachment struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
	Name      string `json:"name"`
}

// https://docs.joinmastodon.org/spec/security/#ld
//type Signature struct {
//	Type    string `json:"type"`
//	Creator string `json:"creator"`
//	Created string `json:"created"`
//	Sig     string `json:"signatureValue"`
//}

// SendSignedPostRequest sends an HTTP POST request with HTTP signature authentication.
// The request is signed using the provided RSA private key.
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

// SendFollowRequest sends a Follow activity to an actor's inbox.
// This is used to subscribe to an actor's posts.
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

// SendUnfollowRequest sends an Undo Follow activity to unsubscribe from an actor.
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

// FetchObject fetches an ActivityPub object from a URL.
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

// FetchActor fetches an actor's profile information with HTTP signature authentication.
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
	if err != nil {
		return nil, err
	}
	if i.ID == "" || i.Inbox == "" || i.Outbox == "" {
		return nil, errors.New("mandatory actor data is missing")
	}
	return i, nil
}

// SaveFavicon downloads and saves user favicon as a storage resource.
func (i *Identity) SaveFavicon() (string, error) {
	var uri string
	if i.Icon != nil && i.Icon.Type == imageType {
		uri = i.Icon.URL
	} else if i.Image != nil {
		uri = i.Image.URL
	} else {
		return "", nil
	}
	c := &http.Client{Timeout: apRequestTimeout}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}
	r, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	return storage.SaveResource(utils.GetExtension(uri), r.Body)
}

// GetName returns the best available display name for the actor.
// Priority: PreferredUsername > Name > ID.
func (i *Identity) GetName() string {
	if i.PreferredUsername != "" {
		return i.PreferredUsername
	}
	if i.Name != "" {
		return i.Name
	}
	return i.ID
}

// UnmarshalJSON implements custom JSON unmarshaling for Context.
// Handles both string and array representations of JSON-LD contexts.
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

// MarshalJSON implements custom JSON marshaling for Context.
// Returns either a string or array depending on the context content.
func (c *Context) MarshalJSON() ([]byte, error) {
	if c.ID != "" {
		return json.Marshal(c.ID)
	}
	return json.Marshal(c.Parts)
}

// UnmarshalJSON implements custom JSON unmarshaling for Image.
// Handles both string URLs and object representations.
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

// MarshalJSON implements custom JSON marshaling for InboxRequestObject.
// Returns just the ID string when inlineID is true, otherwise returns the full object.
func (i *InboxRequestObject) MarshalJSON() ([]byte, error) {
	if i.inlineID {
		return json.Marshal(i.ID)
	}
	return json.Marshal((*shadowInboxRequestObject)(i))
}

// UnmarshalJSON implements custom JSON unmarshaling for InboxRequestObject.
// Handles both string IDs and full object representations.
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
