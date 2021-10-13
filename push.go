package lgo

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/reusee/sb"
)

/*
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
*/
import "C"

func pushValue(l *Lua, cont sink) sink {
	return func(token *sb.Token) (sink, error) {
		if token == nil { // NOCOVER
			return nil, fmt.Errorf("expecting value")
		}

		switch token.Kind {

		case sb.KindNil:
			C.lua_pushnil(l.State)

		case sb.KindBool:
			if token.Value.(bool) {
				C.lua_pushboolean(l.State, C.int(1))
			} else {
				C.lua_pushboolean(l.State, C.int(0))
			}

		case sb.KindString:
			C.lua_pushstring(l.State, C.CString(token.Value.(string)))

		case sb.KindInt:
			C.lua_pushnumber(l.State, C.lua_Number(C.longlong(token.Value.(int))))
		case sb.KindInt8:
			C.lua_pushnumber(l.State, C.lua_Number(C.longlong(token.Value.(int8))))
		case sb.KindInt16:
			C.lua_pushnumber(l.State, C.lua_Number(C.longlong(token.Value.(int16))))
		case sb.KindInt32:
			C.lua_pushnumber(l.State, C.lua_Number(C.longlong(token.Value.(int32))))
		case sb.KindInt64:
			C.lua_pushnumber(l.State, C.lua_Number(C.longlong(token.Value.(int64))))

		case sb.KindUint:
			C.lua_pushnumber(l.State, C.lua_Number(C.ulonglong(token.Value.(uint))))
		case sb.KindUint8:
			C.lua_pushnumber(l.State, C.lua_Number(C.ulonglong(token.Value.(uint8))))
		case sb.KindUint16:
			C.lua_pushnumber(l.State, C.lua_Number(C.ulonglong(token.Value.(uint16))))
		case sb.KindUint32:
			C.lua_pushnumber(l.State, C.lua_Number(C.ulonglong(token.Value.(uint32))))
		case sb.KindUint64:
			C.lua_pushnumber(l.State, C.lua_Number(C.ulonglong(token.Value.(uint64))))

		case sb.KindFloat32:
			C.lua_pushnumber(l.State, C.lua_Number(C.double(token.Value.(float32))))
		case sb.KindFloat64:
			C.lua_pushnumber(l.State, C.lua_Number(C.double(token.Value.(float64))))

		case sb.KindArray:
			C.lua_createtable(l.State, 0, 0)
			return pushArray(l, 1, cont), nil

		case sb.KindMap:
			C.lua_createtable(l.State, 0, 0)
			return pushMap(l, cont), nil

		case sb.KindObject:
			C.lua_createtable(l.State, 0, 0)
			return pushObject(l, cont), nil

		case sb.KindPointer:
			C.lua_pushlightuserdata(l.State, unsafe.Pointer(token.Value.(uintptr)))

		default:
			l.Panic("invalid value: %s", token)
		}

		return cont, nil
	}
}

func pushArray(l *Lua, num int, cont sink) sink {
	return func(token *sb.Token) (sink, error) {
		if token == nil { // NOCOVER
			return nil, io.ErrUnexpectedEOF
		}
		if token.Kind == sb.KindArrayEnd {
			return cont, nil
		}
		C.lua_pushnumber(l.State, C.lua_Number(num))
		return pushValue(
			l,
			func(token *sb.Token) (sink, error) {
				C.lua_settable(l.State, -3)
				return pushArray(l, num+1, cont).Sink(token)
			},
		).Sink(token)
	}
}

func pushMap(l *Lua, cont sink) sink {
	return func(token *sb.Token) (sink, error) {
		if token == nil { // NOCOVER
			return nil, io.ErrUnexpectedEOF
		}
		if token.Kind == sb.KindMapEnd {
			return cont, nil
		}
		return pushValue( // key
			l,
			func(token *sb.Token) (sink, error) {
				return pushValue( // value
					l,
					func(token *sb.Token) (sink, error) {
						C.lua_settable(l.State, -3)
						return pushMap(l, cont).Sink(token)
					},
				).Sink(token)
			},
		).Sink(token)
	}
}

func pushObject(l *Lua, cont sink) sink {
	return func(token *sb.Token) (sink, error) {
		if token == nil { // NOCOVER
			return nil, io.ErrUnexpectedEOF
		}
		if token.Kind == sb.KindObjectEnd {
			return cont, nil
		}
		return pushValue( // key
			l,
			func(token *sb.Token) (sink, error) {
				return pushValue( // value
					l,
					func(token *sb.Token) (sink, error) {
						C.lua_settable(l.State, -3)
						return pushObject(l, cont).Sink(token)
					},
				).Sink(token)
			},
		).Sink(token)
	}
}
