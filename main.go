package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	config "github.com/louiehdev/gatorcli/internal/config"
	database "github.com/louiehdev/gatorcli/internal/database"
)

func main() {
	Config, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", Config.Url)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	appState := state{cfg: &Config, db: dbQueries}

	commands := commands{commandMap: make(map[string]commandHandler)}
	commands.register("login", commandLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Error: not enough arguments provided")
		os.Exit(1)
	}
	cmd := command{name: args[1], arguments: args[2:]}
	if err := commands.run(&appState, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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
	if err := s.cfg.SetUser(cmd.arguments[0]); err != nil {
		return err
	}
	fmt.Println("User has been set")
	return nil
}
