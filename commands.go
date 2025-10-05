package main

import (
	"context"
	"fmt"
	"os"
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
	user := database.CreateUserParams{ID: int32(id), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: name}

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

func commandAgg(_s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments provided")
	}
	ctx := context.Background()
	feedURL := "https://www.wagslane.dev/index.xml"
	rssFeed, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}
	fmt.Println(rssFeed)
	return nil
}

func commandAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("not enough arguments provided")
	}
	ctx := context.Background()
	currentUser, err := s.db.GetUser(ctx, s.cfg.Username)
	if err != nil {
		return err
	}
	feedName := cmd.arguments[0]
	feedURL := cmd.arguments[1]
	id := uuid.New().ID()
	newFeed := database.CreateFeedParams{ID: int32(id), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: feedName, Url: feedURL, UserID: currentUser.ID}

	addedFeed, err := s.db.CreateFeed(ctx, newFeed)
	if err != nil {
		return err
	}

	fmt.Println(addedFeed)
	return nil
}
