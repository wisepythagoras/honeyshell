package plugin

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

// PluginManager handles all things related to the plugins. Use this
// instance to load plugins and get commands.
type PluginManager struct {
	DB              *gorm.DB
	PluginVFS       *VFS
	plugins         []*Plugin
	passwordPlugins []*Plugin
	commandMap      map[string]CommandFn
	PromptPlugin    PromptFn
	LoginMessageFn  LoginMessageFn
}

// LoadPlugins loads the plugin by supplying a `path`.
func (pm *PluginManager) LoadPlugins(path string) error {
	var err error
	pm.plugins, err = LoadPlugins(path, pm.DB)
	pm.passwordPlugins = make([]*Plugin, 0)
	pm.commandMap = make(map[string]CommandFn)

	if err != nil {
		return err
	}

	for _, pl := range pm.plugins {
		err = pl.Init(pm.PluginVFS)

		if err != nil {
			return err
		}

		if pl.HasPasswordIntercept() {
			pm.passwordPlugins = append(pm.passwordPlugins, pl)
		}

		if pl.HasCommandDefined() {
			for cmd, commandFn := range pl.Commands() {
				pm.commandMap[cmd] = commandFn
			}
		}

		if pl.HasPromptFn() {
			pm.PromptPlugin = pl.Config.PromptFn
		}

		if pl.HasLoginMessage() {
			pm.LoginMessageFn = pl.Config.LoginMessageFn
		}
	}

	if pm.PromptPlugin == nil {
		pm.PromptPlugin = pm.defaultPrompt
	}

	return nil
}

// defaultPrompt displays a very basic bash-like prompt.
func (pm *PluginManager) defaultPrompt(s *Session) string {
	if s.User.Username == "root" {
		return "# "
	}

	return "$ "
}

// GetComand returns a function handler and a boolean (if it was found or not)
// for a command (as a string).
func (pm *PluginManager) GetCommand(cmd string) (CommandFn, bool) {
	if cmd, ok := pm.commandMap[cmd]; ok {
		return cmd, ok
	}

	return nil, false
}

func (pm *PluginManager) MatchCommand(part string) ([]CommandFn, []string) {
	commands := make([]string, 0)
	cmdFns := make([]CommandFn, 0)
	reg := regexp.MustCompile(fmt.Sprintf("^%s", part))

	for cmd, cmdFn := range pm.commandMap {
		if reg.MatchString(cmd) {
			commands = append(commands, cmd)
			cmdFns = append(cmdFns, cmdFn)
		}
	}

	return cmdFns, commands
}

func (pm *PluginManager) GetPasswordIntercepts() []*Plugin {
	return pm.passwordPlugins
}
