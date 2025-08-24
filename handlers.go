package main

import (
	"fmt"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>\n", cmd.Name)
	}

	name := cmd.Args[0]

	if err := s.cfg.SetUser(name); err != nil {
		return fmt.Errorf("error setting current user: %w\n", err)
	}

	fmt.Printf("User successfully switched to '%s'!\n", name)

	return nil
}
