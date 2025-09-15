package cli

import (
	"errors"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	ValidCommands map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	f, ok := c.ValidCommands[cmd.Name]
	if !ok {
		return errors.New("Command does not exist")
	}

	return f(s, cmd)
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.ValidCommands[name] = f
}
