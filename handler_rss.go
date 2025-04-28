package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"github.com/Weso1ek/gator-blog-aggregator/internal/database"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	postReq, err := http.NewRequestWithContext( // stored to context
		ctx,
		"GET",
		feedURL,
		bytes.NewBuffer([]byte("something")),
	)

	if err != nil {
		return nil, err
	}

	postReq.Header.Set("User-Agent", "gator")

	resp, errResp := http.DefaultClient.Do(postReq)
	if errResp != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	body, errReadAll := io.ReadAll(resp.Body)

	if errReadAll != nil {
		return nil, err
	}

	var rss *RSSFeed
	if errUnmarshal := xml.Unmarshal(body, &rss); errUnmarshal != nil {
		return nil, fmt.Errorf("Unmarshal error %d", errUnmarshal)
	}

	return rss, nil
}

func handlerRssGet(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")

	if err != nil {
		return fmt.Errorf("couldn't get rss: %w", err)
	}

	for _, j := range feed.Channel.Item {
		fmt.Println(j.Title)
		fmt.Println("====")
		fmt.Println(j.Description)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	name := cmd.Args[0]
	url := cmd.Args[1]

	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	feed, errCreate := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      name,
		Url:       url,
		UserID:    currentUser.ID,
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})

	if errCreate != nil {
		return errCreate
	}

	fmt.Println(feed.Name)
	fmt.Println(feed.Url)

	return nil
}
