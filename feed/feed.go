// Package feed handles RSS/Atom and ActivityPub feed processing and aggregation.
//
// This package provides functionality for subscribing to, fetching, and processing
// content from various feed types:
//   - RSS 2.0 feeds
//   - Atom feeds
//   - ActivityPub (Mastodon, Pleroma, etc.)
//
// Feed items are fetched periodically via UpdateLoop and stored in the database.
// Each user can subscribe to multiple feeds, and feed items are marked as read/unread
// per user. The package handles:
//   - Feed discovery from URLs (including HTML link rel="alternate")
//   - Content sanitization to prevent XSS attacks
//   - Resource downloading and local storage
//   - ActivityPub federation (following actors, receiving posts)
//   - URL resolution and normalization
//   - Duplicate detection
//
// HTML content from feeds is sanitized using a whitelist-based policy that allows
// safe tags and attributes while removing scripts and event handlers. Embedded
// images and resources are downloaded and stored locally.
//
// Example usage:
//
//	// Subscribe to a feed
//	err := feed.AddFeed(cfg, "Hacker News", "https://news.ycombinator.com/rss", userID)
//
//	// Update all feeds
//	err := feed.Update()
//
//	// Run periodic updates
//	go feed.UpdateLoop()
package feed

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	ap "github.com/asciimoo/omnom/activitypub"
	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const imgTplStr = `<hr /><img src="{{ .Src }}" alt="{{ .Alt }}" />`

var imgTpl *template.Template

var htmlSanitizerPolicy *bluemonday.Policy

var errUnknownFeedType = errors.New("unknown feed type")

const (
	srcAttr = "src"
)

func init() {
	var err error
	imgTpl = template.New("image template")
	imgTpl, err = imgTpl.Parse(imgTplStr)
	if err != nil {
		panic(err)
	}
	p := bluemonday.NewPolicy()
	p.AllowElements(
		"a",
		"abbr",
		"b",
		"br",
		"canvas",
		"caption",
		"center",
		"cite",
		"code",
		"del",
		"details",
		"div",
		"dt",
		"em",
		"figcaption",
		"figure",
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
		"hr",
		"i",
		"img",
		"ins",
		"kbd",
		"label",
		"li",
		"math",
		"marquee",
		"media",
		"mediagroup",
		"noscript",
		"ol",
		"p",
		"pre",
		"source",
		"span",
		"strong",
		"sub",
		"summary",
		"sup",
		"svg",
		"table",
		"tbody",
		"td",
		"tfoot",
		"th",
		"thead",
		"title",
		"tr",
		"tt",
		"u",
		"ul",
		"video",
	)
	p.AllowStyles(
		"text-decoration",
		"color",
		"background",
		"background-color",
		"background-image",
		"font-size",
		"text-align",
	).Globally()
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs(srcAttr, "srcset").OnElements("img", "source")
	p.AllowAttrs("alt").Globally()
	p.AllowAttrs("title").Globally()
	p.RequireParseableURLs(true)
	p.AllowDataURIImages()
	p.AllowImages()
	p.AllowTables()
	p.RequireNoFollowOnLinks(false)
	p.AllowURLSchemes("mailto", "http", "https")
	htmlSanitizerPolicy = p
}

// Update fetches and updates all feeds.
func Update() error {
	feeds, err := model.GetFeeds()
	if err != nil {
		return err
	}
	if len(feeds) == 0 {
		return nil
	}
	log.Debug().Msg("Updating feeds")
	for _, f := range feeds {
		switch model.FeedType(f.Type) {
		case model.RSSFeed:
			updateRSSFeed(f)
		case model.ActivityPubFeed:
			continue
		default:
			log.Error().Err(errUnknownFeedType).Str("type", f.Type).Str("url", f.URL).Msg("Failed to update feed")
		}
	}
	return nil
}

// UpdateLoop runs a periodic feed update loop.
func UpdateLoop() {
	interval := 60 * time.Minute
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		err := Update()
		if err != nil {
			log.Error().Err(err).Msg("Failed to update feeds")
		}
	}
}

