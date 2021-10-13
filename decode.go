package lgo

import (
	"fmt"
	"reflect"

	"github.com/reusee/sb"
)

/*
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
*/
import "C"

func decodeStack(
	l *Lua,
	num C.int,
	t reflect.Type,
	cont proc,
) proc {
	return func() (*sb.Token, proc, error) {
		luaType := C.lua_type(l.State, num)
		switch luaType {

		case C.LUA_TNIL:
			return &sb.Token{
				Kind: sb.KindNil,
			}, cont, nil

		case C.LUA_TBOOLEAN:
			return &sb.Token{
				Kind:  sb.KindBool,
				Value: C.lua_toboolean(l.State, num) == C.int(1),
			}, cont, nil

		case C.LUA_TNUMBER:
			switch t.Kind() {

			case reflect.Int:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindInt,
					Value: int(n),
				}, cont, nil
			case reflect.Int8:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindInt8,
					Value: int8(n),
				}, cont, nil
			case reflect.Int16:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindInt16,
					Value: int16(n),
				}, cont, nil
			case reflect.Int32:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindInt32,
					Value: int32(n),
				}, cont, nil
			case reflect.Int64:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindInt64,
					Value: int64(n),
				}, cont, nil

			case reflect.Uint:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindUint,
					Value: uint(n),
				}, cont, nil
			case reflect.Uint8:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindUint8,
					Value: uint8(n),
				}, cont, nil
			case reflect.Uint16:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindUint16,
					Value: uint16(n),
				}, cont, nil
			case reflect.Uint32:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindUint32,
					Value: uint32(n),
				}, cont, nil
			case reflect.Uint64:
				n := C.lua_tointegerx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindUint64,
					Value: uint64(n),
				}, cont, nil

			case reflect.Float32:
				n := C.lua_tonumberx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindFloat32,
					Value: float32(n),
				}, cont, nil
			case reflect.Float64, reflect.Interface:
				n := C.lua_tonumberx(l.State, num, nil)
				return &sb.Token{
					Kind:  sb.KindFloat64,
					Value: float64(n),
				}, cont, nil

			default:
				panic(fmt.Errorf("type mismatch, expecting %v", t))
			}

		case C.LUA_TSTRING:
			str := C.GoString(C.lua_tolstring(l.State, num, nil))
			if t.Kind() == reflect.Slice &&
				t.Elem().Kind() == reflect.Uint8 {
				// []byte
				return &sb.Token{
					Kind:  sb.KindBytes,
					Value: []byte(str),
				}, cont, nil
			}
			return &sb.Token{
				Kind:  sb.KindString,
				Value: str,
			}, cont, nil

		case C.LUA_TTABLE:
			switch t.Kind() {

			case reflect.Slice:
				return &sb.Token{
					Kind: sb.KindArray,
				}, decodeArray(l, num, t, cont), nil

			case reflect.Struct:
				return &sb.Token{
					Kind: sb.KindObject,
				}, decodeObject(l, num, t, cont), nil

			case reflect.Map, reflect.Interface:
				return &sb.Token{
					Kind: sb.KindMap,
				}, decodeMap(l, num, t, cont), nil

			default:
				panic(fmt.Errorf("type mismatch, expecting %v", t))
			}

		case C.LUA_TFUNCTION:
			panic("function type not supported")

		default: // NOCOVER
			panic(fmt.Errorf("bad lua type: %v", luaType))
		}

	}
}

func decodeArray(
	l *Lua,
	num C.int,
	t reflect.Type,
	cont proc,
) proc {

	C.lua_pushnil(l.State)
	elemType := t.Elem()

	var ret proc
	ret = func() (*sb.Token, proc, error) {
		if C.lua_next(l.State, num) == 0 {
			return &sb.Token{
				Kind: sb.KindArrayEnd,
			}, cont, nil
		}

		return decodeStack(l, C.lua_gettop(l.State), elemType,
			func() (*sb.Token, proc, error) {
				C.lua_settop(l.State, -2)
				return nil, ret, nil
			},
		)()
	}

	return ret
}

func decodeObject(
	l *Lua,
	num C.int,
	t reflect.Type,
	cont proc,
) proc {

	C.lua_pushnil(l.State)

	var ret proc
	ret = func() (*sb.Token, proc, error) {
		if C.lua_next(l.State, num) == 0 {
			return &sb.Token{
				Kind: sb.KindObjectEnd,
			}, cont, nil
		}

		name := C.GoString(C.lua_tolstring(l.State, -2, nil))
		field, ok := t.FieldByName(name)
		if !ok {
			C.lua_settop(l.State, -2)
			return nil, ret, nil
		}
		return &sb.Token{
				Kind:  sb.KindString,
				Value: name,
			}, decodeStack(l, C.lua_gettop(l.State), field.Type,
				func() (*sb.Token, proc, error) {
					C.lua_settop(l.State, -2)
					return nil, ret, nil
				},
			), nil
	}
	return ret
}

func decodeMap(
	l *Lua,
	num C.int,
	t reflect.Type,
	cont proc,
) proc {

	C.lua_pushnil(l.State)
	keyType := t.Key()
	elemType := t.Elem()

	var ret proc
	ret = func() (*sb.Token, proc, error) {
		if C.lua_next(l.State, num) == 0 {
			return &sb.Token{
				Kind: sb.KindMapEnd,
			}, cont, nil
		}

		top := C.lua_gettop(l.State)
		return decodeStack(l, top-1, keyType,
			decodeStack(l, top, elemType,
				func() (*sb.Token, proc, error) {
					C.lua_settop(l.State, -2)
					return nil, ret, nil
				},
			),
		)()
	}
	return ret
}
