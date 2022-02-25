package kvmap

import "github.org/jccarlson/collections"

// MapWrapper wraps a built-in map with the Map and
// Iterator interfaces.
type MapWrapper[K comparable, V any] map[K]V

func (m MapWrapper[K, V]) Put(key K, val V) {
	m[key] = val
}

func (m MapWrapper[K, V]) Get(key K) (val V, ok bool) {
	val, ok = m[key]
	return
}

func (m MapWrapper[K, V]) Delete(key K) {
	delete(m, key)
}

func (m MapWrapper[K, V]) Has(key K) bool {
	_, ok := m[key]
	return ok
}

func (m MapWrapper[K, V]) String() string {
	return iterableMapToString[K, V](m)
}

func (m MapWrapper[K, V]) GoString() string {
	return iterableMapToGoString[K, V](m)
}

func (m MapWrapper[K, V]) Len() int {
	return len(m)
}

func (m MapWrapper[K, V]) Iterator() collections.Iterator[Entry[K, V]] {
	i := make(chan Entry[K, V])
	go func() {
		for k, v := range m {
			i <- &wrapperEntry[K, V]{k, v}
		}
		close(i)
	}()
	return entryChanIterator[K, V](i)
}

type wrapperEntry[K comparable, V any] struct {
	key   K
	value V
}

func (e *wrapperEntry[K, V]) Key() K {
	return e.key
}
func (e *wrapperEntry[K, V]) Value() V {
	return e.value
}
