package webapp

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	ap "github.com/asciimoo/omnom/activitypub"
	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/feed"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	announceAction = "Announce"
	createAction   = "Create"
	followAction   = "Follow"
	noteAction     = "Note"
	unfollowAction = "Undo"
	likeAction     = "Like"
)

const contentTpl = `<h1><a href="%[1]s">%[2]s</a></h1>
%[3]s
Bookmarked by <a href="https://github.com/asciimoo/omnom">Omnom</a> - <a href="%[4]s">view bookmark</a>`

var apSigHeaderRe = regexp.MustCompile(`keyId="([^"]+)".*,headers="([^"]+)",.*signature="([^"]+)"`)

func apOutboxResponse(c *gin.Context) {
	user := model.GetUser(c.Param("username"))
	if user == nil {
		log.Debug().Msg("Unknown user")
		notFoundView(c)
		return
	}
	var bs []*model.Bookmark
	var bc int64
	if err := model.DB.Model(&model.Bookmark{}).Where("bookmarks.public = 1 AND bookmarks.user_id = ?", user.ID).Count(&bc).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch bookmarks")
		notFoundView(c)
		return
	}
	//nolint: gosec // conversion is safe
	if err := model.DB.Limit(int(resultsPerPage)).Where("bookmarks.public = 1 AND bookmarks.user_id = ?", user.ID).Preload("Tags").Find(&bs).Error; err != nil {
		log.Error().Err(err).Msg("Failed to count bookmarks")
		notFoundView(c)
		return
	}
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	u := getFullURL(c, URLFor("User", user.Username))
	resp := ap.Outbox{
		Context:      "https://www.w3.org/ns/activitystreams",
		ID:           u,
		Type:         "OrderedCollection",
		Summary:      "Recent bookmarks of " + u,
		TotalItems:   bc,
		OrderedItems: make([]*ap.OutboxItem, len(bs)),
	}
	for i, b := range bs {
		item := apCreateBookmarkItem(c, b, u)
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

func apCreateBookmarkItem(c *gin.Context, b *model.Bookmark, actor string) *ap.OutboxItem {
	id := getFullURL(c, fmt.Sprintf("%s?id=%d", URLFor("Bookmark"), b.ID))
	published := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	title := truncate(b.Title, 300)
	body := ""
	if b.Notes != "" {
		body = fmt.Sprintf("<p>%s</p>", truncate(b.Notes, 350-len(title)))
	}
	item := &ap.OutboxItem{
		Context: "https://www.w3.org/ns/activitystreams",
		ID:      id + "#activity",
		Type:    createAction,
		Actor:   actor,
		To: []string{
			"https://www.w3.org/ns/activitystreams#Public",
		},
		Cc:        []string{},
		Published: published,
		Object: ap.OutboxObject{
			ID:           id,
			Type:         noteAction,
			Summary:      "",
			Content:      fmt.Sprintf(contentTpl, b.URL, title, body, id),
			URL:          b.URL,
			URI:          b.URL,
			AttributedTo: actor,
			To: []string{
				"https://www.w3.org/ns/activitystreams#Public",
			},
			Cc:        []string{},
			Published: published,
			Tag:       make([]ap.Tag, len(b.Tags)),
			Replies:   map[string]string{},
		},
	}
	for i, t := range b.Tags {
		tagBase := getFullURL(c, fmt.Sprintf("%s?tag=", URLFor("Public bookmarks")))
		item.Object.Tag[i] = ap.Tag{
			Type: "Hashtag",
			Href: tagBase + t.Text,
			Name: t.Text,
		}
	}
	return item
}

func apIdentityResponse(c *gin.Context, user *model.User) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	id := getFullURL(c, c.Request.URL.String())
	cfg, _ := c.Get("config")
	pk, err := cfg.(*config.Config).ActivityPub.ExportPubKey()
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize JSON")
		return
	}
	j, err := json.Marshal(ap.Identity{
		Context: &ap.Context{
			Parts: []any{
				"https://www.w3.org/ns/activitystreams",
				"https://w3id.org/security/v1",
			},
		},
		ID:                id,
		Type:              "Person",
		Inbox:             getFullURL(c, URLFor("ActivityPub inbox", user.Username)),
		Outbox:            getFullURL(c, URLFor("ActivityPub outbox", user.Username)),
		PreferredUsername: user.Username,
		Name:              user.Username,
		URL:               id,
		Discoverable:      true,
		Icon: &ap.Image{
			Type:      "Image",
			MediaType: "image/png",
			URL:       getFullURL(c, "/static/icons/addon_icon.png"),
		},
		Image: &ap.Image{
			Type:      "Image",
			MediaType: "image/png",
			URL:       getFullURL(c, "/static/icons/addon_icon.png"),
		},
		PubKey: ap.PubKey{
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
	d := &ap.InboxRequest{}
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
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	actor, err := ap.FetchActor(d.Actor, d.Object.ID+"#key", key)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to fetch actor")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Network error",
		})
		return
	}
	if err := apCheckSignature(c, actor, body); err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to validate signature")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid signature",
		})
		return
	}
	switch d.Type {
	case followAction:
		go apInboxFollowResponse(c, d, actor)
	case unfollowAction:
		go apInboxUnfollowResponse(c, d, actor)
	case createAction:
		go apInboxCreateResponse(c, d)
	case announceAction:
		go apInboxAnnounceResponse(c, d)
	case likeAction:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Not supported",
		})
		return
	default:
		log.Debug().Str("type", d.Type).Bytes("msg", body).Msg("Unhandled ActivityPub inbox message")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Unknown action type",
		})
		return
	}
	c.JSON(http.StatusOK, map[string]string{
		"status": "OK",
	})
}

