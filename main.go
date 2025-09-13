package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/eleinah/gator/internal/config"
	"github.com/eleinah/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	dbQueries := database.New(db)

	appState := state{cfg: &cfg, db: dbQueries}

	cmds := commands{
		validCommands: make(map[string]func(*state, command) error),
	}

	registerCmds(&cmds)

	if len(os.Args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	commandName := os.Args[1]
	commandArgs := os.Args[2:]

	if err := cmds.run(&appState, command{Name: commandName, Args: commandArgs}); err != nil {
		log.Fatal(err)
	}
}

func registerCmds(cmds *commands) {
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
}
