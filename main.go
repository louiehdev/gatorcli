package main

import (
	"fmt"
	"log"
	"os"

	config "github.com/louiehdev/gatorcli/internal/config"
)

func main() {
	Config, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	appState := state{configp: &Config}
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
	}
}

type state struct {
	configp *config.Config
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
	if err := s.configp.SetUser(cmd.arguments[0]); err != nil {
		return err
	}
	fmt.Println("User has been set")
	return nil
}
