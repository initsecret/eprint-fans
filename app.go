package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

const EPRINT_FEED_URL = "https://eprint.iacr.org/rss/atom.xml"

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

type Feed struct {
	mu   sync.Mutex
	feed *feeds.Feed
}

// feed of recent eprints
var eprintFeed Feed

// feed of recent eprints organized by year-then-week
var eprintYearWeekFeed = make(map[int]map[int]*Feed)

func getFeed() (*feeds.Feed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(EPRINT_FEED_URL)
	if err != nil {
		return nil, err
	}
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
	// update the weekly feed
	for _, item := range feed.Items {
		year, week := item.Created.ISOWeek()
		// create a feed for the year, if it doesn't already exist
		_, ok := eprintYearWeekFeed[year]
		if !ok {
			eprintYearWeekFeed[year] = make(map[int]*Feed)
		}
		// append items to the week feed
		if eprintYearWeekFeed[year][week] == nil {
			eprintYearWeekFeed[year][week] = &Feed{feed: &feeds.Feed{}}
		}
		weekFeed := eprintYearWeekFeed[year][week]
		weekFeed.mu.Lock()
		included := false
		for _, i := range weekFeed.feed.Items {
			if i == item {
				included = true
			}
		}
		if !included {
			weekFeed.feed.Items = append(weekFeed.feed.Items, item)
		}
		weekFeed.mu.Unlock()
	}
}

func updateFeedEveryTwoHours() {
	c := time.Tick(2 * time.Hour)
	for range c {
		updateFeed()
	}
}

func serverError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "failed to retrieve page, please try again later.", http.StatusInternalServerError)
}

func weekHandler(w http.ResponseWriter, r *http.Request) {
	url := path.Clean(r.URL.Path)
	matched, err := path.Match("/week/[0-9]*/[0-9]*", url)
	if err != nil {
		serverError(w, r)
		return
	}
	if !matched {
		http.NotFound(w, r)
		return
	}
	url = url[len("/week"):]
	yearStr, weekStr := path.Split(url)
	if matched, err = path.Match("/[0-9]*/", yearStr); !matched || (err != nil) {
		serverError(w, r)
		return
	}
	// strip leading and trailing slashes
	yearStr = yearStr[1:]
	yearStr = yearStr[:len(yearStr)-1]
	year, err1 := strconv.Atoi(yearStr)
	week, err2 := strconv.Atoi(weekStr)
	if err1 != nil || err2 != nil {
		serverError(w, r)
		return
	}

	yearFeed, ok := eprintYearWeekFeed[year]
	if !ok || yearFeed == nil {
		http.NotFound(w, r)
		return
	}
	weekFeed, ok := yearFeed[week]
	if !ok || weekFeed == nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{
		"year":                    year,
		"week":                    week,
		"number":                  len(weekFeed.feed.Items),
		"listings":                weekFeed.feed.Items,
		"LastUpdated":             eprintFeed.feed.Updated.UTC().Format(time.RFC1123Z),
		"ElapsedSinceLastUpdated": strconv.Itoa(int(time.Now().UTC().Sub(eprintFeed.feed.Updated.UTC()).Minutes())),
	}
	err = t.ExecuteTemplate(w, "week-template.html", data)
	if err != nil {
		serverError(w, r)
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
		Description: "generated using eprint.fans from https://eprint.iacr.org/rss/atom.xml",
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
		return
	}
	fmt.Fprint(w, atom)
}

func styleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	err := t.ExecuteTemplate(w, "style.css", nil)
	if err != nil {
		serverError(w, r)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"LastUpdated":             eprintFeed.feed.Updated.UTC().Format(time.RFC1123Z),
		"ElapsedSinceLastUpdated": strconv.Itoa(int(time.Now().UTC().Sub(eprintFeed.feed.Updated.UTC()).Minutes())),
	}
	err := t.ExecuteTemplate(w, "index-template.html", data)
	if err != nil {
		serverError(w, r)
	}
}

func main() {
	updateFeed()
	// update feed again every 2 hours.
	go updateFeedEveryTwoHours()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/feed/", feedHandler)
	http.HandleFunc("/week/", weekHandler)
	http.HandleFunc("/style.css", styleHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
