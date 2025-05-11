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
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	apRequestTimeout = 10 * time.Second
	followAction     = "Follow"
	unfollowAction   = "Undo"
)

const contentTpl = `<h1><a href="%[1]s">%[2]s</a></h1>
%[3]s
Bookmarked by <a href="https://github.com/asciimoo/omnom">Omnom</a> - <a href="%[4]s">view bookmark</a>`

var apSigHeaderRe = regexp.MustCompile(`keyId="([^"]+)".*,headers="([^"]+)",.*signature="([^"]+)"`)

type apOutbox struct {
	Context      string          `json:"@context"`
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Summary      string          `json:"summary"`
	TotalItems   int64           `json:"totalItems"`
	OrderedItems []*apOutboxItem `json:"orderedItems"`
}

type apOutboxItem struct {
	Context   string         `json:"@context"`
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Actor     string         `json:"actor"`
	To        []string       `json:"to"`
	Cc        []string       `json:"cc"`
	Published string         `json:"published"`
	Object    apOutboxObject `json:"object"`
	//Signature *apSignature   `json:"signature,omitempty"`
}

type apOutboxObject struct {
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
	Following         *string    `json:"following,omitempty"`
	Followers         *string    `json:"followers,omitempty"`
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
	ID           string `json:"id"`
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

// https://docs.joinmastodon.org/spec/security/#ld
type apSignature struct {
	Type    string `json:"type"`
	Creator string `json:"creator"`
	Created string `json:"created"`
	Sig     string `json:"signatureValue"`
}

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
	u, err := parseURL(getFullURL(c, c.Request.URL.String()))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse URL")
		return
	}

	resp := apOutbox{
		Context:      "https://www.w3.org/ns/activitystreams",
		ID:           u.String(),
		Type:         "OrderedCollection",
		Summary:      "Recent bookmarks of " + u.String(),
		TotalItems:   bc,
		OrderedItems: make([]*apOutboxItem, len(bs)),
	}
	for i, b := range bs {
		item := apCreateBookmarkItem(c, b, u.String())
		resp.OrderedItems[i] = item
	}

	j, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize JSON")
	}
	_, err = c.Writer.Write(j)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write response")
	}
}

func apCreateBookmarkItem(c *gin.Context, b *model.Bookmark, actor string) *apOutboxItem {
	id := getFullURL(c, fmt.Sprintf("%s?id=%d", URLFor("Bookmark"), b.ID))
	published := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	title := truncate(b.Title, 300)
	body := ""
	if b.Notes != "" {
		body = fmt.Sprintf("<p>%s</p>", truncate(b.Notes, 350-len(title)))
	}
	item := &apOutboxItem{
		Context: "https://www.w3.org/ns/activitystreams",
		ID:      id + "#activity",
		Type:    "Create",
		Actor:   actor,
		To: []string{
			"https://www.w3.org/ns/activitystreams#Public",
		},
		Cc:        []string{},
		Published: published,
		Object: apOutboxObject{
			ID:           id,
			Type:         "Note",
			Summary:      fmt.Sprintf("Bookmark of \"%s\"", b.Title),
			Content:      fmt.Sprintf(contentTpl, b.URL, title, body, id),
			URL:          b.URL,
			URI:          b.URL,
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
	return item
}

func apIdentityResponse(c *gin.Context) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	id := getFullURL(c, c.Request.URL.String())
	u, err := parseURL(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse URL")
		return
	}
	cfg, _ := c.Get("config")
	pk, err := cfg.(*config.Config).ActivityPub.ExportPubKey()
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize JSON")
		return
	}
	inbox := getFullURL(c, URLFor("ActivityPub inbox"))
	feedID := apCreateFeedID(id)
	uname := "omnom" + feedID
	j, err := json.Marshal(apIdentity{
		Context: &apContext{
			Parts: []interface{}{
				"https://www.w3.org/ns/activitystreams",
				"https://w3id.org/security/v1",
			},
		},
		ID:                id,
		Type:              "Person",
		Inbox:             inbox,
		Outbox:            getFullURL(c, addURLParam(u.String(), "format=activitypub")),
		PreferredUsername: uname,
		Name:              uname,
		URL:               id,
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
			ID:           id + "#key",
			Owner:        id,
			PublicKeyPem: string(pk),
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize JSON")
	}
	_, err = c.Writer.Write(j)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write response")
	}
}

