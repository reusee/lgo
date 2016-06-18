#include "lua.h"
#include <stdint.h>

extern int Invoke(int64_t);

int invoke_go_func(lua_State* state) {
  int64_t func_id = lua_tointeger(state, lua_upvalueindex(1));
  return Invoke(func_id);
}

void register_function(lua_State* state, const char* name, int64_t func_id) {
  lua_pushstring(state, name);
  lua_pushinteger(state, func_id);
  lua_pushcclosure(state, (lua_CFunction)invoke_go_func, 1);
  lua_rawset(state, -3);
}

int traceback(lua_State* L) {
  lua_getglobal(L, "debug");
  lua_getfield(L, -1, "traceback");
  lua_pushvalue(L, 1);
  lua_pushinteger(L, 2);
  lua_call(L, 2, 1);
  return 1;
}

void setup_message_handler(lua_State* L) {
  lua_pushcfunction(L, traceback);
}
