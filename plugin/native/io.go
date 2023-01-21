package native

import (
	"io"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func IoModule(L *lua.LState) *lua.LTable {
	module := L.NewTable()

	L.SetField(module, "ReadAll", luar.New(L, io.ReadAll))
	L.SetField(module, "ReadAtLeast", luar.New(L, io.ReadAtLeast))
	L.SetField(module, "ReadFull", luar.New(L, io.ReadFull))
	L.SetField(module, "WriteString", luar.New(L, io.WriteString))
	L.SetField(module, "MultiWriter", luar.New(L, io.MultiWriter))

	return module
}