func updateRSSFeed(f *model.Feed) {
	fp := gofeed.NewParser()
	pu, err := url.Parse(f.URL)
	if err != nil {
		log.Error().Err(err).Str("URL", f.URL).Msg("Failed to parse feed URL")
		return
	}
	rss, err := fp.ParseURL(f.URL)
	if err != nil {
		log.Error().Err(err).Str("URL", f.URL).Msg("Failed to fetch feed")
		return
	}
	var added int64
	for _, i := range rss.Items {
		i.Link = resolveURL(pu, i.Link)
		fi, err := model.GetFeedItem(f.ID, i.Link)
		if fi == nil || err != nil {
			c := i.Content
			if c == "" {
				c = i.Description
			}
			fi = &model.FeedItem{
				Title:   i.Title,
				Content: sanitizeHTML(pu, c),
				URL:     i.Link,
				FeedID:  f.ID,
			}
			fi.Content, err = saveResources(fi.Content)
			if err != nil {
				log.Error().Err(err).Str("feed", f.Name).Str("URL", fi.URL).Msg("Failed to save resources for feed item")
				continue
			}
		}
		added += model.AddFeedItem(fi)
	}
	log.Debug().Int64("new items", added).Str("feed", f.Name).Msg("Feed updated")
}

// AddActivityPubFeedItem adds an ActivityPub post as a feed item.
func AddActivityPubFeedItem(cfg *config.Config, f *model.Feed, u *model.User, d *ap.InboxRequest) error {
	pu, err := url.Parse(f.URL)
	if err != nil {
		return err
	}
	uri := d.Object.ID
	if d.Object.URL != "" {
		uri = d.Object.URL
	}
	fi, err := model.GetFeedItem(f.ID, uri)
	if fi != nil && err == nil {
		return nil
	}
	c := d.Object.Content
	for _, att := range d.Object.Attachments {
		if strings.Contains(att.MediaType, "image") && att.URL != "" {
			var out strings.Builder
			err := imgTpl.Execute(&out, map[string]string{
				"Src": att.URL,
				"Alt": att.Name,
			})
			if err != nil {
				log.Debug().Err(err).Msg("Failed to render image template")
				continue
			}
			c += out.String()
		}
	}
	a := d.Actor
	fi = &model.FeedItem{
		Title:     a,
		Content:   sanitizeHTML(pu, c),
		URL:       uri,
		InReplyTo: d.Object.InReplyTo,
		Context:   d.Object.Context,
		FeedID:    f.ID,
	}
	if d.Object.AttributedTo != "" && d.Object.AttributedTo != a {
		fi.OriginalAuthorID = d.Object.AttributedTo
		userURL, err := getUserURL(cfg, u.ID)
		if err == nil {
			userKey := userURL + "#key"
			pk := cfg.ActivityPub.PrivK
			oa, err := ap.FetchActor(d.Object.AttributedTo, userKey, pk)
			if err == nil {
				fi.OriginalAuthorName = oa.GetName()
				err = oa.SaveFavicon()
				if err == nil {
					fi.Favicon = oa.GetFaviconPath()
				}
			}
		}
	}
	fi.Content, err = saveResources(fi.Content)
	if err != nil {
		return err
	}
	if model.AddFeedItem(fi) != 1 {
		return errors.New("failed to add feed item to DB")
	}
	return nil
}

func createFeed(cfg *config.Config, name, u string, uid uint) (*model.Feed, error) {
	ftype, fu, err := getFeedInfo(u)
	if err != nil {
		return nil, err
	}
	f := &model.Feed{
		Name: name,
		URL:  fu,
		Type: string(ftype),
	}
	switch ftype {
	case model.RSSFeed:
		feed, err := createRSSFeed(fu)
		if err != nil {
			return nil, err
		}
		if feed.Image != nil {
			f.Favicon = fetchImageAsInlineURL(feed.Image.URL)
		} else {
			f.Favicon = fetchImageAsInlineURL(getFaviconURL(u))
		}
	case model.ActivityPubFeed:
		userURL, err := getUserURL(cfg, uid)
		if err != nil {
			return nil, err
		}
		userKey := userURL + "#key"
		pk := cfg.ActivityPub.PrivK
		actor, err := ap.FetchActor(fu, userKey, pk)
		if err != nil {
			return nil, err
		}
		f.URL = actor.ID
		f.Author = actor.GetName()
		err = ap.SendFollowRequest(actor.Inbox, fu, userURL, pk)
		if err != nil {
			return nil, err
		}
		if actor.Icon.URL != "" {
			f.Favicon = fetchImageAsInlineURL(actor.Icon.URL)
		} else {
			f.Favicon = fetchImageAsInlineURL(getFaviconURL(u))
		}
	default:
		return nil, errUnknownFeedType
	}
	err = model.DB.Create(f).Error
	if err != nil {
		return f, err
	}
	err = createUserFeed(name, f, uid)
	return f, err
}

