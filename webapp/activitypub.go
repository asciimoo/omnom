package webapp

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const apRequestTimeout = 10 * time.Second

var apSigHeaderRe = regexp.MustCompile(`keyId="([^"]+)",headers="([^"]+)",signature="([^"]+)"`)

type apOutbox struct {
	Context      string         `json:"@context"`
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Summary      string         `json:"summary"`
	TotalItems   int64          `json:"totalItems"`
	OrderedItems []apOutboxItem `json:"orderedItems"`
}

type apOutboxItem struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Actor     string         `json:"actor"`
	To        []string       `json:"to"`
	Cc        []string       `json:"cc"`
	Published string         `json:"published"`
	Object    apOutboxObject `json:"object"`
}

type apOutboxObject struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Content      string            `json:"content"`
	URL          string            `json:"url"`
	Summary      string            `json:"summary"`
	AttributedTo string            `json:"attributedTo"`
	To           []string          `json:"to"`
	Cc           []string          `json:"cc"`
	Published    string            `json:"published"`
	Tag          []apTag           `json:"tag"`
	Replies      map[string]string `json:"replies"`
}

type apTag struct {
	Type string `json:"type"`
	Href string `json:"href"`
	Name string `json:"name"`
}

type apIdentity struct {
	Context           *apContext `json:"@context"`
	ID                string     `json:"id"`
	Type              string     `json:"type"`
	Following         string     `json:"following"`
	Followers         string     `json:"followers"`
	Inbox             string     `json:"inbox"`
	Outbox            string     `json:"outbox"`
	PreferredUsername string     `json:"preferredUsername"`
	Name              string     `json:"name"`
	Summary           string     `json:"summary"`
	URL               string     `json:"url"`
	Discoverable      bool       `json:"discoverable"`
	Memorial          bool       `json:"memorial"`
	Icon              apImage    `json:"icon"`
	Image             apImage    `json:"image"`
	PubKey            apPubKey   `json:"publicKey"`
}

type apImage struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
}

