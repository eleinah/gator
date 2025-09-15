package cli

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/eleinah/gator/internal/database"
	"github.com/google/uuid"
)

func MiddlewareLoggedIn(Handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		currentUser, err := s.Db.GetUser(context.Background(), s.Cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("couldn't get current user: %w\n", err)
		}

		return Handler(s, cmd, currentUser)
	}
}

func HandlerRegister(s *State, cmd Command) error {
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

	user, err := s.Db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error registering user '%s': %w\n", name, err)
	}

	if err := s.Cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("error setting current user: %w\n", err)
	}

	log.Printf("user '%s' was created with the following:\n", user.Name)
	log.Printf("%+v\n", user)

	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>\n", cmd.Name)
	}

	name := cmd.Args[0]

	_, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("'%s' doesn't exist", name)
	}

	if err := s.Cfg.SetUser(name); err != nil {
		return fmt.Errorf("error setting current user: %w\n", err)
	}

	fmt.Printf("User successfully switched to '%s'!\n", name)

	return nil
}

func HandlerReset(s *State, cmd Command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	if err := s.Db.ResetUsers(context.Background()); err != nil {
		return fmt.Errorf("error resetting table: %w\n", err)
	}

	log.Println("successfully reset table")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users: %w\n", err)
	}

	for _, user := range users {
		if user == s.Cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}

	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("usage: %s <request_wait_time>\n", cmd.Name)
	}

	waitTime, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration given: %w\n", err)
	}

	log.Printf("...collecting feeds every %s...", waitTime)

	ticker := time.NewTicker(waitTime)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *State) {
	feed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Printf("couldn't get feeds to fetch: %w\n", err)
		return
	}

	log.Println("Found feed to fetch!")
	scrapeFeed(s.Db, feed)
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("couldn't mark feed '%s' as fetched: %v\n", feed.Name, err)
		return
	}

	fetchedFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("couldn't fetch feed '%s': %v", feed.Name, err)
		return
	}
	for _, item := range fetchedFeed.Channel.Item {
	publishedAt := sql.NullTime{}
	if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
		publishedAt = sql.NullTime{
			Time:  t,
			Valid: true,
		}
	}

	_, err = db.CreatePost(context.Background(), database.CreatePostParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID:    feed.ID,
		Title:     item.Title,
		Description: sql.NullString{
			String: item.Description,
			Valid:  true,
		},
		Url:         item.Link,
		PublishedAt: publishedAt,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			continue
		}
		log.Printf("Couldn't create post: %v", err)
		continue
	}
	}
	log.Printf("feed '%s' collected, %v posts found", feed.Name, len(fetchedFeed.Channel.Item))
}

func HandlerAddFeed(s *State, cmd Command, currentUser database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <feedName> <feedUrl>\n", cmd.Name)
	}

	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	}

	feed, err := s.Db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}

	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}

	follow, err := s.Db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return fmt.Errorf("couldn't follow feed: %w\n", err)
	}

	fmt.Printf("successfully created feed for '%s'\n", s.Cfg.CurrentUserName)
	fmt.Printf("- name: %s\n", feedName)
	fmt.Printf("- link: %s\n\n", feedUrl)
	fmt.Println("successfully followed feed:")
	fmt.Printf("- user: %s\n", follow.UserName)
	fmt.Printf("- feed: %s\n", follow.FeedName)

	return nil

}

func HandlerFeeds(s *State, cmd Command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	feeds, err := s.Db.GetFeeds(context.Background())
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

func HandlerFollow(s *State, cmd Command, currentUser database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>\n", cmd.Name)
	}

	url := cmd.Args[0]

	feed, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to get feed by url: %w\n", err)
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}

	followRow, err := s.Db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w\n", err)
	}

	fmt.Println("feed follow created:")
	fmt.Printf("- user: %s\n", followRow.UserName)
	fmt.Printf("- name: %s\n", followRow.FeedName)

	return nil
}

func HandlerFollowing(s *State, cmd Command, currentUser database.User) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s\n", cmd.Name)
	}

	following, err := s.Db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
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

func HandlerUnfollow(s *State, cmd Command, currentUser database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>\n", cmd.Name)
	}

	url := cmd.Args[0]

	feed, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to get feed by url: %w\n", err)
	}

	params := database.DeleteFeedFollowParams{
		UserID: currentUser.ID,
		FeedID: feed.ID,
	}

	if err := s.Db.DeleteFeedFollow(context.Background(), params); err != nil {
		return fmt.Errorf("failed to unfollow feed: %w\n", err)
	}

	fmt.Printf("successfully unfollowed feed for '%s':\n", currentUser)
	fmt.Printf("- name: %s\n", feed.Name)
	fmt.Printf("- id: %s\n", feed.ID)
	fmt.Printf("- url: %s\n", feed.Url)
	return nil
}

func HandlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.Args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.Db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("found %d posts for user '%s':\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
