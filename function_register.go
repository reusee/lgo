package lgo

import (
	"sync"
	"sync/atomic"
)

type FunctionRegister struct {
	s atomic.Value
	l sync.Mutex
}

func (m *FunctionRegister) Get(key int64) (value *Function, ok bool) {
	v := m.s.Load()
	if v != nil {
		value, ok = v.(map[int64]*Function)[key]
	}
	return
}

func (m *FunctionRegister) Set(key int64, value *Function) {
	m.l.Lock()
	defer m.l.Unlock()
	var fresh map[int64]*Function
	old, ok := m.s.Load().(map[int64]*Function)
	if ok {
		fresh = make(map[int64]*Function, len(old))
		for k, v := range old {
			fresh[k] = v
		}
	} else {
		fresh = make(map[int64]*Function)
	}
	fresh[key] = value
	m.s.Store(fresh)
}
