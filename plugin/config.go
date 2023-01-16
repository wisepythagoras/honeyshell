package plugin

import "log"

type CommandCallbackFn func(...string)

type CommandFn func([]string, CommandCallbackFn)

type Config struct {
	CommandCallbacks map[string]CommandFn
}

func (c *Config) Init() {
	c.CommandCallbacks = make(map[string]CommandFn)
}

func (c *Config) RegisterCommand(cmd string, cmdFn CommandFn) bool {
	if cmdFn == nil {
		log.Println("No command function for command", cmd)
		return false
	}

	log.Println("Registering command", cmd)
	c.CommandCallbacks[cmd] = cmdFn

	return true
}
