package plugin

import (
	"log"
	"math/rand"
	"net"
	"path/filepath"
	"time"
)

type TermWriteFn func(...string)

type CommandFn func(*CmdArgs, *Session)
type PasswordInterceptFn func(string, string, *net.IP) bool
type PromptFn func(*Session) string
type LoginMessageFn func(*Session) string

// Config is a struct that handles everything related to the sandbox. For
// example, the registered commands, the password interceptor, the fake
// prompt, and other things.
type Config struct {
	CommandCallbacks    map[string]CommandFn
	PasswordInterceptor PasswordInterceptFn
	PromptFn            PromptFn
	LoginMessageFn      LoginMessageFn
	vfs                 *VFS
}

// Init initializes the instance. This must run, as it instanciates the
// command callbacks.
func (c *Config) Init() {
	c.CommandCallbacks = make(map[string]CommandFn)
}

// RegisterCommand adds a command to the list of supported commands. This
// means that an attacker can run the command by the supplied `cmd` and then
// that will run the command function (`cmdFn`).
func (c *Config) RegisterCommand(cmd, dir string, cmdFn CommandFn) bool {
	if cmdFn == nil {
		log.Println("No command function for command", cmd)
		return false
	}

	log.Println("Registering command", cmd)
	c.CommandCallbacks[cmd] = cmdFn

	if c.vfs != nil {
		_, file, err := c.vfs.FindFile(dir)

		if err != nil {
			log.Println("Error:", err)
			return false
		}

		if file.Type == T_SYMLINK {
			_, file, err = c.vfs.FindFile(filepath.Join("/", file.LinkTo))

			if err != nil {
				log.Println("Error:", err)
				return false
			}
		}

		if _, ok := file.Files[cmd]; ok {
			log.Printf("Error: command \"%s%s\" already exists\n", dir, cmd)
			return false
		}

		file.Files[cmd] = VFSFile{
			Type:    T_FILE,
			Mode:    0755,
			Name:    cmd,
			CmdFn:   cmdFn,
			Owner:   "root",
			Group:   "root",
			ModTime: time.Now().Add(time.Duration(-rand.Intn(30)) * time.Hour),
		}
		c.CommandCallbacks[filepath.Join(dir, cmd)] = cmdFn
	}

	return true
}

func (c *Config) RegisterLoginMessage(loginMsgFn LoginMessageFn) {
	c.LoginMessageFn = loginMsgFn
}

func (c *Config) RegisterPasswordIntercept(interceptor PasswordInterceptFn) {
	c.PasswordInterceptor = interceptor
}

func (c *Config) RegisterPrompt(promptFn PromptFn) {
	c.PromptFn = promptFn
}
