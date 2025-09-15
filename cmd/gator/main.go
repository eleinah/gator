package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/eleinah/gator/internal/config"
	"github.com/eleinah/gator/internal/cli"
	"github.com/eleinah/gator/internal/database"
	_ "github.com/lib/pq"
)

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

	appState := cli.State{Cfg: &cfg, Db: dbQueries}

	cmds := cli.Commands{
		ValidCommands: make(map[string]func(*cli.State, cli.Command) error),
	}

	registerCmds(&cmds)

	if len(os.Args) < 2 {
		log.Fatal("usage: gator <Command> [args...]")
	}

	CommandName := os.Args[1]
	CommandArgs := os.Args[2:]

	if err := cmds.Run(&appState, cli.Command{Name: CommandName, Args: CommandArgs}); err != nil {
		log.Fatal(err)
	}
}

func registerCmds(cmds *cli.Commands) {
	cmds.Register("login", cli.HandlerLogin)
	cmds.Register("register", cli.HandlerRegister)
	cmds.Register("reset", cli.HandlerReset)
	cmds.Register("users", cli.HandlerUsers)
	cmds.Register("agg", cli.HandlerAgg)
	cmds.Register("addfeed", cli.MiddlewareLoggedIn(cli.HandlerAddFeed))
	cmds.Register("feeds", cli.HandlerFeeds)
	cmds.Register("follow", cli.MiddlewareLoggedIn(cli.HandlerFollow))
	cmds.Register("following", cli.MiddlewareLoggedIn(cli.HandlerFollowing))
	cmds.Register("unfollow", cli.MiddlewareLoggedIn(cli.HandlerUnfollow))
	cmds.Register("browse", cli.MiddlewareLoggedIn(cli.HandlerBrowse))
}