func apInboxAnnounceResponse(c *gin.Context, d *ap.InboxRequest) {
	obj, err := ap.FetchObject(d.Object.ID)
	if err != nil {
		log.Error().Err(err).Str("ID", d.Object.ID).Msg("Failed to fetch object")
		return
	}
	if obj.Type != "Note" {
		log.Error().Err(err).Str("Type", obj.Type).Msg("Unsupported object type")
		return
	}
	d.Object = obj
	apInboxCreateResponse(c, d)
}

func apInboxCreateResponse(c *gin.Context, d *ap.InboxRequest) {
	if d.Object.Type != noteAction {
		return
	}
	user := model.GetUser(c.Param("username"))
	if user == nil {
		log.Debug().Msg("Unknown user")
		notFoundView(c)
		return
	}
	f, err := model.GetFeedByURL(d.Actor)
	if err != nil {
		log.Error().Err(err).Str("URL", d.Actor).Msg("No feed found")
		return
	}
	cfg, _ := c.Get("config")
	err = feed.AddActivityPubFeedItem(cfg.(*config.Config), f, user, d)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add AP feed item")
		return
	}
}

func apInboxFollowResponse(c *gin.Context, d *ap.InboxRequest, actor *ap.Identity) {
	user := model.GetUser(c.Param("username"))
	if user == nil {
		log.Debug().Msg("Unknown user")
		notFoundView(c)
		return
	}
	if !strings.HasPrefix(d.Object.ID, getFullURLPrefix(c)) {
		log.Error().Msg("Inbox request objectID points to different host")
		return
	}
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	data, err := json.Marshal(ap.FollowResponseItem{
		Context: "https://www.w3.org/ns/activitystreams",
		ID:      getFullURL(c, "/"+uuid.New().String()),
		Type:    "Accept",
		Actor:   d.Object.ID,
		Object: ap.FollowResponseObject{
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
	err = ap.SendSignedPostRequest(actor.Inbox, d.Object.ID+"#key", data, key)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to send HTTP request")
		return
	}
	err = model.CreateAPFollower(user.ID, d.Actor)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to create AP follower")
		return
	}
}
func apInboxUnfollowResponse(c *gin.Context, d *ap.InboxRequest, actor *ap.Identity) {
	user := model.GetUser(c.Param("username"))
	if user == nil {
		log.Debug().Msg("Unknown user")
		notFoundView(c)
		return
	}
	if !strings.HasPrefix(d.Object.Object, getFullURLPrefix(c)) {
		log.Error().Msg("Inbox request objectID points to different host")
		return
	}
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	data, err := json.Marshal(ap.FollowResponseItem{
		Context: "https://www.w3.org/ns/activitystreams",
		ID:      getFullURL(c, "/"+uuid.New().String()),
		Type:    "Accept",
		Actor:   d.Object.ID,
		Object: ap.FollowResponseObject{
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
	err = ap.SendSignedPostRequest(actor.Inbox, d.Object.Object+"#key", data, key)
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to send HTTP request")
		return
	}
	err = model.DB.Delete(&model.APFollower{}, "user_id= ? and follower = ?", user.ID, d.Actor).Error
	if err != nil {
		log.Error().Err(err).Str("actor", d.Actor).Msg("Failed to delete AP follower")
		return
	}
}

func apNotifyFollowers(c *gin.Context, b *model.Bookmark) {
	if !b.Public {
		return
	}
	var followers []*model.APFollower
	err := model.DB.Model(&model.APFollower{}).Where("user_id = ?", b.UserID).Find(&followers).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch followers")
		return
	}
	cfg, _ := c.Get("config")
	key := cfg.(*config.Config).ActivityPub.PrivK
	for _, f := range followers {
		u := getFullURL(c, URLFor("User", b.User.Username))
		item := apCreateBookmarkItem(c, b, u)
		if item == nil {
			continue
		}
		actor, err := ap.FetchActor(f.Follower, u+"#key", key)
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
		err = ap.SendSignedPostRequest(actor.Inbox, u+"#key", data, key)
		if err != nil {
			log.Error().Err(err).Str("actor", f.Follower).Msg("Failed to send HTTP request")
			continue
		}
		log.Debug().Str("actor", f.Follower).Msg("Bookmark sent to inbox")
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
	for h := range strings.SplitSeq(sigParts[2], " ") {
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

func apCheckSignature(c *gin.Context, actor *ap.Identity, payload []byte) error {
	pubKey := actor.PubKey.PublicKeyPem
	pHash := sha256.Sum256(payload)
	digest := fmt.Sprintf("SHA-256=%s", base64.StdEncoding.EncodeToString(pHash[:]))
	signature, hash, err := apParseSigHeader(c, digest)
	if err != nil {
		return fmt.Errorf("failed to parse signature header: %w", err)
	}
	pb, _ := pem.Decode([]byte(pubKey))
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

func apWebfingerResponse(c *gin.Context) {
	c.Header("Content-Type", "application/activity+json; charset=utf-8")
	s := c.Query("resource")
	sParts := strings.Split(strings.TrimPrefix(s, "acct:"), "@")
	uname := sParts[0]
	if uname == "" && len(sParts) > 1 {
		uname = sParts[1]
	}
	u := getFullURL(c, URLFor("User", uname))
	j, err := json.Marshal(ap.Webfinger{
		Subject: s,
		Aliases: []string{u},
		Links: []ap.Link{
			ap.Link{Rel: "self", Type: "application/activity+json", Href: u},
			ap.Link{Rel: "http://webfinger.net/rel/profile-page", Type: "text/html", Href: u},
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
