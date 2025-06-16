package feed

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/asciimoo/omnom/model"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
)

func Update() error {
	feeds, err := model.GetFeeds()
	if err != nil {
		return err
	}
	for _, f := range feeds {
		updateRSSFeed(f)
	}
	return nil
}

func updateRSSFeed(f *model.Feed) {
	fp := gofeed.NewParser()
	rss, err := fp.ParseURL(f.URL)
	if err != nil {
		log.Error().Err(err).Str("URL", f.URL).Msg("Failed to fetch feed")
		return
	}
	var added int64
	for _, i := range rss.Items {
		added += model.AddFeedItem(&model.FeedItem{
			Title:   i.Title,
			Content: i.Content,
			URL:     i.Link,
			FeedID:  f.ID,
		})
	}
	log.Debug().Int64("items", added).Str("feed", f.Name).Msg("Feed fetched")
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
