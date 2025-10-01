package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IronWill79/blog-aggregator/internal/config"
	"github.com/IronWill79/blog-aggregator/internal/database"
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

func main() {
	cfg := config.Read()
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Printf("error opening DB: %s\n", err)
	}
	dbQueries := database.New(db)
	s := state{cfg: &cfg, db: dbQueries}
	cmds := commands{cmds: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
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
