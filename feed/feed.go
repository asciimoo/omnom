package feed

import (
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
		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(f.URL)
		if err != nil {
			log.Error().Err(err).Str("URL", f.URL).Msg("Failed to fetch feed")
			continue
		}
		added := 0
		for _, i := range feed.Items {
			rows := model.AddFeedItem(&model.FeedItem{
				Title:   i.Title,
				Content: i.Content,
				URL:     i.Link,
				FeedID:  f.ID,
			}, f.URL)
			if rows > 0 {
				added += 1
			}
		}
		log.Debug().Int("items", added).Str("feed", f.Name).Msg("Feed fetched")
	}
	return nil
}
