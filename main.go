package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IronWill79/blog-aggregator/internal/config"
	"github.com/IronWill79/blog-aggregator/internal/database"
	"github.com/IronWill79/blog-aggregator/internal/rss"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.cmds[cmd.name]
	if !ok {
		return errors.New("no valid command found")
	}
	return f(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return errors.New("username required")
	}
	name := cmd.arguments[0]
	u, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return err
	}
	if err := s.cfg.SetUser(u.Name); err != nil {
		return err
	}
	fmt.Println("Username has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return errors.New("username required")
	}
	name := cmd.arguments[0]
	u, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(u.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User set to %s\n", name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}
	log.Println("Database reset")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if s.cfg.Username == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s", user.Name)
		}
	}
	return nil
}

func handlerAggregate(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Printf("%v", feed)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*rss.RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", "gator")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	xmlFeed, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var feed rss.RSSFeed
	err = xml.Unmarshal(xmlFeed, &feed)
	if err != nil {
		return nil, err
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
	return &feed, nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 2 {
		return errors.New("addfeed requires 2 arguments, name and url")
	}
	name := cmd.arguments[0]
	url := cmd.arguments[1]
	id := user.ID
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Url:       url,
		Name:      name,
		UserID:    id,
	})
	if err != nil {
		return err
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    id,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Feed ID: %s\n", feed.ID)
	fmt.Printf("Created at: %s\n", feed.CreatedAt)
	fmt.Printf("Updated at: %s\n", feed.UpdatedAt)
	fmt.Printf("Name: %s\n", feed.Name)
	fmt.Printf("Url: %s\n", feed.Url)
	fmt.Printf("User ID: %s\n", feed.UserID)
	return nil
}

func handlerPrintFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		user, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("Feed name: %s\n", feed.Name)
		fmt.Printf("Feed URL: %s\n", feed.Url)
		fmt.Printf("Username: %s\n", user.Name)
	}
	return nil
}

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return errors.New("follow requires a single argument, url")
	}
	url := cmd.arguments[0]
	userId := user.ID
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}
	feedId := feed.ID
	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userId,
		FeedID:    feedId,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Feed %s followed by %s\n", follow.FeedName, follow.UserName)
	return nil
}

func handlerPrintFollowedFeeds(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Followed feeds for %s\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("* %s\n", feed.FeedName)
	}
	return nil
}

func (s *state) middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.Username)
	if err != nil {
		return func(s *state, c command) error { return err }
	}
	return func(s *state, c command) error { return handler(s, c, user) }
}

func handlerUnfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return errors.New("unfollow requires one argument, url")
	}
	url := cmd.arguments[0]
	id := user.ID
	err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: id,
		Url:    url,
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	cfg := config.Read()
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Printf("error opening DB: %s\n", err)
	}
	dbQueries := database.New(db)
	s := state{cfg: &cfg, db: dbQueries}
	cmds := commands{cmds: make(map[string]func(*state, command) error)}
	cmds.register("addfeed", s.middlewareLoggedIn(handlerAddFeed))
	cmds.register("agg", handlerAggregate)
	cmds.register("feeds", handlerPrintFeeds)
	cmds.register("follow", s.middlewareLoggedIn(handlerFollowFeed))
	cmds.register("following", s.middlewareLoggedIn(handlerPrintFollowedFeeds))
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("unfollow", s.middlewareLoggedIn(handlerUnfollowFeed))
	cmds.register("users", handlerGetUsers)
	if len(os.Args) < 2 {
		fmt.Println("no command found")
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	err = cmds.run(&s, command{name: cmd, arguments: args})
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
