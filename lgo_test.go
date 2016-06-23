package lgo

import (
	"bytes"
	"strings"
	"testing"
	"unsafe"
)

func TestAll(t *testing.T) {
	lua := NewLua()
	lua.PrintTraceback = false

	t.Run("simple function", func(t *testing.T) {
		lua.RegisterFunction("foo", func() {})
		lua.RunString(`foo()`)
	})

	t.Run("bool argument", func(t *testing.T) {
		lua.RegisterFunction("bool", func(b bool) {
			if b != true {
				t.Fail()
			}
		})
		lua.RunString(`bool(true)`)
	})

	t.Run("int argument", func(t *testing.T) {
		lua.RegisterFunction("int", func(i int) {
			if i != 42 {
				t.Fail()
			}
		})
		lua.RunString(`int(42)`)
	})

	t.Run("uint argument", func(t *testing.T) {
		lua.RegisterFunction("uint", func(i uint) {
			if i != 42 {
				t.Fail()
			}
		})
		lua.RunString(`uint(42)`)
	})

	t.Run("float argument", func(t *testing.T) {
		lua.RegisterFunction("float", func(f float64) {
			if f != 42.5 {
				t.Fail()
			}
		})
		lua.RunString(`float(42.5)`)
	})

	t.Run("interface argument", func(t *testing.T) {
		lua.RegisterFunction("interface", func(a, b interface{}) {
			if i, ok := a.(float64); !ok || i != 42 {
				t.Fail()
			}
			if s, ok := b.(string); !ok || s != "foo" {
				t.Fail()
			}
		})
		lua.RunString(`interface(42, 'foo')`)
	})

	t.Run("string argument", func(t *testing.T) {
		lua.RegisterFunction("str", func(s string) {
			if s != "foo" {
				t.Fail()
			}
		})
		lua.RunString(`str('foo')`)
	})

	t.Run("bytes argument", func(t *testing.T) {
		lua.RegisterFunction("bytes", func(bs []byte) {
			if !bytes.Equal(bs, []byte("foo")) {
				t.Fail()
			}
		})
		lua.RunString(`bytes('foo')`)
	})

	t.Run("slice argument", func(t *testing.T) {
		lua.RegisterFunction("slice", func(is []int) {
			if len(is) != 2 || is[0] != 4 || is[1] != 2 {
				t.Fail()
			}
		})
		lua.RunString(`slice{4, 2}`)
	})

	t.Run("pointer argument", func(t *testing.T) {
		s := "foo"
		expected := unsafe.Pointer(&s)
		lua.RegisterFunction("get_a_pointer", func() unsafe.Pointer {
			return expected
		})
		lua.RegisterFunction("pointer", func(p unsafe.Pointer) {
			if p != expected {
				t.Fail()
			}
		})
		lua.RunString(`pointer(get_a_pointer())`)
	})

	t.Run("map argument", func(t *testing.T) {
		lua.RegisterFunction("map", func(m map[string]int) {
			if len(m) != 2 || m["foo"] != 4 || m["bar"] != 2 {
				t.Fail()
			}
		})
		lua.RunString(`map{foo = 4, bar = 2}`)
	})

	t.Run("unsafepointer argument", func(t *testing.T) {
		lua.RegisterFunction("unsafepointer", func(m unsafe.Pointer) {
			if m != nil {
				t.Fail()
			}
		})
		lua.RunString(`unsafepointer(nil)`)
	})

	t.Run("unknown argument", func(t *testing.T) {
		lua.RegisterFunction("foo", func(s struct{}) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if p.(string) != "unknown argument type struct {}" {
				t.Fail()
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

	t.Run("pointer return", func(t *testing.T) {
		lua.RegisterFunction("ptr", func() *int {
			i := 5
			return &i
		})
		lua.RunString(`if type(ptr()) ~= 'userdata' then error('not userdata') end`)
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
				t.Fail()
			}
			if !strings.HasPrefix(p.(string), "cannot register variadic function") {
				t.Fail()
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
				t.Fail()
			}
			if p.(string) != "foo" {
				t.Fail()
			}
		}()
		lua.RunString(`panic()`)
	})

	t.Run("namespace", func(t *testing.T) {
		lua.RegisterFunction("Foo.bar", func(i int) {
			if i != 42 {
				t.Fail()
			}
		})
		lua.RunString(`Foo.bar(42)`)
	})

	t.Run("namespace2", func(t *testing.T) {
		lua.RegisterFunction("bar.bar.bar", func(i int) {
			if i != 42 {
				t.Fail()
			}
		})
		lua.RunString(`bar.bar.bar(42)`)
	})

	t.Run("namespace3", func(t *testing.T) {
		lua.RegisterFunction("bar.foo.baz.quux", func(i int) {
			if i != 42 {
				t.Fail()
			}
		})
		lua.RunString(`bar.foo.baz.quux(42)`)
	})

	t.Run("invalid global namespace", func(t *testing.T) {
		lua.RunString(`Foo = 1`)
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if p.(string) != "global Foo is not a table" {
				t.Fail()
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
				t.Fail()
			}
			if p.(string) != "namespace Bar is not a table" {
				t.Fail()
			}
		}()
		lua.RegisterFunction("Bar.Bar.bar", func() {})
	})

	t.Run("call function", func(t *testing.T) {
		lua.RunString(`
			function foo(arg)
				if arg ~= 42 then error('not 42') end
			end
		`)
		lua.CallFunction("foo", 42)
	})

	t.Run("argument not match", func(t *testing.T) {
		lua.RegisterFunction("foo", func(a, b int) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if !strings.HasPrefix(p.(string), "arguments not match") {
				t.Fail()
			}
		}()
		lua.RunString(`foo(1)`)
	})

	t.Run("invalid bool", func(t *testing.T) {
		lua.RegisterFunction("foo", func(b bool) {
			if !b {
				t.Fail()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if p.(string) != "not a boolean" {
				t.Fail()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(true)`)
	})

	t.Run("invalid int", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i int) {
			if i != 42 {
				t.Fail()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if p.(string) != "not an integer" {
				t.Fail()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(42)`)
	})

	t.Run("invalid unsigned", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i uint) {
			if i != 42 {
				t.Fail()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if p.(string) != "not an unsigned" {
				t.Fail()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(42)`)
	})

	t.Run("invalid float", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i float64) {
			if i != 4.2 {
				t.Fail()
			}
		})
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if p.(string) != "not a float" {
				t.Fail()
			}
		}()
		lua.RunString(`foo({})`)
		lua.RunString(`foo(4.2)`)
	})

	t.Run("bool interface", func(t *testing.T) {
		lua.RegisterFunction("foo", func(i interface{}) {
			if i, ok := i.(bool); !ok {
				t.Fail()
			} else {
				if !i {
					t.Fail()
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
				t.Fail()
			}
			if !strings.HasPrefix(p.(string), "wrong interface argument") {
				t.Fail()
			}
		}()
		lua.RunString(`foo(function() end)`)
	})
}
