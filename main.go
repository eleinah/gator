package main

import (
	"log"
	"os"

	"github.com/eleinah/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	appState := state{cfg: &cfg}

	cmds := commands{
		validCommands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)

	if len(os.Args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	commandName := os.Args[1]
	commandArgs := os.Args[2:]

	if err := cmds.run(&appState, command{Name: commandName, Args: commandArgs}); err != nil {
		log.Fatal(err)
	}
}
