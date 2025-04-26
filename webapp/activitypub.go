package webapp

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
)

type apOutbox struct {
	Context      string   `json:"@context"`
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Summary      string   `json:"summary"`
	TotalItems   int64    `json:"totalItems"`
	OrderedItems []apItem `json:"orderedItems"`
}

type apItem struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Actor     string   `json:"actor"`
	To        []string `json:"to"`
	Cc        []string `json:"cc"`
	Published string   `json:"published"`
	Object    apObject `json:"object"`
}

type apObject struct {
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
	Context           string   `json:"@context"`
	ID                string   `json:"id"`
	Type              string   `json:"type"`
	Following         string   `json:"following"`
	Followers         string   `json:"followers"`
	Inbox             string   `json:"inbox"`
	Outbox            string   `json:"outbox"`
	PreferredUsername string   `json:"preferredUsername"`
	Name              string   `json:"name"`
	Summary           string   `json:"summary"`
	URL               string   `json:"url"`
	Discoverable      bool     `json:"discoverable"`
	Memorial          bool     `json:"memorial"`
	Icon              apImage  `json:"icon"`
	Image             apImage  `json:"image"`
	Pubkey            apPubkey `json:"publicKey"`
}

type apImage struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
}

type apPubkey struct {
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
	ID string `json:"id"`
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
		OrderedItems: make([]apItem, len(bs)),
	}
	for i, b := range bs {
		id := fmt.Sprintf("%s?id=%d", URLFor("Bookmark"), b.ID)
		if strings.HasPrefix(id, "/") {
			id = baseU + id
		}
		actor := fmt.Sprintf("%s/bookmarks?owner=%s", baseU, b.User.Username)
		published := b.CreatedAt.Format(time.RFC3339)
		item := apItem{
			ID:    id + "/activity",
			Type:  "Create",
			Actor: actor,
			To: []string{
				"https://www.w3.org/ns/activitystreams#Public",
			},
			Cc:        []string{},
			Published: published,
			Object: apObject{
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
			tagBase := fmt.Sprintf("%s?tag=", URLFor("Public bookmarks"))
			if strings.HasPrefix(tagBase, "/") {
				tagBase = baseU + tagBase
			}
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

func apIdentityResponse(c *gin.Context, p *searchParams) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	baseU := getFullURLPrefix(c)
	id := baseU + c.Request.URL.String()
	u, err := parseURL(baseU + c.Request.URL.String())
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
	inbox := URLFor("ActivityPub inbox")
	if strings.HasPrefix(inbox, "/") {
		inbox = baseU + inbox
	}
	uname := "omnom" + p.String()
	j, err := json.Marshal(apIdentity{
		Context:           "https://www.w3.org/ns/activitystreams",
		ID:                id,
		Type:              "Person",
		Inbox:             inbox,
		Outbox:            addURLParam(u.String(), "format=activitypub"),
		PreferredUsername: uname,
		Name:              baseU + "/" + p.String(),
		URL:               baseU + "/",
		Discoverable:      true,
		Icon: apImage{
			Type:      "Image",
			MediaType: "image/png",
			URL:       baseU + "/static/icons/addon_icon.png",
		},
		Image: apImage{
			Type:      "Image",
			MediaType: "image/png",
			URL:       baseU + "/static/icons/addon_icon.png",
		},
		Pubkey: apPubkey{
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
		go apInboxFollowResponse(c, d)
	case "Unfollow":
		go apInboxUnfollowResponse(c, d)
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

func apInboxFollowResponse(c *gin.Context, d *apInboxRequest) {
	// TODO get actor info (inbox address) && validate d.Actor signature
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	inURL := URLFor("ActivityPub inbox")
	if strings.HasPrefix(inURL, "/") {
		inURL = getFullURLPrefix(c) + inURL
	}
	actorInbox := "TODO"
	sendSignedRequest(inURL, actorInbox, d.Object.ID+"#key", []byte("testdata"), key)
}

func apInboxUnfollowResponse(c *gin.Context, d *apInboxRequest) {
	log.Println("UNFOLLOW", d.Actor, d.Object.ID)
}

func sendSignedRequest(inURL string, targetURL, keyID string, data []byte, key *rsa.PrivateKey) {
	u, err := url.Parse(inURL)
	if err != nil {
		return
	}
	d := time.Now().UTC().Format(time.RFC3339)
	h := sha256.New()
	h.Write(data)
	hash := h.Sum(nil)
	digest := fmt.Sprintf("SHA-256=%s", base64.URLEncoding.EncodeToString(hash))
	sigData := []byte(fmt.Sprintf("(request-target): post %s\nhost: %s\ndate: %s\ndigest: %s", inURL, u.Host, d, digest))
	sigHash := sha256.Sum256(sigData)
	sig, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, []byte(sigHash[:]))
	if err != nil {
		log.Println("Can't sign request data:", err)
		return
	}
	sigHeader := fmt.Sprintf(`keyId="%s",headers="(request-target) host date digest",signature="%s",algorithm="rsa-sha256"`, keyID, base64.URLEncoding.EncodeToString(sig))
	log.Println("SIGNING DATA", sigHeader)
	return
}

func apWebfingerResponse(c *gin.Context) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	s := c.Query("resource")
	u := URLFor("Public bookmarks")
	if strings.HasPrefix(u, "/") {
		u = getFullURLPrefix(c) + u
	}
	j, err := json.Marshal(apWebfinger{
		Subject: s,
		Aliases: []string{},
		Links: []apLink{
			apLink{Rel: "self", Type: "application/activity+json", Href: addURLParam(u, "format=ap-identity")},
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

func (i *apInboxRequestObject) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' && data[len(data)-1] == '"' {
		return json.Unmarshal(data, &i.ID)
	}
	if data[0] == '{' && data[len(data)-1] == '}' {
		type T struct {
			ID string `json:"ID"`
		}
		return json.Unmarshal(data, (*T)(i))
	}
	return nil
}
