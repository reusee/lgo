package lgo

/*
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
#include <stdlib.h>
#include <string.h>
#cgo pkg-config: lua

void register_function(lua_State*, const char*, void*);
void setup_message_handler(lua_State*);

*/
import "C"

import (
	"log"
	"reflect"
	"unsafe"
)

type Lua struct {
	State     *C.lua_State
	Functions map[string]*Function
}

type Function struct {
	name  string
	fun   interface{}
	state *C.lua_State
}

func NewLua() *Lua {
	state := C.luaL_newstate()
	if state == nil {
		log.Fatalf("lua state create error")
	}
	C.luaL_openlibs(state)
	lua := &Lua{
		State:     state,
		Functions: make(map[string]*Function),
	}
	return lua
}

func (self *Lua) RegisterFunction(name string, fun interface{}) {
	funcType := reflect.TypeOf(fun)
	if funcType.IsVariadic() {
		log.Fatalf("cannot register variadic function: %v", fun)
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	function := &Function{
		fun:   fun,
		state: self.State,
		name:  name,
	}
	C.register_function(self.State, cName, unsafe.Pointer(function))
	self.Functions[name] = function
}

//export Invoke
func Invoke(p unsafe.Pointer) int {
	function := (*Function)(p)
	state := function.state
	funcType := reflect.TypeOf(function.fun)
	// check argument count
	argc := C.lua_gettop(state)
	if int(argc) != funcType.NumIn() {
		log.Fatalf("arguments not match: %v", function.fun)
	}
	// arguments
	var args []reflect.Value
	var paramType reflect.Type
	for i := C.int(1); i <= argc; i++ {
		args = append(args, toGoValue(state, i, paramType))
	}
	// call
	funcValue := reflect.ValueOf(function.fun)
	// returns
	returnValues := funcValue.Call(args)
	for _, v := range returnValues {
		pushGoValue(state, v)
	}
	if len(returnValues) != funcType.NumOut() {
		log.Fatalf("return values not match: %v", function.fun)
	}
	return len(returnValues)
}

var stringType = reflect.TypeOf("")
var intType = reflect.TypeOf(int(0))

func toGoValue(state *C.lua_State, i C.int, paramType reflect.Type) (ret reflect.Value) {
	luaType := C.lua_type(state, i)
	if luaType == C.LUA_TBOOLEAN { // boolean
		ret = reflect.ValueOf(C.lua_toboolean(state, i) == C.int(1))
	} else if luaType == C.LUA_TNUMBER { // int, uint or float
		switch paramType.Kind() {
		case reflect.Interface:
			ret = reflect.New(intType).Elem()
			ret.SetInt(int64(C.lua_tointegerx(state, i, nil)))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ret = reflect.New(paramType).Elem()
			ret.SetInt(int64(C.lua_tointegerx(state, i, nil)))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ret = reflect.New(paramType).Elem()
			ret.SetUint(uint64(C.lua_tointegerx(state, i, nil)))
		default:
			ret = reflect.New(paramType).Elem()
			ret.SetFloat(float64(C.lua_tonumberx(state, i, nil)))
		}
	} else if luaType == C.LUA_TSTRING { // string or bytes
		switch paramType.Kind() {
		case reflect.Interface:
			ret = reflect.New(stringType).Elem()
			ret.SetString(C.GoString(C.lua_tolstring(state, i, nil)))
		case reflect.String:
			ret = reflect.New(paramType).Elem()
			ret.SetString(C.GoString(C.lua_tolstring(state, i, nil)))
		case reflect.Slice:
			ret = reflect.New(paramType).Elem()
			cstr := C.lua_tolstring(state, i, nil)
			ret.SetBytes(C.GoBytes(unsafe.Pointer(cstr), C.int(C.strlen(cstr))))
		default:
			log.Fatalf("invalid string argument")
		}
	} else if luaType == C.LUA_TLIGHTUSERDATA { // pointer
		ret = reflect.ValueOf(C.lua_topointer(state, i))
	} else if luaType == C.LUA_TTABLE { // slice or map
		switch paramType.Kind() {
		case reflect.Slice: // slice
			ret = reflect.MakeSlice(paramType, 0, 0)
			C.lua_pushnil(state)
			elemType := paramType.Elem()
			for C.lua_next(state, i) != 0 {
				ret = reflect.Append(ret, toGoValue(state, -1, elemType))
				C.lua_settop(state, -2)
			}
		case reflect.Map: // map
			ret = reflect.MakeMap(paramType)
			C.lua_pushnil(state)
			keyType := paramType.Key()
			elemType := paramType.Elem()
			for C.lua_next(state, i) != 0 {
				ret.SetMapIndex(
					toGoValue(state, -2, keyType),
					toGoValue(state, -1, elemType))
				C.lua_settop(state, -2)
			}
		default:
			log.Fatalf("cannot assign lua table to %v", paramType)
		}
	} else {
		log.Fatalf("invalid argument type")
	}
	return
}

func pushGoValue(state *C.lua_State, value reflect.Value) {
	switch t := value.Type(); t.Kind() {
	case reflect.Bool:
		if value.Bool() {
			C.lua_pushboolean(state, C.int(1))
		} else {
			C.lua_pushboolean(state, C.int(0))
		}
	case reflect.String:
		C.lua_pushstring(state, C.CString(value.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		C.lua_pushnumber(state, C.lua_Number(C.longlong(value.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		C.lua_pushnumber(state, C.lua_Number(C.ulonglong(value.Uint())))
	case reflect.Float32, reflect.Float64:
		C.lua_pushnumber(state, C.lua_Number(C.double(value.Float())))
	case reflect.Slice:
		length := value.Len()
		C.lua_createtable(state, C.int(length), 0)
		for i := 0; i < length; i++ {
			C.lua_pushnumber(state, C.lua_Number(i+1))
			pushGoValue(state, value.Index(i))
			C.lua_settable(state, -3)
		}
	case reflect.Interface:
		pushGoValue(state, value.Elem())
	case reflect.Ptr:
		C.lua_pushlightuserdata(state, unsafe.Pointer(value.Pointer()))
	default:
		log.Fatalf("wrong return value %v %v", value, t.Kind())
	}
}

func (self *Lua) RunString(code string) {
	cCode := C.CString(code)
	defer C.free(unsafe.Pointer(cCode))
	C.setup_message_handler(self.State)
	if ret := C.luaL_loadstring(self.State, cCode); ret != C.int(0) {
		log.Fatalf("%s", C.GoString(C.lua_tolstring(self.State, -1, nil)))
	}
	ret := C.lua_pcallk(self.State, 0, 0, C.lua_gettop(self.State)-C.int(1), 0, nil)
	if ret != C.int(0) {
		log.Fatalf("%s", C.GoString(C.lua_tolstring(self.State, -1, nil)))
	}
}
