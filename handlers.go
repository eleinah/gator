package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eleinah/gator/internal/database"
	"github.com/google/uuid"
)

func handlerRegister (s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>\n", cmd.Name)
	}

	name := cmd.Args[0]

	params := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
	}

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error registering user '%s': %w\n", name, err)
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("error setting current user: %w\n", err)
	}

	log.Printf("user '%s' was created with the following:\n", user.Name)
	log.Printf("%+v\n", user)

	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>\n", cmd.Name)
	}

	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("'%s' doesn't exist", name)
	}

	if err := s.cfg.SetUser(name); err != nil {
		return fmt.Errorf("error setting current user: %w\n", err)
	}

	fmt.Printf("User successfully switched to '%s'!\n", name)

	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	if err := s.db.ResetUsers(context.Background()); err != nil {
		return fmt.Errorf("error resetting table: %w\n", err)
	}

	log.Println("successfully reset table")
	return nil
}
