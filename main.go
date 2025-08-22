package main

import (
	"fmt"

	"github.com/eleinah/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
	}

	if err := cfg.SetUser("ellie"); err != nil {
		fmt.Printf("Error setting user: %v\n", err)
	}

	cfg, err = config.Read()
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
	}

	fmt.Printf("db_url: %v\n", cfg.DbUrl)
	fmt.Printf("current_user_name: %v\n", cfg.CurrentUserName)
}
