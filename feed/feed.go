package feed

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/asciimoo/omnom/model"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

var htmlSanitizerPolicy *bluemonday.Policy

func init() {
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
	p.AllowAttrs("src", "srcset").OnElements("img", "source")
	p.AllowAttrs("alt").Globally()
	p.AllowAttrs("title").Globally()
	p.RequireParseableURLs(true)
	p.AllowDataURIImages()
	p.AllowImages()
	p.AllowTables()
	p.AllowURLSchemes("mailto", "http", "https")
	htmlSanitizerPolicy = p
}

func Update() error {
	log.Debug().Msg("Updating feeds")
	feeds, err := model.GetFeeds()
	if err != nil {
		return err
	}
	for _, f := range feeds {
		updateRSSFeed(f)
	}
	return nil
}

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
		added += model.AddFeedItem(&model.FeedItem{
			Title:   i.Title,
			Content: sanitizeHTML(pu, i.Content),
			URL:     i.Link,
			FeedID:  f.ID,
		})
	}
	log.Debug().Int64("new items", added).Str("feed", f.Name).Msg("Feed updated")
}

func createFeed(name, u string) (*model.Feed, error) {
	f := &model.Feed{
		Name: name,
		URL:  u,
	}
	// TODO parse feed URL if u is HTML
	// TODO add support for ActivityPub feeds
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(u)
	if err != nil {
		return nil, errors.New("unsupported feed type; " + err.Error())
	} else {
		f.Type = model.RSSFeed
	}
	if feed.Image != nil {
		f.Favicon = fetchImageAsInlineURL(feed.Image.URL)
	} else {
		f.Favicon = fetchImageAsInlineURL(getFaviconURL(u))
	}
	return f, model.DB.Create(f).Error
}

func AddFeed(name, u string, uid uint) error {
	f, err := model.GetFeedByURL(u)
	if f == nil || err != nil {
		var err error
		f, err = createFeed(name, u)
		if err != nil {
			return err
		}
	}
	err = createUserFeed(name, f, uid)
	if err != nil {
		return err
	}
	switch f.Type {
	case model.RSSFeed:
		updateRSSFeed(f)
	default:
		log.Error().Str("Type", f.Type).Msg("Unsupported feed type")
	}
	return nil
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

func fetchImageAsInlineURL(u string) string {
	if u == "" {
		return ""
	}
	r, err := http.Get(u) //nolint: gosec //safe url
	if err != nil {
		return ""
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("data:%s;base64,%s", r.Header.Get("Content-Type"), base64.StdEncoding.EncodeToString(data))
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
	doc, err := html.Parse(strings.NewReader(h))
	if err != nil {
		log.Debug().Err(err).Str("HTML", h).Msg("Failed to parse HTML")
		return ""
	}
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}
		for _, a := range n.Attr {
			if a.Key == "src" || a.Key == "href" {
				a.Val = toFullURL(base, a.Val)
			}
		}
	}
	var out strings.Builder
	err = html.Render(&out, doc)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to render HTML")
	}
	return out.String()
}

func toFullURL(base *url.URL, u string) string {
	pu, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return base.ResolveReference(pu).String()
}
