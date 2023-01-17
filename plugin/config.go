package plugin

import (
	"log"
	"net"
)

type CommandCallbackFn func(...string)

type CommandFn func([]string, CommandCallbackFn)
type PasswordInterceptFn func(string, string, *net.IP) bool

type Config struct {
	CommandCallbacks    map[string]CommandFn
	PasswordInterceptor PasswordInterceptFn
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

func (c *Config) RegisterPasswordIntercept(interceptor PasswordInterceptFn) {
	c.PasswordInterceptor = interceptor
}
