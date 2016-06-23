package lgo

import (
	"bytes"
	"strings"
	"testing"
	"unsafe"
)

func TestArgument(t *testing.T) {
	lua := NewLua()

	lua.RegisterFunction("foo", func() {})
	lua.RunString(`foo()`)

	lua.RegisterFunction("bool", func(b bool) {
		if b != true {
			t.Fail()
		}
	})
	lua.RunString(`bool(true)`)
	//lua.RunString(`bool(1)`)

	lua.RegisterFunction("int", func(i int) {
		if i != 42 {
			t.Fail()
		}
	})
	lua.RunString(`int(42)`)
	//lua.RunString(`int(true)`)

	lua.RegisterFunction("uint", func(i uint) {
		if i != 42 {
			t.Fail()
		}
	})
	lua.RunString(`uint(42)`)
	//lua.RunString(`uint(true)`)

	lua.RegisterFunction("float", func(f float64) {
		if f != 42.5 {
			t.Fail()
		}
	})
	lua.RunString(`float(42.5)`)
	//lua.RunString(`float(true)`)

	lua.RegisterFunction("interface", func(a, b interface{}) {
		if i, ok := a.(float64); !ok || i != 42 {
			t.Fail()
		}
		if s, ok := b.(string); !ok || s != "foo" {
			t.Fail()
		}
	})
	lua.RunString(`interface(42, 'foo')`)

	lua.RegisterFunction("str", func(s string) {
		if s != "foo" {
			t.Fail()
		}
	})
	lua.RunString(`str('foo')`)

	lua.RegisterFunction("bytes", func(bs []byte) {
		if !bytes.Equal(bs, []byte("foo")) {
			t.Fail()
		}
	})
	lua.RunString(`bytes('foo')`)

	lua.RegisterFunction("slice", func(is []int) {
		if len(is) != 2 || is[0] != 4 || is[1] != 2 {
			t.Fail()
		}
	})
	lua.RunString(`slice{4, 2}`)
	//lua.RunString(`slice{4, 'foo'}`)

	lua.RegisterFunction("pointer", func(p unsafe.Pointer) {
	})

	lua.RegisterFunction("map", func(m map[string]int) {
		if len(m) != 2 || m["foo"] != 4 || m["bar"] != 2 {
			t.Fail()
		}
	})
	lua.RunString(`map{foo = 4, bar = 2}`)

	lua.RegisterFunction("unsafepointer", func(m unsafe.Pointer) {
		if m != nil {
			t.Fail()
		}
	})
	lua.RunString(`unsafepointer(nil)`)

	lua.RegisterFunction("foo", func(s struct{}) {})
	func() {
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
	}()
}

func TestReturns(t *testing.T) {
	lua := NewLua()

	lua.RegisterFunction("bool", func() bool {
		return true
	})
	lua.RunString(`if bool() ~= true then error('not true') end`)

	lua.RegisterFunction("str", func() string {
		return "foo"
	})
	lua.RunString(`if str() ~= 'foo' then error('not string') end`)

	lua.RegisterFunction("num", func() (int32, uint64, float32) {
		return 42, 99, 33.3
	})
	lua.RunString(`
	i, u, f = num()
	if i ~= 42 then error('not int32') end
	if u ~= 99 then error('not uint64') end
	if f - 33.3 > 0.000001 then error('not float32') end
	`)

	lua.RegisterFunction("slice", func() []int {
		return []int{1, 2, 3}
	})
	lua.RunString(`
	ret = slice()
	if #ret ~= 3 then error('slice error') end
	if ret[1] ~= 1 or ret[2] ~= 2 or ret[3] ~= 3 then error('slice error') end
	`)

	lua.RegisterFunction("interface", func() interface{} {
		return "foo"
	})
	lua.RunString(`if interface() ~= 'foo' then error('interface error') end`)

	lua.RegisterFunction("ptr", func() *int {
		i := 5
		return &i
	})
	lua.RunString(`if type(ptr()) ~= 'userdata' then error('not userdata') end`)

	lua.RegisterFunctions(map[string]interface{}{
		"foo": func() {},
		"bar": func() {},
	})
	lua.RunString(`
	foo()
	bar()
	`)

	func() {
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
	}()
}

func TestPanic(t *testing.T) {
	lua := NewLua()
	lua.RegisterFunction("foo", func(i int) {})
	//lua.RunString(`foo(true)`)

	lua.RegisterFunction("panic", func() {
		panic("foo")
	})
	//lua.RunString(`panic()`)
}

func TestNamespace(t *testing.T) {
	lua := NewLua()
	lua.RegisterFunction("foo.bar", func(i int) {
		if i != 42 {
			t.Fail()
		}
	})
	lua.RunString(`foo.bar(42)`)

	lua.RegisterFunction("bar.bar.bar", func(i int) {
		if i != 42 {
			t.Fail()
		}
	})
	lua.RunString(`bar.bar.bar(42)`)

	lua.RegisterFunction("bar.foo.baz.quux", func(i int) {
		if i != 42 {
			t.Fail()
		}
	})
	lua.RunString(`bar.foo.baz.quux(42)`)

	lua.RunString(`Foo = 1`)
	func() {
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
	}()

	lua.RunString(`Bar = {}`)
	lua.RunString(`Bar.Bar = 1`)
	func() {
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
	}()
}

func TestMethod(t *testing.T) {
	lua := NewLua()
	lua.RegisterFunction("method", func(self *Lua) {
		if self != lua {
			t.Fail()
		}
	})
	lua.RunString(`method()`)
}

func TestCallFunction(t *testing.T) {
	lua := NewLua()
	lua.RunString(`
	function foo(arg)
		if arg ~= 42 then error('not 42') end
	end
	`)
	lua.CallFunction("foo", 42)

	lua.RegisterFunction("foo", func(a, b int) {
	})
	func() {
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
	}()

	s := "foo"
	lua.RegisterFunction("foo", func(a *string) {
		if a != &s {
			t.Fatal("passing wrong pointer")
		}
	})
	lua.RegisterFunction("a_pointer", func() *string {
		return &s
	})
	func() {
		defer func() {
			p := recover()
			if p == nil {
				t.Fail()
			}
			if p.(string) != "not a pointer" {
				t.Fail()
			}
		}()
		lua.RunString(`foo(1)`)
	}()
	lua.RunString(`foo(a_pointer())`)

	lua.RegisterFunction("foo", func(b bool) {
		if !b {
			t.Fail()
		}
	})
	func() {
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
	}()
	lua.RunString(`foo(true)`)

	lua.RegisterFunction("foo", func(i int) {
		if i != 42 {
			t.Fail()
		}
	})
	func() {
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
	}()
	lua.RunString(`foo(42)`)

	lua.RegisterFunction("foo", func(i uint) {
		if i != 42 {
			t.Fail()
		}
	})
	func() {
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
	}()
	lua.RunString(`foo(42)`)

	lua.RegisterFunction("foo", func(i float64) {
		if i != 4.2 {
			t.Fail()
		}
	})
	func() {
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
	}()
	lua.RunString(`foo(4.2)`)

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

	lua.RegisterFunction("foo", func(i interface{}) {
	})
	func() {
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
	}()
}
