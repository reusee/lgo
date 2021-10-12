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
	"unsafe"

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
		args = append(args, function.lua.toGoValue(i, function.funcType.In(int(i-1))))
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

var stringType = reflect.TypeOf((*string)(nil)).Elem()
var intType = reflect.TypeOf((*int)(nil)).Elem()
var floatType = reflect.TypeOf((*float64)(nil)).Elem()
var boolType = reflect.TypeOf((*bool)(nil)).Elem()

func (lua *Lua) toGoValue(i C.int, paramType reflect.Type) (ret reflect.Value) {
	luaType := C.lua_type(lua.State, i)
	paramKind := paramType.Kind()
	switch paramKind {
	case reflect.Bool:
		if luaType != C.LUA_TBOOLEAN {
			lua.Panic("not a boolean")
		}
		ret = reflect.ValueOf(C.lua_toboolean(lua.State, i) == C.int(1))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if luaType != C.LUA_TNUMBER {
			lua.Panic("not an integer")
		}
		ret = reflect.New(paramType).Elem()
		ret.SetInt(int64(C.lua_tointegerx(lua.State, i, nil)))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if luaType != C.LUA_TNUMBER {
			lua.Panic("not an unsigned")
		}
		ret = reflect.New(paramType).Elem()
		ret.SetUint(uint64(C.lua_tointegerx(lua.State, i, nil)))
	case reflect.Float32, reflect.Float64:
		if luaType != C.LUA_TNUMBER {
			lua.Panic("not a float")
		}
		ret = reflect.New(paramType).Elem()
		ret.SetFloat(float64(C.lua_tonumberx(lua.State, i, nil)))
	case reflect.Interface:
		switch luaType {
		case C.LUA_TNUMBER:
			ret = reflect.New(floatType).Elem()
			ret.SetFloat(float64(C.lua_tonumberx(lua.State, i, nil)))
		case C.LUA_TSTRING:
			ret = reflect.New(stringType).Elem()
			ret.SetString(C.GoString(C.lua_tolstring(lua.State, i, nil)))
		case C.LUA_TLIGHTUSERDATA:
			ret = reflect.ValueOf(C.lua_topointer(lua.State, i))
		case C.LUA_TBOOLEAN:
			ret = reflect.New(boolType).Elem()
			ret.SetBool(C.lua_toboolean(lua.State, i) == C.int(1))
		//TODO nil
		//TODO table
		default:
			lua.Panic("wrong interface argument: %v", paramKind)
		}
	case reflect.String:
		if luaType != C.LUA_TSTRING {
			lua.Panic("not a string")
		}
		ret = reflect.New(paramType).Elem()
		ret.SetString(C.GoString(C.lua_tolstring(lua.State, i, nil)))
	case reflect.Slice:
		switch luaType {
		case C.LUA_TSTRING:
			ret = reflect.New(paramType).Elem()
			cstr := C.lua_tolstring(lua.State, i, nil)
			ret.SetBytes(C.GoBytes(unsafe.Pointer(cstr), C.int(C.strlen(cstr))))
		case C.LUA_TTABLE:
			ret = reflect.MakeSlice(paramType, 0, 0)
			C.lua_pushnil(lua.State)
			elemType := paramType.Elem()
			for C.lua_next(lua.State, i) != 0 {
				ret = reflect.Append(ret, lua.toGoValue(-1, elemType))
				C.lua_settop(lua.State, -2)
			}
		default:
			lua.Panic("wrong slice argument")
		}
	case reflect.Ptr:
		if luaType != C.LUA_TLIGHTUSERDATA {
			lua.Panic("not a pointer")
		}
		pointer := C.lua_topointer(lua.State, i)
		ret = reflect.NewAt(paramType, unsafe.Pointer(&pointer)).Elem()
	case reflect.Map:
		if luaType != C.LUA_TTABLE {
			lua.Panic("not a map")
		}
		ret = reflect.MakeMap(paramType)
		C.lua_pushnil(lua.State)
		keyType := paramType.Key()
		elemType := paramType.Elem()
		for C.lua_next(lua.State, i) != 0 {
			ret.SetMapIndex(
				lua.toGoValue(-2, keyType),
				lua.toGoValue(-1, elemType))
			C.lua_settop(lua.State, -2)
		}
	case reflect.UnsafePointer:
		ret = reflect.ValueOf(C.lua_topointer(lua.State, i))
	//TODO complex64/128
	//TODO array
	//TODO chan
	//TODO func
	//TODO struct
	default:
		lua.Panic("unknown argument type %v", paramType)
	}
	return
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
