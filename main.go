package main

import (
	"fmt"
	"log"

	config "github.com/louiehdev/gatorcli/internal/config"
)

func main() {
	Config, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	if err := Config.SetUser("Louie"); err != nil {
		log.Fatal(err)
	}
	updatedConfig, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(updatedConfig)
}