func apInboxResponse(c *gin.Context) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	body, err := c.GetRawData()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch Inbox request body")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	d := &apInboxRequest{}
	err = json.Unmarshal(body, d)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON")
		log.Debug().Msg(string(body))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if d.Object.ID == "" || d.Actor == "" {
		log.Error().Bytes("body", body).Msg("Inbox request has missing objectID or actor")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Missing attributes",
		})
		return
	}
	switch d.Type {
	case followAction:
		go apInboxFollowResponse(c, d, body)
	case unfollowAction:
		go apInboxUnfollowResponse(c, d, body)
	default:
		log.Debug().Str("type", d.Type).Msg("Unhandled ActivityPub inbox message")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Unknown action type",
		})
		return
	}
	c.JSON(http.StatusOK, map[string]string{
		"status": "OK",
	})
}

func apInboxFollowResponse(c *gin.Context, d *apInboxRequest, payload []byte) {
	if !strings.HasPrefix(d.Object.ID, getFullURLPrefix(c)) {
		log.Error().Bytes("payload", payload).Msg("Inbox request objectID points to different host")
		return
	}
	actor, err := apFetchActor(d.Actor)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to fetch actor information")
		return
	}
	if err := apCheckSignature(c, actor.PubKey.PublicKeyPem, payload); err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to validate actor signature")
		return
	}
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	data, err := json.Marshal(apFollowResponseItem{
		Context: "https://www.w3.org/ns/activitystreams",
		ID:      getFullURL(c, "/"+uuid.New().String()),
		Type:    "Accept",
		Actor:   d.Object.ID,
		Object: apFollowResponseObject{
			ID:     d.ID,
			Type:   followAction,
			Actor:  d.Actor,
			Object: d.Object.ID,
		},
	})
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to serialize AP inbox response")
		return
	}
	err = apSendSignedPostRequest(actor.Inbox, d.Object.ID+"#key", data, key)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to send HTTP request")
		return
	}
	u, err := url.Parse(d.Object.ID)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to parse URL")
		return
	}
	err = model.CreateAPFollower(d.Actor, u.RawQuery)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to create AP follower")
		return
	}
}
func apInboxUnfollowResponse(c *gin.Context, d *apInboxRequest, payload []byte) {
	if !strings.HasPrefix(d.Object.Object, getFullURLPrefix(c)) {
		log.Error().Bytes("payload", payload).Msg("Inbox request objectID points to different host")
		return
	}
	actor, err := apFetchActor(d.Actor)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to fetch actor information")
		return
	}
	if err := apCheckSignature(c, actor.PubKey.PublicKeyPem, payload); err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to validate actor signature")
		return
	}
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	data, err := json.Marshal(apFollowResponseItem{
		Context: "https://www.w3.org/ns/activitystreams",
		ID:      getFullURL(c, "/"+uuid.New().String()),
		Type:    "Accept",
		Actor:   d.Object.ID,
		Object: apFollowResponseObject{
			ID:     d.ID,
			Type:   unfollowAction,
			Actor:  d.Actor,
			Object: d.Object.ID,
		},
	})
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to serialize AP inbox response")
		return
	}
	err = apSendSignedPostRequest(actor.Inbox, d.Object.Object+"#key", data, key)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to send HTTP request")
		return
	}
	u, err := url.Parse(d.Object.Object)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to parse URL")
		return
	}
	err = model.DB.Delete(&model.APFollower{}, "name = ? and filter = ?", d.Actor, u.RawQuery).Error
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to delete AP follower")
		return
	}
}

