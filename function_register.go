package lgo

import (
	"sync"
	"sync/atomic"
)

type _FunctionRegister struct {
	s atomic.Value
	l sync.Mutex
}

func (m *_FunctionRegister) Get(key int64) (value *_Function, ok bool) {
	v := m.s.Load()
	if v != nil {
		value, ok = v.(map[int64]*_Function)[key]
	}
	return
}

func (m *_FunctionRegister) Set(key int64, value *_Function) {
	m.l.Lock()
	defer m.l.Unlock()
	var fresh map[int64]*_Function
	old, ok := m.s.Load().(map[int64]*_Function)
	if ok {
		fresh = make(map[int64]*_Function, len(old)+1)
		for k, v := range old {
			fresh[k] = v
		}
	} else {
		fresh = make(map[int64]*_Function)
	}
	fresh[key] = value
	m.s.Store(fresh)
}
