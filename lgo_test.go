package lgo

import (
	"testing"
)

func TestLgo(t *testing.T) {
	lua := NewLua()
	lua.RegisterFunction("foo", func() {})
	lua.RunString(`
	foo()
	`)
}
