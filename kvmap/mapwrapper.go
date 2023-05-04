package kvmap

import (
	"github.org/jccarlson/collections"
	"github.org/jccarlson/collections/internal"
)

func initMapWrapperOptions(opts []Option) kvMapOpts {
	r := kvMapOpts{capacity: -1}

	for _, opt := range opts {
		opt.setOpt(&r)
	}
	return r
}

// MapWrapper wraps a built-in map with the Map and
// Iterator interfaces.
type MapWrapper[K comparable, V any] map[K]V

// NewMapWrapper returns an MapWrapper wrapping a new, empty map. The only
// supported Option is Capacity(), which sets the initial capacity of the
// underlying map. Options other than Capacity are ignored.
func NewMapWrapper[K comparable, V any](opts ...Option) MapWrapper[K, V] {
	o := initMapWrapperOptions(opts)
	if o.capacity >= 0 {
		return MapWrapper[K, V](make(map[K]V, o.capacity))
	}
	return MapWrapper[K, V](make(map[K]V))
}

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
	return IterableMapToString[K, V](m)
}

func (m MapWrapper[K, V]) GoString() string {
	return IterableMapToGoString[K, V](m)
}

func (m MapWrapper[K, V]) Len() int {
	return len(m)
}

func (m MapWrapper[K, V]) Iterator() collections.Iterator[Entry[K, V]] {

	sender, it := internal.NewChanIteratorPair[Entry[K, V]]()
	go func() {
		for k, v := range m {
			if !sender.Send(&wrapperEntry[K, V]{map[K]V(m), k, v}) {
				break
			}
		}
		sender.Close()
	}()
	return it
}

type wrapperEntry[K comparable, V any] struct {
	m     map[K]V
	key   K
	value V
}

func (e *wrapperEntry[K, V]) Key() K {
	return e.key
}
func (e *wrapperEntry[K, V]) Value() V {
	return e.value
}
func (e *wrapperEntry[K, V]) SetValue(v V) {
	e.value = v
	e.m[e.key] = v
}
