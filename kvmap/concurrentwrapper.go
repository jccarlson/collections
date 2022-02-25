package kvmap

import (
	"sync"
)

// ConcurrentWrapper wraps any kvmap.Interface so that its operations are
// thread-safe.
type ConcurrentWrapper[K, V any] struct {
	Base Interface[K, V]
	lock sync.RWMutex
}

func (m *ConcurrentWrapper[K, V]) Put(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Base.Put(key, value)
}

func (m *ConcurrentWrapper[K, V]) Get(key K) (value V, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Base.Get(key)
}

func (m *ConcurrentWrapper[K, V]) Has(key K) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Base.Has(key)
}

func (m *ConcurrentWrapper[K, V]) Delete(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Base.Delete(key)
}

func (m *ConcurrentWrapper[K, V]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Base.Len()
}