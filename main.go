package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/IronWill79/blog-aggregator/internal/config"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
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
	if err := s.cfg.SetUser(cmd.arguments[0]); err != nil {
		return err
	}
	fmt.Println("Username has been set")
	return nil
}

func main() {
	cfg := config.Read()
	s := state{cfg: &cfg}
	cmds := commands{cmds: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	if len(os.Args) < 2 {
		fmt.Println("no command found")
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	err := cmds.run(&s, command{name: cmd, arguments: args})
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