// AddFeed adds a new feed subscription for a user.
func AddFeed(cfg *config.Config, name, u string, uid uint) error {
	f, err := model.GetFeedByURL(u)
	if f == nil || err != nil {
		var err error
		f, err = createFeed(cfg, name, u, uid)
		if err != nil {
			return err
		}
	}
	switch model.FeedType(f.Type) {
	case model.RSSFeed:
		updateRSSFeed(f)
	case model.ActivityPubFeed:
		break
	default:
		log.Error().Err(errUnknownFeedType).Str("Type", f.Type)
	}
	return nil
}

// DeleteFeed removes a user's feed subscription.
func DeleteFeed(cfg *config.Config, uf *model.UserFeed) error {
	f, err := model.GetFeedByID(uf.FeedID)
	if err != nil {
		return err
	}
	switch model.FeedType(f.Type) {
	case model.ActivityPubFeed:
		userURL, err := getUserURL(cfg, uf.UserID)
		if err != nil {
			return err
		}
		userKey := userURL + "#key"
		pk := cfg.ActivityPub.PrivK
		actor, err := ap.FetchActor(f.URL, userKey, pk)
		if err != nil {
			log.Info().Err(err).Msg("Failed to fetch actor")
			break
		}
		err = ap.SendUnfollowRequest(actor.Inbox, userURL, pk)
		if err != nil {
			log.Info().Err(err).Msg("Failed to send unfollow request")
		}
	}
	return model.DeleteUserFeed(uf)
}

func getUserURL(cfg *config.Config, uid uint) (string, error) {
	var user model.User
	err := model.DB.Where("id = ?", uid).First(&user).Error
	if err != nil {
		return "", err
	}
	// TODO get users endpoint url from api _somehow_
	return cfg.BaseURL("/users/" + user.Username), nil
}

func createUserFeed(name string, f *model.Feed, uid uint) error {
	var uf *model.UserFeed
	if err := model.DB.Where("feed_id = ? and user_id = ?", f.ID, uid).First(uf).Error; err == nil {
		return nil
	}
	uf = &model.UserFeed{
		Name:   name,
		FeedID: f.ID,
		UserID: uid,
	}
	return model.DB.Create(uf).Error
}

func getFeedInfo(u string) (model.FeedType, string, error) {
	req, err := http.NewRequest("GET", u, nil) //nolint: gosec //safe url
	if err != nil {
		return "", "", err
	}
	//cli := &http.Client{Timeout: TODO}
	cli := &http.Client{}
	req.Header.Set("Accept", "application/rss+xml;q=0.7, application/rdf+xml;q=0.65, application/atom+xml;q=0.6, application/xml;q=0.5, text/xml;q=0.4, text/html;q=0.9, text/json;q=0.8, application/activity+json;q=1")
	resp, err := cli.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if isAPFeed(ct) {
		return model.ActivityPubFeed, u, nil
	}
	if isRSSFeed(ct) {
		return model.RSSFeed, u, nil
	}
	if strings.Contains(ct, "html") || strings.Contains(ct, "xml") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", "", err
		}
		return parseHTMLFeedInfo(u, body)
	}
	return "", "", errors.New("unknown feed type")
}

func isRSSFeed(s string) bool {
	return strings.Contains(s, "xml") || strings.Contains(s, "rss") || strings.Contains(s, "atom") || strings.Contains(s, "rdf") || strings.Contains(s, "feed+json")
}

func isAPFeed(s string) bool {
	return strings.Contains(s, "activity+json")
}

func parseHTMLFeedInfo(u string, body []byte) (model.FeedType, string, error) {
	r := bytes.NewReader(body)
	doc := html.NewTokenizer(r)
	pu, err := url.Parse(u)
	if err != nil {
		return "", "", err
	}
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			return "", "", errors.New("no feed found")
		case html.StartTagToken, html.SelfClosingTagToken:
			tn, hasAttr := doc.TagName()
			if !hasAttr || !bytes.Equal(tn, []byte("link")) {
				continue
			}
			isAlternate := false
			href := ""
			ftype := ""
			for {
				aName, aVal, moreAttr := doc.TagAttr()
				if bytes.Equal(aName, []byte("rel")) && bytes.Equal(aVal, []byte("alternate")) {
					isAlternate = true
				}
				if bytes.Equal(aName, []byte("href")) {
					href = string(aVal)
				}
				if bytes.Equal(aName, []byte("type")) {
					ftype = string(aVal)
				}
				if !moreAttr {
					break
				}
			}
			if isAlternate && href != "" && isAPFeed(ftype) {
				return model.ActivityPubFeed, resolveURL(pu, href), nil
			}
			if isAlternate && href != "" && isRSSFeed(ftype) {
				return model.RSSFeed, resolveURL(pu, href), nil
			}
		}
	}
	return "", "", errors.New("no feed found")
}

