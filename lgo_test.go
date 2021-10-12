package lgo

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

func TestAll(t *testing.T) {
	lua := New()
	lua.PrintTraceback = false

	t.Run("simple function", func(t *testing.T) {
		lua.RegisterFunction("foo", func() {})
		lua.RunString(`foo()`)
	})

	t.Run("bool argument", func(t *testing.T) {
		lua.RegisterFunction("bool", func(b bool) {
			if b != true {
				t.Fatal()
			}
		})
		lua.RunString(`bool(true)`)
	})

	t.Run("int argument", func(t *testing.T) {
		lua.RegisterFunction("int", func(i int) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.RunString(`int(42)`)
	})

	t.Run("uint argument", func(t *testing.T) {
		lua.RegisterFunction("uint", func(i uint) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.RunString(`uint(42)`)
	})

	t.Run("float argument", func(t *testing.T) {
		lua.RegisterFunction("float", func(f float64) {
			if f != 42.5 {
				t.Fatal()
			}
		})
		lua.RunString(`float(42.5)`)
	})

	t.Run("interface argument", func(t *testing.T) {
		lua.RegisterFunction("interface", func(a, b interface{}) {
			if i, ok := a.(float64); !ok || i != 42 {
				t.Fatal()
			}
			if s, ok := b.(string); !ok || s != "foo" {
				t.Fatal()
			}
		})
		lua.RunString(`interface(42, 'foo')`)
	})

	t.Run("string argument", func(t *testing.T) {
		lua.RegisterFunction("str", func(s string) {
			if s != "foo" {
				t.Fatal()
			}
		})
		lua.RunString(`str('foo')`)
	})

	t.Run("bytes argument", func(t *testing.T) {
		lua.RegisterFunction("bytes", func(bs []byte) {
			if !bytes.Equal(bs, []byte("foo")) {
				t.Fatal()
			}
		})
		lua.RunString(`bytes('foo')`)
	})

	t.Run("slice argument", func(t *testing.T) {
		lua.RegisterFunction("slice", func(is []int) {
			if len(is) != 2 || is[0] != 4 || is[1] != 2 {
				t.Fatal()
			}
		})
		lua.RunString(`slice{4, 2}`)
	})

	t.Run("map argument", func(t *testing.T) {
		lua.RegisterFunction("map", func(m map[string]int) {
			if len(m) != 2 || m["foo"] != 4 || m["bar"] != 2 {
				t.Fatal()
			}
		})
		lua.RunString(`map{foo = 4, bar = 2}`)
	})

	t.Run("unsafepointer argument", func(t *testing.T) {
		lua.RegisterFunction("unsafepointer", func(m unsafe.Pointer) {
			if m != nil {
				t.Fatal()
			}
		})
		lua.RunString(`unsafepointer(nil)`)
	})

	t.Run("unknown argument", func(t *testing.T) {
		lua.RegisterFunction("foo", func(s struct{}) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "unknown argument type struct {}" {
				t.Fatal()
			}
		}()
		lua.RunString(`foo(1)`)
	})

	t.Run("bool return", func(t *testing.T) {
		lua.RegisterFunction("bool", func() bool {
			return true
		})
		lua.RunString(`if bool() ~= true then error('not true') end`)
	})

	t.Run("string return", func(t *testing.T) {
		lua.RegisterFunction("str", func() string {
			return "foo"
		})
		lua.RunString(`if str() ~= 'foo' then error('not string') end`)
	})

	t.Run("numeric return", func(t *testing.T) {
		lua.RegisterFunction("num", func() (int32, uint64, float32) {
			return 42, 99, 33.3
		})
		lua.RunString(`
			i, u, f = num()
			if i ~= 42 then error('not int32') end
			if u ~= 99 then error('not uint64') end
			if f - 33.3 > 0.000001 then error('not float32') end
		`)
	})

	t.Run("slice return", func(t *testing.T) {
		lua.RegisterFunction("slice", func() []int {
			return []int{1, 2, 3}
		})
		lua.RunString(`
			ret = slice()
			if #ret ~= 3 then error('slice error') end
			if ret[1] ~= 1 or ret[2] ~= 2 or ret[3] ~= 3 then error('slice error') end
		`)
	})

	t.Run("interface return", func(t *testing.T) {
		lua.RegisterFunction("interface", func() interface{} {
			return "foo"
		})
		lua.RunString(`if interface() ~= 'foo' then error('interface error') end`)
	})

	t.Run("register functions in map", func(t *testing.T) {
		lua.RegisterFunctions(map[string]interface{}{
			"rewqfdsafdsaf4312": func() {},
			"jl45321dfsafsda":   func() {},
		})
		lua.RunString(`
			rewqfdsafdsaf4312()
			jl45321dfsafsda()
		`)
	})

	t.Run("register variadic function", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if !strings.HasPrefix(p.(string), "cannot register variadic function") {
				t.Fatal()
			}
		}()
		lua.RegisterFunction("foo", func(is ...int) {})
	})

	t.Run("panic", func(t *testing.T) {
		lua.RegisterFunction("panic", func() {
			panic("foo")
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "foo" {
				t.Fatal()
			}
		}()
		lua.RunString(`panic()`)
	})

	t.Run("namespace", func(t *testing.T) {
		lua.RegisterFunction("Foo.bar", func(i int) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.RunString(`Foo.bar(42)`)
	})

	t.Run("namespace2", func(t *testing.T) {
		lua.RegisterFunction("bar.bar.bar", func(i int) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.RunString(`bar.bar.bar(42)`)
	})

	t.Run("namespace3", func(t *testing.T) {
		lua.RegisterFunction("bar.foo.baz.quux", func(i int) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.RunString(`bar.foo.baz.quux(42)`)
	})

	t.Run("invalid global namespace", func(t *testing.T) {
		lua.RunString(`Foo = 1`)
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "global Foo is not a table" {
				t.Fatal()
			}
		}()
		lua.RegisterFunction("Foo.bar", func() {})
	})

	t.Run("invalid namespace", func(t *testing.T) {
		lua.RunString(`Bar = {}`)
		lua.RunString(`Bar.Bar = 1`)
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "namespace Bar is not a table" {
				t.Fatal()
			}
		}()
		lua.RegisterFunction("Bar.Bar.bar", func() {})
	})

	t.Run("call function", func(t *testing.T) {
		lua.RunString(`
			function foo(i, b1, b2)
				if i ~= 42 then error('not 42') end
				if b1 ~= true then error('not true') end
				if b2 ~= false then error('not false') end
			end
		`)
		lua.CallFunction("foo", 42, true, false)
	})

	t.Run("argument not match", func(t *testing.T) {
		lua.RegisterFunction("foo", func(a, b int) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if !strings.HasPrefix(p.(string), "arguments not match") {
				t.Fatal()
			}
		}()
		lua.RunString(`foo(1)`)
	})

	t.Run("invalid bool", func(t *testing.T) {
		lua.RegisterFunction("foo", func(b bool) {
			if !b {
				t.Fatal()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "not a boolean" {
				t.Fatal()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(true)`)
	})

	t.Run("invalid int", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i int) {
			if i != 42 {
				t.Fatal()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "not an integer" {
				t.Fatal()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(42)`)
	})

	t.Run("invalid unsigned", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i uint) {
			if i != 42 {
				t.Fatal()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "not an unsigned" {
				t.Fatal()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(42)`)
	})

	t.Run("invalid float", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i float64) {
			if i != 4.2 {
				t.Fatal()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "not a float" {
				t.Fatal()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(4.2)`)
	})

	t.Run("bool interface", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i interface{}) {
			if i, ok := i.(bool); !ok {
				t.Fatal()
			} else {
				if !i {
					t.Fatal()
				}
			}
		})
		lua.RunString(`foo(true)`)
	})

	t.Run("wrong interface argument", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i interface{}) {
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if !strings.HasPrefix(p.(string), "wrong interface argument") {
				t.Fatal()
			}
		}()
		lua.RunString(`foo(function() end)`)
	})

	t.Run("invalid string", func(t *testing.T) {
		lua.RegisterFunction("invalidstring", func(s string) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "not a string" {
				t.Fatal()
			}
		}()
		lua.RunString("invalidstring(1)")
	})

	t.Run("invalid slice", func(t *testing.T) {
		lua.RegisterFunction("invalidslice", func(slice []int) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "wrong slice argument" {
				t.Fatal()
			}
		}()
		lua.RunString("invalidslice(1)")
	})

	t.Run("invalid pointer", func(t *testing.T) {
		lua.RegisterFunction("invalidpointer", func(p *int) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "not a pointer" {
				t.Fatal()
			}
		}()
		lua.RunString(`invalidpointer(1)`)
	})

	t.Run("invalid map", func(t *testing.T) {
		lua.RegisterFunction("invalidmap", func(m map[int]bool) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if p.(string) != "not a map" {
				t.Fatal()
			}
		}()
		lua.RunString(`invalidmap(1)`)
	})

	t.Run("wrong return type", func(t *testing.T) {
		lua.RunString(`
			function foo(fn)
				return fn
			end
		`)
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			msg := fmt.Sprintf("%s", p)
			if !strings.HasPrefix(msg, "invalid value") {
				t.Fatalf("got %s", msg)
			}
		}()
		lua.CallFunction("foo", func() {})
	})

	t.Run("invalid lua code", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if !strings.Contains(p.(string), "syntax error") {
				t.Fatal()
			}
		}()
		lua.RunString(`func end`)
	})

	t.Run("lua error", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if !strings.Contains(p.(string), "foobarbaz") {
				t.Fatal()
			}
		}()
		lua.RunString(`error('foobarbaz')`)
	})

	t.Run("call lua error", func(t *testing.T) {
		lua.RunString(`
			function foobarbaz()
				error('foobarbaz')
			end
		`)
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			if !strings.Contains(p.(string), "foobarbaz") {
				t.Fatal()
			}
		}()
		lua.CallFunction("foobarbaz")
	})
}
