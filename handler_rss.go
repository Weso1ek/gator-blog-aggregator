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
	"log"
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
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println("Couldn't get next feeds to fetch", err)
		return
	}
	log.Println("Found a feed to fetch!")
	scrapeFeed(s.db, feed)
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Name, err)
		return
	}
	for _, item := range feedData.Channel.Item {
		fmt.Printf("Found post: %s\n", item.Title)
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}

func handlerGetFeeds(s *state, cmd command) error {

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, j := range feeds {
		fmt.Println(j.Name)
		fmt.Println(j.Url)
		fmt.Println(j.Name_2)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, currentUser database.User) error {
	name := cmd.Args[0]
	url := cmd.Args[1]

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

	_, errCreateFollow := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		FeedID:    feed.ID,
		UserID:    currentUser.ID,
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})

	if errCreateFollow != nil {
		return errCreateFollow
	}

	fmt.Println(feed.Name)
	fmt.Println(feed.Url)

	return nil
}

func handlerFollow(s *state, cmd command, currentUser database.User) error {
	url := cmd.Args[0]

	feedUrl, err := s.db.GetFeedsByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	_, errCreate := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		FeedID:    feedUrl.ID,
		UserID:    currentUser.ID,
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})

	if errCreate != nil {
		return errCreate
	}

	fmt.Println(feedUrl.Name)
	fmt.Println(currentUser.Name)

	return nil
}

func handlerUnfollow(s *state, cmd command, currentUser database.User) error {
	url := cmd.Args[0]

	feedUrl, err := s.db.GetFeedsByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	errDelete := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		FeedID: feedUrl.ID,
		UserID: currentUser.ID,
	})

	if errDelete != nil {
		return errDelete
	}

	return nil
}

func handlerFollowing(s *state, cmd command, currentUser database.User) error {
	fmt.Println(currentUser.Name)
	fmt.Println(currentUser.ID)

	currentUserFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)

	if err != nil {
		return err
	}

	for _, j := range currentUserFeeds {
		fmt.Println(j.FeedName)
		fmt.Println(j.UserName)
	}

	return nil
}
