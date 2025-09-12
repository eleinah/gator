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
		fmt.Printf("==>\n%s\n<==\n", item.Description)
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

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}

	followParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: currentUser.ID,
		FeedID: feed.ID,
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return fmt.Errorf("couldn't follow feed: %w\n", err)
	}

	fmt.Printf("successfully created feed for '%s'\n", s.cfg.CurrentUserName)
	fmt.Printf("- name: %s\n", feedName)
	fmt.Printf("- link: %s\n\n", feedUrl)
	fmt.Println("successfully followed feed:")
	fmt.Printf("- user: %s\n", follow.UserName)
	fmt.Printf("- feed: %s\n", follow.FeedName)

	return nil

}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	fmt.Println(`------------
   Feeds
------------`)

	for _, feed := range feeds {
		fmt.Printf("\n- Name: '%s'\n", feed.Feedname)
		fmt.Printf("- URL: '%s'\n", feed.Url)
		fmt.Printf("- Created by: '%s'\n\n", feed.Createdby)
	}

	fmt.Println(`------------
    End
------------`)

	return nil

}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>\n", cmd.Name)
	}

	url := cmd.Args[0]

	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	feed, err := s.db.GetFeedByURL(context.Background(), url)

	params := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: currentUser.ID,
		FeedID: feed.ID,
	}

	followRow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w\n", err)
	}

	fmt.Println("feed follow created:")
	fmt.Printf("- user: %s\n", followRow.UserName)
	fmt.Printf("- name: %s\n", followRow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	following, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("failed to get followed feeds for user: %w\n", err)
	}

	if len(following) == 0 {
		fmt.Println("user is not following any feeds")
		return nil
	}

	fmt.Printf("'%s' is following:\n", currentUser.Name)

	for _, feed := range following {
		fmt.Printf("- %s\n", feed.FeedName)
	}

	return nil
}
