package main

import (
	"errors"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	validCommands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.validCommands[cmd.Name]
	if !ok {
		return errors.New("command does not exist")
	}

	return f(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.validCommands[name] = f
}
