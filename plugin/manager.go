package plugin

import "gorm.io/gorm"

type PluginManager struct {
	plugins         []*Plugin
	passwordPlugins []*Plugin
	DB              *gorm.DB
}

func (pm *PluginManager) LoadPlugins(path string) error {
	var err error
	pm.plugins, err = LoadPlugins(path, pm.DB)
	pm.passwordPlugins = make([]*Plugin, 0)

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
	}

	return nil
}

func (pm *PluginManager) GetPasswordIntercepts() []*Plugin {
	return pm.passwordPlugins
}
