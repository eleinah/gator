package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eleinah/gator/internal/database"
	"github.com/google/uuid"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>\n", cmd.Name)
	}

	name := cmd.Args[0]

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
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

func handlerUsers(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users: %w\n", err)
	}

	for _, user := range users {
		if user == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error fetching feed: %w\n", err)
	}

	channel := feed.Channel
	items := channel.Item

	fmt.Printf("Channel Title: %s\n", channel.Title)
	fmt.Printf("Channel Description: %s\n", channel.Description)
	fmt.Println(`------------
   Items
------------`)
	for _, item := range items {
		fmt.Printf("Title: %s\n", item.Title)
		fmt.Printf("Date: %s    Link: %s\n", item.PubDate, item.Link)
		fmt.Printf("==>\n%s\n<==\n\n", item.Description)
	}
	fmt.Println(`------------
    End
------------`)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <feedName> <feedUrl>\n", cmd.Name)
	}

	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]

	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	}

	_, err = s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}

	fmt.Printf("successfully created feed for '%s'\n", s.cfg.CurrentUserName)
	fmt.Printf("- name: %s\n", feedName)
	fmt.Printf("- link: %s\n", feedUrl)

	return nil

}