func createRSSFeed(u string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	return fp.ParseURL(u)
}

func fetchImageAsInlineURL(u string) string {
	if u == "" {
		return ""
	}
	r, err := http.Get(u) //nolint: gosec //safe url
	if err != nil {
		return ""
	}
	defer r.Body.Close()
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return ""
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("data:%s;base64,%s", ct, base64.StdEncoding.EncodeToString(data))
}

func getFaviconURL(u string) string {
	pu, err := url.Parse(u)
	if err != nil {
		// TODO parse / to check favicon path
		return ""
	}
	pu.Path = "favicon.ico"
	pu.RawQuery = ""
	return pu.String()
}

func sanitizeHTML(u *url.URL, h string) string {
	// TODO fetch resources to local storage
	return htmlSanitizerPolicy.Sanitize(resolveURLs(u, h))
}

func resolveURLs(base *url.URL, h string) string {
	if h == "" {
		return ""
	}
	frags, err := html.ParseFragment(strings.NewReader(h), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body})
	if err != nil {
		log.Debug().Err(err).Str("HTML", h).Msg("Failed to parse HTML")
		return ""
	}
	chunks := make([]string, 0, len(frags))
	for _, doc := range frags {
		resolveElementURLs(base, doc)
		for n := range doc.Descendants() {
			resolveElementURLs(base, n)
		}
		var out strings.Builder
		err = html.Render(&out, doc)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to render HTML")
			continue
		}
		chunks = append(chunks, out.String())
	}
	return strings.Join(chunks, "")
}

func resolveElementURLs(base *url.URL, n *html.Node) {
	if n.Type != html.ElementNode {
		return
	}
	for i, a := range n.Attr {
		if a.Key == srcAttr || a.Key == "href" {
			n.Attr[i] = html.Attribute{Key: a.Key, Val: resolveURL(base, a.Val)}
		}
	}
}

func resolveURL(base *url.URL, u string) string {
	pu, err := url.Parse(u)
	if err != nil {
		return ""
	}
	q := pu.Query()
	qChange := false
	for k, _ := range q {
		if k == "utm" || strings.HasPrefix(k, "utm_") {
			qChange = true
			q.Del(k)
		}
	}
	if qChange {
		pu.RawQuery = q.Encode()
	}

	return base.ResolveReference(pu).String()
}

func saveResources(h string) (string, error) {
	if h == "" {
		return "", nil
	}
	frags, err := html.ParseFragment(strings.NewReader(h), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body})
	if err != nil {
		log.Debug().Err(err).Str("HTML", h).Msg("Failed to parse HTML")
		return "", err
	}
	chunks := make([]string, 0, len(frags))
	for _, doc := range frags {
		rewriteResourceAttributes(doc)
		for n := range doc.Descendants() {
			rewriteResourceAttributes(n)
		}
		var out strings.Builder
		err = html.Render(&out, doc)
		if err != nil {
			return "", err
		}
		chunks = append(chunks, out.String())
	}
	return strings.Join(chunks, ""), nil
}
func rewriteResourceAttributes(n *html.Node) {
	if n.Type != html.ElementNode || n.DataAtom != atom.Img {
		return
	}
	for i, a := range n.Attr {
		if a.Key == srcAttr {
			key, err := saveResource(a.Val)
			if err != nil {
				// TODO how to handle?
				continue
			}
			n.Attr[i] = html.Attribute{Key: a.Key, Val: fmt.Sprintf("/static/data/resources/%s/%s", key[:2], key)}
		}
	}
}

func saveResource(u string) (string, error) {
	resp, err := http.Get(u) //nolint: gosec //safe url
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	mts, err := mime.ExtensionsByType(resp.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}
	if len(mts) < 1 {
		return "", errors.New("failed to identify file extension")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	key := storage.Hash(body) + mts[0]
	err = storage.SaveResource(key, body)
	if err != nil {
		return "", err
	}
	// TODO model.GetOrCreateResource(key, r.Mimetype, r.Filename, size)
	return key, nil
}
