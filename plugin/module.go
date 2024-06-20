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

// importFn will load a supported module into the Lua runtime.
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
	case "filepath":
		return native.FilepathModule(nm.L)
	case "tabwriter":
		return native.TabWriterModule(nm.L)
	case "opts":
		return OptsModule(nm.L)
	default:
		log.Fatalf("No module %q found", module)
	}

	return nil
}

// createImportFn exposes the `import` function in Lua so that you can
// call `import("something")` to load native modules.
func (nm *nativeModule) createImportFn() {
	nm.L.SetGlobal("import", luar.New(nm.L, nm.importFn))
}
