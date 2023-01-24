package native

import (
	"text/tabwriter"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func TabWriterModule(L *lua.LState) *lua.LTable {
	module := L.NewTable()

	L.SetField(module, "AlignRight", luar.New(L, tabwriter.AlignRight))
	L.SetField(module, "Debug", luar.New(L, tabwriter.Debug))
	L.SetField(module, "DiscardEmptyColumns", luar.New(L, tabwriter.DiscardEmptyColumns))
	L.SetField(module, "Escape", luar.New(L, tabwriter.Escape))
	L.SetField(module, "FilterHTML", luar.New(L, tabwriter.FilterHTML))
	L.SetField(module, "NewWriter", luar.New(L, tabwriter.NewWriter))
	L.SetField(module, "StripEscape", luar.New(L, tabwriter.StripEscape))
	L.SetField(module, "TabIndent", luar.New(L, tabwriter.TabIndent))

	return module
}
