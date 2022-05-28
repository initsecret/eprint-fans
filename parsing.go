package main

import (
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

// Convert a gofeed.Feed into a gorilla/feeds.Feed
func gofeedToGorillaFeed(goFeed *gofeed.Feed) (*feeds.Feed, error) {
	feed := &feeds.Feed{}
	feed.Items = []*feeds.Item{}

	for _, item := range goFeed.Items {

		itemAuthors := ""
		n := len(item.Authors)
		for i, author := range item.Authors {
			itemAuthors += author.Name
			if i <= (n - 3) {
				itemAuthors += ", "
			}
			if i == (n - 2) {
				itemAuthors += ", and "
			}
		}

		feed.Items = append(
			feed.Items,
			&feeds.Item{
				Title:       item.Title,
				Link:        &feeds.Link{Href: item.Link},
				Author:      &feeds.Author{Name: itemAuthors},
				Description: item.Description,
				Id:          item.GUID,
				Updated:     *item.UpdatedParsed,
				Created:     *item.PublishedParsed,
			},
		)
	}

	return feed, nil
}
