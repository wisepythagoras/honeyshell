package plugin

import "gorm.io/gorm"

type PluginManager struct {
	plugins         []*Plugin
	passwordPlugins []*Plugin
	DB              *gorm.DB
	commandMap      map[string]CommandFn
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

		if pl.HasPasswordIntercept() {
			pm.passwordPlugins = append(pm.passwordPlugins, pl)
		}

		if pl.HasCommandDefined() {
			for cmd, commandFn := range pl.Commands() {
				pm.commandMap[cmd] = commandFn
			}
		}
	}

	return nil
}

func (pm *PluginManager) GetCommand(cmd string) (CommandFn, bool) {
	if cmd, ok := pm.commandMap[cmd]; ok {
		return cmd, ok
	}

	return nil, false
}

func (pm *PluginManager) GetPasswordIntercepts() []*Plugin {
	return pm.passwordPlugins
}
