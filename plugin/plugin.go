package plugin

import (
	"fmt"
	"io/fs"
	"net"
	"path/filepath"

	"github.com/wisepythagoras/honeyshell/plugin/native"
	lua "github.com/yuin/gopher-lua"
	"gorm.io/gorm"
	luar "layeh.com/gopher-luar"
)

type Plugin struct {
	Path   string
	Dir    fs.DirEntry
	Main   fs.DirEntry
	L      *lua.LState
	Config *Config
	DB     *gorm.DB
	vfs    *VFS
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
	nativeMod := nativeModule{L: p.L}
	nativeMod.createImportFn()

	p.L.SetGlobal("db", luar.New(p.L, native.DBModule(p.L, p.DB)))
	p.L.SetGlobal("dirname", luar.New(p.L, p.GetPath(false)))
	p.L.SetGlobal("toBytes", luar.New(p.L, stringToBytes))
	p.L.SetGlobal("toString", luar.New(p.L, bytesToString))

	// Allow requiring lua files from the plugin's directory.
	pkg := p.L.GetGlobal("package")
	newPath := fmt.Sprintf("%s/?.lua;%s", p.GetPath(false), pkg.String())
	p.L.SetField(pkg, "path", luar.New(p.L, newPath))

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

func (p *Plugin) SetVFS(vfs *VFS) {
	p.vfs = vfs
}

func (p *Plugin) HasPasswordIntercept() bool {
	return p.Config.PasswordInterceptor != nil
}

func (p *Plugin) HasCommandDefined() bool {
	return len(p.Config.CommandCallbacks) > 0
}

func (p *Plugin) HasPromptFn() bool {
	return p.Config.PromptFn != nil
}

func (p *Plugin) Commands() map[string]CommandFn {
	return p.Config.CommandCallbacks
}

func (p *Plugin) CallPasswordInterceptor(username, password string, ip *net.IP) bool {
	if !p.HasPasswordIntercept() {
		return false
	}

	return p.Config.PasswordInterceptor(username, password, ip)
}
