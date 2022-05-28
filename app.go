package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

type Feed struct {
	mu   sync.Mutex
	feed *feeds.Feed
}

var eprintFeed Feed

const EPRINT_FEED_URL = "https://eprint.iacr.org/rss/atom.xml"

func getFeed() (*feeds.Feed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(EPRINT_FEED_URL)
	if err != nil {
		return nil, err
	}
	fmt.Println(feed.Title)

	return gofeedToGorillaFeed(feed)
}

func updateFeed() {
	eprintFeed.mu.Lock()
	defer eprintFeed.mu.Unlock()

	feed, err := getFeed()
	if err != nil {
		log.Printf("failed to parse rss at %s, encountered error: %s\n", time.Now().Format(time.RFC1123Z), err)
	}

	eprintFeed.feed = feed
}

func updateFeedEveryTwoHours() {
	c := time.Tick(2 * time.Hour)
	for range c {
		updateFeed()
	}
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	eprintFeed.mu.Lock()
	defer eprintFeed.mu.Unlock()

	query := r.URL.Query()
	keywords := query["keyword"]

	showAllItems := (query.Get("show_all_items") == "true")

	title := "custom eprint feed with keywords: \"" + strings.Join(keywords, "\", \"") + "\""
	if showAllItems {
		title = "full eprint feed"
	}

	customfeed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: "https://eprint.fans" + r.URL.String()},
		Description: "generated using eprint.fans from https://eprint.iacr.org/rss/rss.xml",
		Created:     eprintFeed.feed.Updated,
		Updated:     eprintFeed.feed.Updated,
	}

	customfeed.Items = []*feeds.Item{}

	for _, item := range eprintFeed.feed.Items {
		if showAllItems {
			customfeed.Items = append(
				customfeed.Items,
				&feeds.Item{
					Title:       item.Title,
					Link:        item.Link,
					Author:      item.Author,
					Description: item.Description,
					Id:          item.Id,
					Updated:     item.Updated,
					Created:     item.Created,
				},
			)
		} else {
			relevant := false
			triggeringKeyword := ""
			for _, keyword := range keywords {
				if strings.Contains(
					strings.ToLower(item.Title),
					strings.ToLower(keyword),
				) ||
					strings.Contains(
						strings.ToLower(item.Description),
						strings.ToLower(keyword),
					) ||
					strings.Contains(
						strings.ToLower(item.Author.Name),
						strings.ToLower(keyword),
					) {
					relevant = true
					triggeringKeyword = keyword
				}
			}
			if !relevant {
				continue
			}

			customfeed.Items = append(
				customfeed.Items,
				&feeds.Item{
					Title:       item.Title,
					Link:        item.Link,
					Author:      item.Author,
					Description: fmt.Sprintf("[[Triggering Keyword: \"%s\"]] ", triggeringKeyword) + item.Description,
					Id:          item.Id,
					Updated:     item.Updated,
					Created:     item.Created,
				},
			)
		}
	}
	atom, err := customfeed.ToAtom()
	if err != nil {
		fmt.Fprintf(w, "failed to generate feed with error: %s", err)
	}
	fmt.Fprint(w, atom)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"LastUpdated":             eprintFeed.feed.Updated.UTC().Format(time.RFC1123Z),
		"ElapsedSinceLastUpdated": strconv.Itoa(int(time.Now().UTC().Sub(eprintFeed.feed.Updated.UTC()).Minutes())),
	}
	err := t.ExecuteTemplate(w, "index-template.html", data)
	if err != nil {
		http.Error(w, "failed to generate page, please try again later.", http.StatusInternalServerError)
	}
}

func main() {
	feed, err := getFeed()
	if err != nil {
		panic(err)
	}
	eprintFeed = Feed{
		feed: feed,
	}
	// update feed again every 2 hours.
	go updateFeedEveryTwoHours()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/feed/", feedHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