type apPubKey struct {
	Context      string `json:"@context"`
	ID           string `json:"id"`
	Type         string `json:"type"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

type apWebfinger struct {
	Subject string   `json:"subject"`
	Aliases []string `json:"aliases"`
	Links   []apLink `json:"links"`
}

type apLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Type string `json:"type"`
}

type apInboxRequest struct {
	ID     string               `json:"id"`
	Type   string               `json:"type"`
	Actor  string               `json:"actor"`
	Object apInboxRequestObject `json:"object"`
}

type apInboxRequestObject struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Actor  string `json:"actor"`
	Object string `json:"object"`
}

type apContext struct {
	ID    string
	Parts []interface{}
}

type apFollowResponseItem struct {
	Context string                 `json:"@context"`
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Actor   string                 `json:"actor"`
	Object  apFollowResponseObject `json:"object"`
}

type apFollowResponseObject struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Actor  string `json:"actor"`
	Object string `json:"object"`
}

const contentTpl = `<h1><a href="%[1]s">%[2]s</a></h1>
<p>%[3]s</p>
<small>Bookmarked by <a href="https://github.com/asciimoo/omnom">Omnom</a> - <a href="%[4]s">view bookmark</a></small>`

func parseURL(us string) (*url.URL, error) {
	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Del("format")
	q.Del("pageno")
	u.RawQuery = q.Encode()
	return u, nil
}

func apOutboxResponse(c *gin.Context, bs []*model.Bookmark, bc int64) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	baseU := getFullURLPrefix(c)
	u, err := parseURL(baseU + c.Request.URL.String())
	if err != nil {
		log.Println("ActivityPub URL parse error", err)
		return
	}

	resp := apOutbox{
		Context:      "https://www.w3.org/ns/activitystreams",
		ID:           u.String(),
		Type:         "OrderedCollection",
		Summary:      "Recent bookmarks of " + u.String(),
		TotalItems:   bc,
		OrderedItems: make([]apOutboxItem, len(bs)),
	}
	for i, b := range bs {
		id := getFullURL(c, fmt.Sprintf("%s?id=%d", URLFor("Bookmark"), b.ID))
		actor := fmt.Sprintf("%s/bookmarks?owner=%s", baseU, b.User.Username)
		published := b.CreatedAt.Format(time.RFC3339)
		item := apOutboxItem{
			ID:    id + "#activity",
			Type:  "Create",
			Actor: actor,
			To: []string{
				"https://www.w3.org/ns/activitystreams#Public",
			},
			Cc:        []string{},
			Published: published,
			Object: apOutboxObject{
				ID:           id,
				Type:         "Note",
				Summary:      fmt.Sprintf("Bookmark of \"%s\"", b.Title),
				Content:      fmt.Sprintf(contentTpl, b.URL, b.Title, b.Notes, id),
				URL:          b.URL,
				AttributedTo: actor,
				To: []string{
					"https://www.w3.org/ns/activitystreams#Public",
				},
				Cc:        []string{},
				Published: published,
				Tag:       make([]apTag, len(b.Tags)),
				Replies:   map[string]string{},
			},
		}
		for i, t := range b.Tags {
			tagBase := getFullURL(c, fmt.Sprintf("%s?tag=", URLFor("Public bookmarks")))
			item.Object.Tag[i] = apTag{
				Type: "Hashtag",
				Href: tagBase + t.Text,
				Name: t.Text,
			}
		}
		resp.OrderedItems[i] = item
	}

	j, err := json.Marshal(resp)
	if err != nil {
		log.Println("ActivityPub JSON serialization error", err)
	}
	_, err = c.Writer.Write(j)
	if err != nil {
		log.Println("ActivityPub ident write error", err)
	}
}

func apIdentityResponse(c *gin.Context) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	id := getFullURL(c, c.Request.URL.String())
	u, err := parseURL(id)
	if err != nil {
		log.Println("ActivityPub URL parse error", err)
		return
	}
	cfg, _ := c.Get("config")
	pk, err := cfg.(*config.Config).ActivityPub.ExportPubKey()
	if err != nil {
		log.Println("ActivityPub JSON serialization error", err)
		return
	}
	inbox := getFullURL(c, URLFor("ActivityPub inbox"))
	feedID := apCreateFeedID(id)
	j, err := json.Marshal(apIdentity{
		Context: &apContext{
			ID: "https://www.w3.org/ns/activitystreams",
		},
		ID:                id,
		Type:              "Person",
		Inbox:             inbox,
		Outbox:            getFullURL(c, addURLParam(u.String(), "format=activitypub")),
		PreferredUsername: "omnom" + feedID,
		Name:              getFullURL(c, "/"+feedID),
		URL:               getFullURL(c, "/"),
		Discoverable:      true,
		Icon: apImage{
			Type:      "Image",
			MediaType: "image/png",
			URL:       getFullURL(c, "/static/icons/addon_icon.png"),
		},
		Image: apImage{
			Type:      "Image",
			MediaType: "image/png",
			URL:       getFullURL(c, "/static/icons/addon_icon.png"),
		},
		PubKey: apPubKey{
			Context:      "https://w3id.org/security/v1",
			ID:           id + "#key",
			Type:         "Key",
			Owner:        id,
			PublicKeyPem: string(pk),
		},
	})
	if err != nil {
		log.Println("ActivityPub JSON serialization error", err)
	}
	_, err = c.Writer.Write(j)
	if err != nil {
		log.Println("ActivityPub ident write error")
	}
}

func apInboxResponse(c *gin.Context) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	body, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	d := &apInboxRequest{}
	err = json.Unmarshal(body, d)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if d.Object.ID == "" || d.Actor == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Missing attributes",
		})
		return
	}
	if !strings.HasPrefix(d.Object.ID, getFullURLPrefix(c)) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}
	switch d.Type {
	case "Follow":
		go apInboxFollowResponse(c, "Follow", d, body)
	case "Undo":
		go apInboxFollowResponse(c, "Undo", d, body)
	default:
		log.Println("Unhandled ActivityPub inbox message type: " + d.Type)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Unknown action type",
		})
		return
	}
	c.JSON(http.StatusOK, map[string]string{
		"status": "OK",
	})
}

func apInboxFollowResponse(c *gin.Context, action string, d *apInboxRequest, payload []byte) {
	actor, err := apFetchActor(d.Actor)
	// validate d.Actor signature
	if err != nil {
		log.Println("Failed to fetch actor information:", d.Actor, err)
		return
	}
	if err := apCheckSignature(c, actor.PubKey.PublicKeyPem, payload); err != nil {
		log.Println("Failed to validate actor signature:", d.Actor, err)
		return
	}
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	data, err := json.Marshal(apFollowResponseItem{
		Context: "https://www.w3.org/ns/activitystreams",
		ID:      getFullURL(c, "/"+uuid.New().String()),
		Type:    "Accept",
		Object: apFollowResponseObject{
			ID:     d.ID,
			Type:   action,
			Actor:  d.Actor,
			Object: d.Object.ID,
		},
	})
	if err != nil {
		log.Println("Failed to serialize AP inbox response:", d.Actor, err)
		return
	}
	err = apSendSignedPostRequest(actor.Inbox, d.Object.ID+"#key", data, key)
	if err != nil {
		log.Println("Failed to send HTTP request:", d.Actor, err)
		return
	}
	u, err := url.Parse(d.Object.ID)
	if err != nil {
		log.Println("Invalid subscription url:", d.Actor, err)
		return
	}
	err = model.CreateAPFollower(d.Actor, u.RawQuery)
	if err != nil {
		log.Println("Failed to create AP follower", d.Actor, err)
		return
	}
}

func apParseSigHeader(c *gin.Context, digest string) (string, []byte, error) {
	sigParts := apSigHeaderRe.FindStringSubmatch(c.Request.Header.Get("Signature"))
	if len(sigParts) != 4 {
		return "", nil, errors.New("invalid Signature header format")
	}
	signature := sigParts[3]
	sigHeaders := make([]string, 0, 3)
	for _, h := range strings.Split(sigParts[2], " ") {
		if h == "(request-target)" {
			method := strings.ToLower(c.Request.Method)
			// TODO perhaps it is better to use the path information from
			// the ActivityPub request payload, because c.Request.URL.Path
			// can be different in case of some weird reverse proxying shenanigans
			s := fmt.Sprintf("%s %s", method, c.Request.URL.Path)
			if c.Request.URL.RawQuery != "" {
				s += "?" + c.Request.URL.RawQuery
			}
			sigHeaders = append(sigHeaders, s)
		} else {
			hv := c.Request.Header.Get(strings.Title(h))
			if h == "digest" {
				if hv != digest {
					return "", nil, errors.New("digest hash mismatch")
				}
			}
			sigHeaders = append(sigHeaders, fmt.Sprintf("%s: %s", h, hv))
		}
	}
	sigH := []byte(strings.Join(sigHeaders, " "))
	hash := sha256.Sum256(sigH)
	return signature, hash[:], nil
}

func apCheckSignature(c *gin.Context, key string, payload []byte) error {
	pHash := sha256.Sum256(payload)
	digest := fmt.Sprintf("SHA-256=%s", base64.StdEncoding.EncodeToString(pHash[:]))
	signature, hash, err := apParseSigHeader(c, digest)
	if err != nil {
		return fmt.Errorf("failed to parse signature header: %w", err)
	}
	pb, _ := pem.Decode([]byte(key))
	if pb == nil || pb.Type != "PUBLIC KEY" {
		return errors.New("failed to decode PEM block containing public key")
	}
	pk, err := x509.ParsePKIXPublicKey(pb.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse client cert: %w", err)
	}
	if _, ok := pk.(*rsa.PublicKey); !ok {
		return errors.New("invalid key type")
	}
	signatureDecoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}
	err = rsa.VerifyPKCS1v15(pk.(*rsa.PublicKey), crypto.SHA256, hash, signatureDecoded)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}
	return nil
}

func apSendSignedPostRequest(us, keyID string, data []byte, key *rsa.PrivateKey) error {
	u, err := url.Parse(us)
	if err != nil {
		return err
	}
	d := time.Now().UTC().Format(http.TimeFormat)
	hash := sha256.Sum256(data)
	digest := fmt.Sprintf("SHA-256=%s", base64.StdEncoding.EncodeToString(hash[:]))
	sigData := []byte(fmt.Sprintf("(request-target): post %s\nhost: %s\ndate: %s\ndigest: %s", u.Path, u.Host, d, digest))
	sigHash := sha256.Sum256(sigData)
	sig, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, sigHash[:])
	if err != nil {
		log.Println("Can't sign request data:", err)
		return err
	}
	sigHeader := fmt.Sprintf(`keyId="%s",headers="(request-target) host date digest",signature="%s",algorithm="rsa-sha256"`, keyID, base64.StdEncoding.EncodeToString(sig))
	cli := &http.Client{Timeout: apRequestTimeout}
	req, err := http.NewRequest("POST", us, bytes.NewReader(data))
	if err != nil {
		log.Println("Can't create signed POST request:", err)
		return err
	}
	req.Header.Set("Host", u.Host)
	req.Header.Set("Date", d)
	req.Header.Set("Digest", digest)
	req.Header.Set("Signature", sigHeader)
	req.Header.Set("Content-Type", "application/activity+json; charset=utf-8")
	req.Header.Set("Accept", "application/activity+json; charset=utf-8")
	r, err := cli.Do(req)
	if err != nil {
		log.Println("Failed to send follow accept response:", err)
		return err
	}
	defer r.Body.Close()
	rb, _ := io.ReadAll(r.Body)
	if bytes.Contains(rb, []byte("error")) {
		log.Println("Follow accept response contains error:", string(rb))
		return errors.New("invalid response")
	}
	return nil
}

func apFetchActor(u string) (*apIdentity, error) {
	c := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/activity+json; charset=utf-8")
	r, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	i := &apIdentity{}
	err = json.NewDecoder(r.Body).Decode(i)
	return i, err
}

func apWebfingerResponse(c *gin.Context) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	s := c.Query("resource")
	u := getFullURL(c, URLFor("Public bookmarks"))
	j, err := json.Marshal(apWebfinger{
		Subject: s,
		Aliases: []string{},
		Links: []apLink{
			apLink{Rel: "self", Type: "application/activity+json", Href: getFullURL(c, addURLParam(u, "format=ap-identity"))},
			apLink{Rel: "http://webfinger.net/rel/profile-page", Type: "text/html", Href: u},
		},
	})
	if err != nil {
		log.Println("Webfinger JSON serialization error", err)
	}
	_, err = c.Writer.Write(j)
	if err != nil {
		log.Println("ActivityPub webfinger write error")
	}
}

func apCreateFeedID(us string) string {
	u, err := url.Parse(us)
	if err != nil {
		return ""
	}
	q := u.Query()
	s := ""
	if q.Get("owner") != "" {
		s += "@user." + q.Get("owner")
	}
	if q.Get("domain") != "" {
		s += "@domain." + q.Get("domain")
	}
	if q.Get("tag") != "" {
		s += "@tag." + strings.ReplaceAll(q.Get("tag"), " ", "_")
	}
	if q.Get("query") != "" {
		s += "@query." + strings.ReplaceAll(q.Get("query"), " ", "_")
	}
	return s
}

func (i *apInboxRequestObject) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' && data[len(data)-1] == '"' {
		return json.Unmarshal(data, &i.ID)
	}
	if data[0] == '{' && data[len(data)-1] == '}' {
		type T struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Actor  string `json:"actor"`
			Object string `json:"object"`
		}
		return json.Unmarshal(data, (*T)(i))
	}
	return nil
}

func (c *apContext) UnmarshalJSON(data []byte) error {
	if data[0] == '"' && data[len(data)-1] == '"' {
		return json.Unmarshal(data, &c.ID)
	}
	if data[0] == '[' && data[len(data)-1] == ']' {
		d := []interface{}{}
		err := json.Unmarshal(data, &d)
		c.Parts = d
		return err
	}
	return nil
}

func (c *apContext) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, strings.ReplaceAll(c.ID, `"`, `\"`))), nil
}
