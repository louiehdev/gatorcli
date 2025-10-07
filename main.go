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
	commands.register("register", commandRegister)
	commands.register("users", commandUsers)
	commands.register("reset", commandReset)
	commands.register("agg", commandAgg)
	commands.register("addfeed", middlewareLoggedIn(commandAddFeed))
	commands.register("feeds", commandFeeds)
	commands.register("follow", middlewareLoggedIn(commandFollow))
	commands.register("following", middlewareLoggedIn(commandFollowing))
	commands.register("unfollow", middlewareLoggedIn(commandUnfollow))
	commands.register("browse", middlewareLoggedIn(commandBrowse))
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
