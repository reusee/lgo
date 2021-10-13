package lgo

/*
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
#include <string.h>
#include <stdint.h>
#cgo pkg-config: lua

void register_function(lua_State*, const char*, int64_t);
void setup_message_handler(lua_State*);
int traceback(lua_State*);

*/
import "C"

import (
	"fmt"
	"reflect"
	"runtime/cgo"
	"strings"
	"sync"

	"github.com/reusee/sb"
)

var (
	cstrs sync.Map
)

func cstr(str string) *C.char {
	v, ok := cstrs.Load(str)
	if ok {
		return v.(*C.char)
	}
	c := C.CString(str)
	cstrs.Store(str, c)
	return c
}

type Lua struct {
	State          *C.lua_State
	PrintTraceback bool
}

type _Function struct {
	name      string
	lua       *Lua
	fun       interface{}
	funcType  reflect.Type
	funcValue reflect.Value
	argc      int
}

func New() *Lua {
	state := C.luaL_newstate()
	if state == nil { //NOCOVER
		panic("lua state create error")
	}
	C.luaL_openlibs(state)
	lua := &Lua{
		State:          state,
		PrintTraceback: true,
	}
	return lua
}

var NewLua = New

func (l *Lua) RegisterFunction(name string, fun interface{}) {
	path := strings.Split(name, ".")
	name = path[len(path)-1]
	path = path[0 : len(path)-1]
	if len(path) == 0 {
		path = append(path, "_G")
	}

	// ensure namespaces
	for i, namespace := range path {
		cNamespace := cstr(namespace)
		if i == 0 { // top namespace
			what := C.lua_getglobal(l.State, cNamespace)
			if what == C.LUA_TNIL { // not exists
				C.lua_settop(l.State, -2)
				C.lua_createtable(l.State, 0, 0)
				C.lua_setglobal(l.State, cNamespace)
				C.lua_getglobal(l.State, cNamespace)
			}
			if C.lua_type(l.State, -1) != C.LUA_TTABLE {
				l.Panic("global %s is not a table", namespace)
			}
		} else { // sub namespace
			C.lua_pushstring(l.State, cNamespace)
			C.lua_rawget(l.State, -2)
			if C.lua_type(l.State, -1) == C.LUA_TNIL {
				C.lua_settop(l.State, -2)
				C.lua_pushstring(l.State, cNamespace)
				C.lua_createtable(l.State, 0, 0)
				C.lua_rawset(l.State, -3)
				C.lua_pushstring(l.State, cNamespace)
				C.lua_rawget(l.State, -2)
			}
			if C.lua_type(l.State, -1) != C.LUA_TTABLE {
				l.Panic("namespace %s is not a table", namespace)
			}
		}
	}

	// register function
	funcType := reflect.TypeOf(fun)
	if funcType.IsVariadic() {
		l.Panic("cannot register variadic function: %v", fun)
	}
	argc := funcType.NumIn()
	cName := cstr(name)
	function := &_Function{
		fun:       fun,
		lua:       l,
		name:      name,
		funcType:  funcType,
		funcValue: reflect.ValueOf(fun),
		argc:      argc,
	}
	handle := cgo.NewHandle(function)
	C.register_function(l.State, cName, (C.int64_t)(handle))
	C.lua_settop(l.State, -2)
}

func (l *Lua) RegisterFunctions(funcs map[string]interface{}) {
	for name, fun := range funcs {
		l.RegisterFunction(name, fun)
	}
}

//export invoke
func invoke(_handle uint64) int {
	handle := cgo.Handle(_handle)
	function := handle.Value().(*_Function)
	// check argument count
	argc := C.lua_gettop(function.lua.State)
	if int(argc) != function.argc {
		function.lua.Panic("arguments not match: %v", function.fun)
	}
	// arguments
	var args []reflect.Value
	for i := C.int(1); i <= argc; i++ {
		t := function.funcType.In(int(i - 1))
		arg := reflect.New(t)
		proc := decodeStack(function.lua, C.int(i), t, nil)
		ce(sb.Copy(
			&proc,
			sb.Unmarshal(arg.Interface()),
		))
		args = append(args, arg.Elem())
	}
	// call and returns
	returnValues := function.funcValue.Call(args)
	if len(returnValues) != function.funcType.NumOut() { //NOCOVER
		function.lua.Panic("return values not match: %v", function.fun)
	}
	for _, v := range returnValues {
		proc := sb.MarshalValue(sb.DefaultCtx, v, nil)
		ce(sb.Copy(
			&proc,
			pushValue(function.lua, nil),
		))
	}
	return len(returnValues)
}

func (l *Lua) RunString(code string) {
	defer func() {
		if r := recover(); r != nil {
			if l.PrintTraceback { //NOCOVER
				print("============ start lua traceback ============\n")
				l.RunString(`print(debug.traceback())`)
				print("============ end lua traceback ==============\n")
			}
			panic(r)
		}
	}()
	cCode := cstr(code)
	C.setup_message_handler(l.State)
	if ret := C.luaL_loadstring(l.State, cCode); ret != C.int(0) {
		l.Panic("%s", C.GoString(C.lua_tolstring(l.State, -1, nil)))
	}
	ret := C.lua_pcallk(l.State, 0, 0, C.lua_gettop(l.State)-C.int(1), 0, nil)
	if ret != C.int(0) {
		l.Panic("%s", C.GoString(C.lua_tolstring(l.State, -1, nil)))
	}
	C.lua_settop(l.State, 0)
}

func (l *Lua) CallFunction(name string, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			if l.PrintTraceback { //NOCOVER
				print("============ start lua traceback ============\n")
				l.RunString(`print(debug.traceback())`)
				print("============ end lua traceback ==============\n")
			}
			panic(r)
		}
	}()
	cName := cstr(name)
	C.setup_message_handler(l.State)
	C.lua_getglobal(l.State, cName)
	for _, arg := range args {
		ce(sb.Copy(
			sb.Marshal(arg),
			pushValue(l, nil),
		))
	}
	ret := C.lua_pcallk(l.State, C.int(len(args)), 0, C.lua_gettop(l.State)-C.int(len(args)+2), 0, nil)
	if ret != C.int(0) {
		l.Panic("%s", C.GoString(C.lua_tolstring(l.State, -1, nil)))
	}
}

func (l *Lua) Panic(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
