package kvmap

import (
	"iter"
)

func initMapWrapperOptions(opts []Option) kvMapOpts {
	r := kvMapOpts{capacity: -1}

	for _, opt := range opts {
		opt.setOpt(&r)
	}
	return r
}

// MapWrapper wraps a built-in map with kvmap.Interface.
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

// Put adds a key-value pair to the wrapped map.
func (m MapWrapper[K, V]) Put(key K, val V) {
	m[key] = val
}

// Get returns the value for the given key and ok == true if present, and ok ==
// false if not.
func (m MapWrapper[K, V]) Get(key K) (val V, ok bool) {
	val, ok = m[key]
	return
}

// Delete removes the value for the given key if present.
func (m MapWrapper[K, V]) Delete(key K) {
	delete(m, key)
}

// Has returns true if the given key is present in the map.
func (m MapWrapper[K, V]) Has(key K) bool {
	_, ok := m[key]
	return ok
}

// String returns a string representation of the map which is similar to the
// built-in map String() representation.
func (m MapWrapper[K, V]) String() string {
	return IterableMapToString(m)
}

// GoString returns a string representation of the map which is similar to the
// built-in map GoString() representation.
func (m MapWrapper[K, V]) GoString() string {
	return IterableMapToGoString(m)
}

// Len returns the number of key-value pairs in the map.
func (m MapWrapper[K, V]) Len() int {
	return len(m)
}

// All returns an iterator which yields the key-value pairs of the map.
func (m MapWrapper[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
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

// Entries returns an iterator which yields the key-value pairs wrapped in the
// Entry interface, which allows values to be modified via SetValue.
func (m MapWrapper[K, V]) Entries() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		for k, v := range m {
			if !yield(&wrapperEntry[K, V]{map[K]V(m), k, v}) {
				return
			}
		}
	}
}
