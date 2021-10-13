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
		lua.CallFunction("foo")
	})

	t.Run("bool argument", func(t *testing.T) {
		lua.RegisterFunction("bool", func(b bool) {
			if b != true {
				t.Fatal()
			}
		})
		lua.CallFunction("bool", true)
	})

	t.Run("int argument", func(t *testing.T) {
		lua.RegisterFunction("int", func(i int) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("int", 42)
	})

	t.Run("int8 argument", func(t *testing.T) {
		lua.RegisterFunction("int8", func(i int8) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("int8", int8(42))
	})

	t.Run("int16 argument", func(t *testing.T) {
		lua.RegisterFunction("int16", func(i int16) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("int16", int16(42))
	})

	t.Run("int32 argument", func(t *testing.T) {
		lua.RegisterFunction("int32", func(i int32) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("int32", int32(42))
	})

	t.Run("int64 argument", func(t *testing.T) {
		lua.RegisterFunction("int64", func(i int64) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("int64", int64(42))
	})

	t.Run("uint argument", func(t *testing.T) {
		lua.RegisterFunction("uint", func(i uint) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("uint", uint(42))
	})

	t.Run("uint8 argument", func(t *testing.T) {
		lua.RegisterFunction("uint8", func(i uint8) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("uint8", uint8(42))
	})

	t.Run("uint16 argument", func(t *testing.T) {
		lua.RegisterFunction("uint16", func(i uint16) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("uint16", uint16(42))
	})

	t.Run("uint32 argument", func(t *testing.T) {
		lua.RegisterFunction("uint32", func(i uint32) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("uint32", uint32(42))
	})

	t.Run("uint64 argument", func(t *testing.T) {
		lua.RegisterFunction("uint64", func(i uint64) {
			if i != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("uint64", uint64(42))
	})

	t.Run("float argument", func(t *testing.T) {
		lua.RegisterFunction("float", func(f float64) {
			if f != 42.5 {
				t.Fatalf("got %v", f)
			}
		})
		lua.CallFunction("float", float64(42.5))
	})

	t.Run("float32 argument", func(t *testing.T) {
		lua.RegisterFunction("float32", func(f float32) {
			if f != 42.5 {
				t.Fatalf("got %v", f)
			}
		})
		lua.CallFunction("float32", float32(42.5))
	})

	t.Run("interface argument", func(t *testing.T) {
		lua.RegisterFunction("interface", func(a, b interface{}) {
			if i, ok := a.(float64); !ok || i != 42 {
				t.Fatalf("got %v", a)
			}
			if s, ok := b.(string); !ok || s != "foo" {
				t.Fatal()
			}
		})
		lua.CallFunction("interface", 42, "foo")
	})

	t.Run("string argument", func(t *testing.T) {
		lua.RegisterFunction("str", func(s string) {
			if s != "foo" {
				t.Fatal()
			}
		})
		lua.CallFunction("str", "foo")
	})

	t.Run("bytes argument", func(t *testing.T) {
		lua.RegisterFunction("bytes", func(bs []byte) {
			if !bytes.Equal(bs, []byte("foo")) {
				t.Fatal()
			}
		})
		lua.CallFunction("bytes", "foo")
	})

	t.Run("slice argument", func(t *testing.T) {
		lua.RegisterFunction("slice", func(is []int) {
			if len(is) != 2 || is[0] != 4 || is[1] != 2 {
				t.Fatal()
			}
		})
		lua.CallFunction("slice", []int{4, 2})
	})

	t.Run("map argument", func(t *testing.T) {
		lua.RegisterFunction("map", func(m map[string]int) {
			if len(m) != 2 || m["foo"] != 4 || m["bar"] != 2 {
				t.Fatal()
			}
		})
		lua.CallFunction("map", map[string]int{
			"foo": 4,
			"bar": 2,
		})
	})

	t.Run("struct argument", func(t *testing.T) {
		type Baz struct {
			I int
		}
		type Bar struct {
			Baz Baz
		}
		type Foo struct {
			Bar Bar
		}
		lua.RegisterFunction("struct", func(foo Foo) {
			if foo.Bar.Baz.I != 42 {
				t.Fatal()
			}
		})
		lua.CallFunction("struct", Foo{
			Bar: Bar{
				Baz: Baz{
					I: 42,
				},
			},
		})
		lua.NonStrict = true
		lua.RunString(`struct{
      Bar = {
        Baz = {
          I = 42,
          Foo = 1
        }
      }
    }`)
	})

	t.Run("unsafepointer argument", func(t *testing.T) {
		lua.RegisterFunction("unsafepointer", func(m unsafe.Pointer) {
			if m != nil {
				t.Fatal()
			}
		})
		lua.CallFunction("unsafepointer", nil)
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
		lua.CallFunction("panic")
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
		lua.CallFunction("foo", 1)
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
			msg := fmt.Sprintf("%s", p)
			if msg != "type mismatch, expecting bool" {
				t.Fatalf("got %s", msg)
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
			msg := fmt.Sprintf("%s", p)
			if msg != "type mismatch, expecting int" {
				t.Fatalf("got %s", msg)
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
			msg := fmt.Sprintf("%s", p)
			if msg != "type mismatch, expecting uint" {
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
			msg := fmt.Sprintf("%s", p)
			if msg != "type mismatch, expecting float64" {
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
			msg := fmt.Sprintf("%s", p)
			if msg != "function type not supported" {
				t.Fatalf("got %s", msg)
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
			msg := fmt.Sprintf("%s", p)
			if msg != "type mismatch, expecting string" {
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
			msg := fmt.Sprintf("%s", p)
			if msg != "type mismatch, expecting []int" {
				t.Fatal()
			}
		}()
		lua.RunString("invalidslice(1)")
	})

	t.Run("pointer", func(t *testing.T) {
		lua.RegisterFunction("pointer", func(p *int) {
			if p == nil {
				t.Fatal()
			}
			if *p != 1 {
				t.Fatal()
			}
		})
		lua.CallFunction("pointer", 1)
	})

	t.Run("invalid map", func(t *testing.T) {
		lua.RegisterFunction("invalidmap", func(m map[int]bool) {})
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal()
			}
			msg := fmt.Sprintf("%s", p)
			if msg != "type mismatch, expecting map[int]bool" {
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
			msg := fmt.Sprintf("%s", p)
			if !strings.Contains(msg, "syntax error") {
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
			msg := fmt.Sprintf("%s", p)
			if !strings.Contains(msg, "foobarbaz") {
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
