package plugin

import (
	"fmt"
	"io/fs"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type Plugin struct {
	Path   string
	Dir    fs.DirEntry
	Main   fs.DirEntry
	L      *lua.LState
	Config *Config
}

func (p *Plugin) GetPath(withMain bool) string {
	pluginPath := filepath.Join(p.Path, p.Dir.Name())

	if withMain {
		pluginPath = filepath.Join(pluginPath, p.Main.Name())
	}

	return pluginPath
}

func (p *Plugin) Init() error {
	p.L = lua.NewState()

	// Set up the environment here.

	// Run the extension's main file.
	if err := p.L.DoFile(p.GetPath(true)); err != nil {
		panic(err)
	}

	// Finally find and call the install function.
	installFn, ok := p.L.GetGlobal("install").(*lua.LFunction)

	if !ok {
		return fmt.Errorf("the install function wasn't found")
	}

	p.Config = &Config{}
	p.Config.Init()

	err := p.L.CallByParam(lua.P{
		Fn:      installFn,
		NRet:    0,
		Protect: true,
	}, luar.New(p.L, p.Config))

	if err != nil {
		return err
	}

	return nil
}
