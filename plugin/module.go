package plugin

import (
	"log"

	"github.com/wisepythagoras/honeyshell/plugin/native"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type nativeModule struct {
	L *lua.LState
}

func (nm *nativeModule) importFn(module string) *lua.LTable {
	switch module {
	case "fmt":
		return native.FmtModule(nm.L)
	case "time":
		return native.TimeModule(nm.L)
	case "io":
		return native.IoModule(nm.L)
	case "os":
		return native.OsModule(nm.L)
	case "json":
		return native.JsonModule(nm.L)
	case "strings":
		return native.StringsModule(nm.L)
	default:
		log.Fatalf("No module %q found", module)
	}

	return nil
}

func (nm *nativeModule) createImportFn() {
	nm.L.SetGlobal("import", luar.New(nm.L, nm.importFn))
}
