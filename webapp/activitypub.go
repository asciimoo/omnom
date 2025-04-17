package webapp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

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
	Context   string   `json:"@context"`
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Actor     string   `json:"actor"`
	To        []string `json:"to"`
	Cc        []string `json:"cc"`
	Published string   `json:"published"`
	Object    apObject `json:"object"`
}

type apObject struct {
	Context      string            `json:"@context"`
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Content      string            `json:"content"`
	URL          string            `json:"url"`
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

const contentTpl = `<h1>%[1]s</h1>
<p>%[2]s</p>
<small>Bookmarked by <a href="https://github.com/asciimoo/omnom">Omnom</a> - <a href="%[3]s">view bookmark</a></small>`

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
	log.Println(len(bs), bc)
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
		id := fmt.Sprintf("%s%s?id=%d", baseU, URLFor("Bookmark"), b.ID)
		actor := fmt.Sprintf("%s/@%s", baseU, b.User.Username)
		published := b.CreatedAt.Format(time.RFC3339)
		item := apItem{
			Context: "https://www.w3.org/ns/activitystreams",
			ID:      id + "/activity",
			Type:    "Create",
			Actor:   actor,
			To: []string{
				"https://www.w3.org/ns/activitystreams#Public",
			},
			Cc:        []string{},
			Published: published,
			Object: apObject{
				Context:      "https://www.w3.org/ns/activitystreams",
				ID:           id,
				Type:         "Note",
				Content:      fmt.Sprintf(contentTpl, b.Title, b.Notes, id),
				URL:          b.URL,
				AttributedTo: actor,
				To: []string{
					"https://www.w3.org/ns/activitystreams#Public",
				},
				Cc:        []string{},
				Published: published,
				Tag:       make([]apTag, len(b.Tags)),
			},
		}
		for i, t := range b.Tags {
			tagBase := fmt.Sprintf("%s%s?tag=", baseU, URLFor("Public bookmarks"))
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
	u, err := parseURL(baseU + c.Request.URL.String())
	if err != nil {
		log.Println("ActivityPub URL parse error", err)
		return
	}
	_, err = c.Writer.WriteString(
		fmt.Sprintf(`{
	"@context": "https://www.w3.org/ns/activitystreams",
	"id": "%[4]s",
	"type": "Application",
	"following": "",
	"followers": "",
	"inbox": "",
	"outbox": "%[1]s",
	"preferredUsername": "%[4]s",
	"name": "%[2]s",
	"summary": "",
	"url": "%[3]s/",
	"discoverable": true,
	"memorial": false,
	"icon": {
	  "type": "Image",
	  "mediaType": "image/png",
	  "url": "%[3]s/static/icons/addon_icon.png"
	}
}`,
			addURLParam(u.String(), "format=activitypub"),
			p.String(),
			baseU,
			baseU+"/"+p.String(),
		),
	)
	if err != nil {
		log.Println("ActivityPub ident write error")
	}
}
