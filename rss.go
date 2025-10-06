package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/louiehdev/gatorcli/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Items       []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var feed RSSFeed
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &feed, err
	}
	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		return &feed, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &feed, err
	}
	if err := xml.Unmarshal(data, &feed); err != nil {
		return &RSSFeed{}, err
	}

	cleanedFeed := htmlCleanup(&feed)
	return cleanedFeed, nil
}

func htmlCleanup(feed *RSSFeed) *RSSFeed {
	for idx, item := range feed.Channel.Items {
		feed.Channel.Items[idx].Title = html.UnescapeString(item.Title)
		feed.Channel.Items[idx].Description = html.UnescapeString(item.Description)
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	return feed
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	nextFeed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}
	if err := s.db.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
		ID: nextFeed.ID,
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}}); err != nil {
		return err
	}
	rssFeed, err := fetchFeed(ctx, nextFeed.Url)
	if err != nil {
		return err
	}
	fmt.Printf("Feed %s Titles:\n", nextFeed.Name)
	for _, item := range rssFeed.Channel.Items {
		fmt.Printf(" - %s\n", item.Title)
	}

	return nil
}
