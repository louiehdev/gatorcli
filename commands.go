package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/lib/pq"

	config "github.com/louiehdev/gatorcli/internal/config"

	database "github.com/louiehdev/gatorcli/internal/database"

	uuid "github.com/google/uuid"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type commandHandler func(s *state, cmd command) error

type command struct {
	name      string
	arguments []string
}

type commands struct {
	commandMap map[string]commandHandler
}

func (c *commands) run(s *state, cmd command) error {
	if err := c.commandMap[cmd.name](s, cmd); err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f commandHandler) error {
	c.commandMap[name] = f
	return nil
}

func middlewareLoggedIn(authedCommand func(s *state, cmd command, user database.User) error) commandHandler {
	return func(s *state, cmd command) error {
		ctx := context.Background()
		currentUser, err := s.db.GetUser(ctx, s.cfg.Username)
		if err != nil {
			return err
		}
		cmdErr := authedCommand(s, cmd, currentUser)
		return cmdErr
	}
}

func commandLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("username required")
	}
	ctx := context.Background()
	username := cmd.arguments[0]

	if _, err := s.db.GetUser(ctx, username); err != nil {
		fmt.Println("Username not recognized in database")
		os.Exit(1)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}
	fmt.Println("User has been set")
	return nil
}

func commandRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("name argument required")
	}
	ctx := context.Background()
	name := cmd.arguments[0]
	id := uuid.New().ID()
	user := database.CreateUserParams{
		ID:        int32(id),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name}

	registeredUser, err := s.db.CreateUser(ctx, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if string(pqErr.Code) == "23505" {
				fmt.Println("User already exists")
				os.Exit(1)
			}
		}
		return err
	}
	if err := s.cfg.SetUser(name); err != nil {
		return err
	}
	fmt.Printf("User data: %v", registeredUser)
	return nil
}

func commandUsers(s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments provided")
	}
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		if user == s.cfg.Username {
			fmt.Printf("%s (current)\n", user)
		} else {
			fmt.Println(user)
		}
	}

	return nil
}

func commandReset(s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments provided")
	}
	ctx := context.Background()
	if err := s.db.ResetUsers(ctx); err != nil {
		fmt.Printf("Reset unsuccessful: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Reset successful")
	os.Exit(0)
	return nil
}

func commandAgg(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("incorrect arguments provided")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenRequests)
	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func commandAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("not enough arguments provided")
	}
	ctx := context.Background()
	feedName := cmd.arguments[0]
	feedURL := cmd.arguments[1]
	newFeed := database.CreateFeedParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedURL,
		UserID:    user.ID}

	addedFeed, err := s.db.CreateFeed(ctx, newFeed)
	if err != nil {
		return err
	}
	newFeedFollow := database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    addedFeed.ID}

	if _, followErr := s.db.CreateFeedFollow(ctx, newFeedFollow); followErr != nil {
		return followErr
	}

	fmt.Println(addedFeed)
	return nil
}

func commandFeeds(s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments provided")
	}
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Printf("Feed Name: %s | URL: %s | User Name: %s\n", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func commandFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("incorrect arguments provided")
	}
	ctx := context.Background()
	feed, err := s.db.GetFeedFromURL(ctx, cmd.arguments[0])
	if err != nil {
		return err
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	feedFollow, err := s.db.CreateFeedFollow(ctx, feedFollowParams)
	if err != nil {
		return err
	}
	fmt.Printf("%s followed %s", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func commandFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("too many arguments provided")
	}
	ctx := context.Background()
	feedFollows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}
	fmt.Printf("User %s is currently following:\n", user.Name)
	for _, feedFollow := range feedFollows {
		fmt.Printf(" - %s\n", feedFollow.FeedName)
	}

	return nil
}

func commandUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("incorrect arguments provided")
	}
	ctx := context.Background()
	feed, err := s.db.GetFeedFromURL(ctx, cmd.arguments[0])
	if err != nil {
		return err
	}
	if err := s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{UserID: user.ID, FeedID: feed.ID}); err != nil {
		return err
	}
	return nil
}

func commandBrowse(s *state, cmd command, user database.User) error {
	var limit int32
	if len(cmd.arguments) == 1 {
		limitarg, err := strconv.Atoi(cmd.arguments[0])
		if err != nil {
			return err
		}
		limit = int32(limitarg)
	} else if len(cmd.arguments) < 1 {
		limit = 2
	} else {
		return fmt.Errorf("too many arguments provided")
	}

	ctx := context.Background()
	posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{UserID: user.ID, Limit: limit})
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Println(post)
	}

	return nil
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
	for _, item := range rssFeed.Channel.Items[:3] {
		publishTime, _ := time.Parse(time.RFC1123Z, item.PubDate)
		newPost := database.CreatePostParams{
			ID:          int32(uuid.New().ID()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PublishedAt: publishTime,
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			FeedID:      nextFeed.ID,
		}
		if _, err := s.db.CreatePost(ctx, newPost); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if string(pqErr.Code) == "23505" {
					fmt.Println("Post already exists, ignoring error for now")
					return nil
				}
			}
			return err
		}
	}
	return nil
}