func apNotifyFollowers(c *gin.Context, b *model.Bookmark, s *model.Snapshot) {
	if !b.Public {
		return
	}
	var followers []*model.APFollower
	err := model.DB.Model(&model.APFollower{}).Find(&followers).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch followers")
		return
	}
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	for _, f := range followers {
		kv, err := url.ParseQuery(f.Filter)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse URL")
			return
		}
		match := true
		if kv.Get("query") != "" {
			q := strings.ToLower(kv.Get("query"))
			textFound := false
			if kv.Get("search_in_note") != "" && strings.Contains(strings.ToLower(b.Notes), q) {
				textFound = true
			}
			if kv.Get("search_in_snapshot") != "" && strings.Contains(strings.ToLower(s.Text), q) {
				textFound = true
			}
			if strings.Contains(strings.ToLower(b.Title), q) {
				textFound = true
			}
			if !textFound {
				match = false
			}
		}
		if kv.Get("owner") != "" {
			if kv.Get("owner") != b.User.Username {
				match = false
			}
		}
		if kv.Get("domain") != "" {
			if kv.Get("domain") != b.Domain {
				match = false
			}
		}
		if kv.Get("tag") != "" {
			tagFound := false
			for _, t := range b.Tags {
				if t.Text == kv.Get("tag") {
					tagFound = true
					break
				}
			}
			if !tagFound {
				match = false
			}
		}
		if !match {
			continue
		}

		u := getFullURL(c, URLFor("Public bookmarks"))
		if f.Filter != "" {
			u = u + "?" + f.Filter
		}
		item := apCreateBookmarkItem(c, b, u)
		if item == nil {
			continue
		}
		actor, err := apFetchActor(f.Name)
		if err != nil {
			log.Error().Err(err).Msg("Failed to fetch actor")
			continue
		}
		item.To = append(item.To, actor.ID)
		item.Object.To = append(item.Object.To, actor.ID)
		data, err := json.Marshal(item)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal bookmark")
			continue
		}
		err = apSendSignedPostRequest(actor.Inbox, u+"#key", data, key)
		if err != nil {
			log.Error().Err(err).Str("actor", f.Name).Msg("Failed to send HTTP request")
			continue
		}
		log.Debug().Str("actor", f.Name).Str("filter", f.Filter).Msg("Bookmark sent to inbox")
	}
	return
}

func apParseSigHeader(c *gin.Context, digest string) (string, []byte, error) {
	sh := c.Request.Header.Get("Signature")
	sigParts := apSigHeaderRe.FindStringSubmatch(sh)
	if len(sigParts) < 4 {
		return "", nil, errors.New("invalid Signature header format: " + sh)
	}
	signature := sigParts[3]
	sigHeaders := make([]string, 0, 3)
	for _, h := range strings.Split(sigParts[2], " ") {
		if h == "(request-target)" {
			method := strings.ToLower(c.Request.Method)
			// TODO perhaps it is better to use the path information from
			// the ActivityPub request payload, because c.Request.URL.Path
			// can be different in case of some weird reverse proxying shenanigans
			s := fmt.Sprintf("(request-target): %s %s", method, c.Request.URL.Path)
			if c.Request.URL.RawQuery != "" {
				s += "?" + c.Request.URL.RawQuery
			}
			sigHeaders = append(sigHeaders, s)
		} else if h == "host" {
			u, err := url.Parse(getFullURL(c, "/"))
			if err != nil {
				return "", nil, err
			}
			sigHeaders = append(sigHeaders, "host: "+u.Host)
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
	sigH := []byte(strings.Join(sigHeaders, "\n"))
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
		log.Error().Err(err).Msg("Failed to sign request data")
		return err
	}
	sigHeader := fmt.Sprintf(`keyId="%s",headers="(request-target) host date digest",signature="%s",algorithm="rsa-sha256"`, keyID, base64.StdEncoding.EncodeToString(sig))
	cli := &http.Client{Timeout: apRequestTimeout}
	req, err := http.NewRequest("POST", us, bytes.NewReader(data))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create signed POST request")
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
		log.Error().Err(err).Msg("Failed to send accept response")
		return err
	}
	defer r.Body.Close()
	rb, _ := io.ReadAll(r.Body)
	if bytes.Contains(rb, []byte("error")) {
		log.Error().Bytes("payload", rb).Msg("Accept response contains error")
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
		Aliases: []string{u},
		Links: []apLink{
			apLink{Rel: "self", Type: "application/activity+json", Href: u},
			apLink{Rel: "http://webfinger.net/rel/profile-page", Type: "text/html", Href: u},
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize JSON")
	}
	_, err = c.Writer.Write(j)
	if err != nil {
		log.Error().Msg("Failed to write response")
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
	if c.ID != "" {
		return json.Marshal(c.ID)
	}
	return json.Marshal(c.Parts)
}
