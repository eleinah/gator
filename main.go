package main

import (
	"fmt"
	"log"

	"github.com/eleinah/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}
	fmt.Printf("Read config: %+v\n", cfg)

	if err := cfg.SetUser("ellie"); err != nil {
		log.Fatalf("Error setting user: %v", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("Error reading configuration file: %v\n", err)
	}
	fmt.Printf("Read config again: %+v\n", cfg)
}
