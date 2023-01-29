package plugin

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

type PluginManager struct {
	DB              *gorm.DB
	PluginVFS       *VFS
	plugins         []*Plugin
	passwordPlugins []*Plugin
	commandMap      map[string]CommandFn
	PromptPlugin    PromptFn
}

func (pm *PluginManager) LoadPlugins(path string) error {
	var err error
	pm.plugins, err = LoadPlugins(path, pm.DB)
	pm.passwordPlugins = make([]*Plugin, 0)
	pm.commandMap = make(map[string]CommandFn)

	if err != nil {
		return err
	}

	for _, pl := range pm.plugins {
		err = pl.Init()

		if err != nil {
			return err
		}

		pl.SetVFS(pm.PluginVFS)

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
	}

	if pm.PromptPlugin == nil {
		pm.PromptPlugin = pm.defaultPrompt
	}

	return nil
}

func (pm *PluginManager) defaultPrompt(s *Session) string {
	if s.User.Username == "root" {
		return "# "
	}

	return "$ "
}

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
