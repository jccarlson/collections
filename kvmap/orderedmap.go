package kvmap

import (
	"iter"

	"golang.org/x/exp/constraints"

	"github.org/jccarlson/collections/compare"
	"github.org/jccarlson/collections/internal/ds"
)

// orderedMapEntry is a struct wrapping a Key-Value pair in a
// TreeMap which satisfies the Entry interface.
type orderedMapEntry[K, V any] struct {
	key   K
	value *V
}

func (e *orderedMapEntry[K, V]) Key() K {
	return e.key
}

func (e *orderedMapEntry[K, V]) Value() V {
	return *e.value
}

func (e *orderedMapEntry[K, V]) SetValue(v V) {
	*e.value = v
}

// NewOrderedMap returns a new, empty OrderedMap with constraints.Ordered keys
// (i.e. keys which support the '<' operator) and any value type.
func NewOrderedMap[K constraints.Ordered, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		Ordering: func(o1, o2 Entry[K, V]) bool {
			return compare.Less(o1.Key(), o2.Key())
		},
	}
}

// NewOrderedMapWithOrderableKeys returns a new, empty OrderedMap with
// compare.Orderable keys and any value type.
func NewOrderedMapWithOrderableKeys[K compare.Orderable[K], V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		Ordering: func(o1, o2 Entry[K, V]) bool {
			return compare.OrderableOrdering(o1.Key(), o2.Key())
		},
	}
}

// NewOrderedMapWithOrdering returns a new, empty OrderedMap with any key
// and value type, using ordering to order keys.
func NewOrderedMapWithOrdering[K, V any](ordering compare.Ordering[K]) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		Ordering: func(o1, o2 Entry[K, V]) bool {
			return ordering(o1.Key(), o2.Key())
		},
	}
}

// OrderedMap is a mapping of keys of type K to values of type
// V, which iterates over entries in key order.
type OrderedMap[K, V any] ds.RedBlackTree[Entry[K, V]]

// Put adds a key-value pair to the wrapped map.
func (m *OrderedMap[K, V]) Put(key K, value V) {
	(*ds.RedBlackTree[Entry[K, V]])(m).Put(&orderedMapEntry[K, V]{
		key:   key,
		value: &value,
	})
}

// Get returns the value for the given key and ok == true if present, and ok ==
// false if not.
func (m *OrderedMap[K, V]) Get(key K) (value V, ok bool) {
	entry, ok := (*ds.RedBlackTree[Entry[K, V]])(m).Get(&orderedMapEntry[K, V]{key: key})
	if ok {
		value = entry.Value()
	}
	return value, ok
}

// Has returns true if the given key is present in the map.
func (m *OrderedMap[K, V]) Has(key K) bool {
	return (*ds.RedBlackTree[Entry[K, V]])(m).Has(&orderedMapEntry[K, V]{key: key})
}

// Delete removes the value for the given key if present.
func (m *OrderedMap[K, V]) Delete(key K) {
	(*ds.RedBlackTree[Entry[K, V]])(m).Delete(&orderedMapEntry[K, V]{key: key})
}

// Len returns the number of key-value pairs in the map.
func (m *OrderedMap[K, V]) Len() int {
	return (*ds.RedBlackTree[Entry[K, V]])(m).Len()
}

// String returns a string representation of the map which is similar to the
// built-in map String() representation.
func (m *OrderedMap[K, V]) String() string {
	return IterableMapToString(m)
}

// GoString returns a string representation of the map which is similar to the
// built-in map GoString() representation.
func (m *OrderedMap[K, V]) GoString() string {
	return IterableMapToGoString(m)
}

type orderedMapIterator[K, V any] struct {
	direction ds.Direction
	tn        *ds.TreeNode[Entry[K, V]]
}

func (i *orderedMapIterator[K, V]) next() (e Entry[K, V], ok bool) {
	if i.tn == nil {
		return
	}
	e = i.tn.Elem
	i.tn = i.tn.Walk(i.direction)
	return e, true
}

// All returns an iterator which yields the key-value pairs of the map in
// order.
func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		it := &orderedMapIterator[K, V]{
			direction: ds.Right,
			tn:        (*ds.RedBlackTree[Entry[K, V]])(m).First(),
		}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e.Key(), e.Value()) {
				return
			}
		}
	}
}

// Backwards returns an iterator which yields the key-value pairs of the map in
// reverse order.
func (m *OrderedMap[K, V]) Backwards() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		it := &orderedMapIterator[K, V]{
			direction: ds.Left,
			tn:        (*ds.RedBlackTree[Entry[K, V]])(m).Last(),
		}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e.Key(), e.Value()) {
				return
			}
		}
	}
}

// Entries returns an iterator which yields the key-value pairs wrapped in the
// Entry interface in order, which allows values to be modified via SetValue.
func (m *OrderedMap[K, V]) Entries() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		it := &orderedMapIterator[K, V]{
			direction: ds.Right,
			tn:        (*ds.RedBlackTree[Entry[K, V]])(m).First(),
		}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e) {
				return
			}
		}
	}
}

// EntriesBackwards returns an iterator which yields the key-value pairs
// wrapped in the Entry interface in reverse order, which allows values to be
// modified via SetValue.
func (m *OrderedMap[K, V]) EntriesBackwards() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		it := &orderedMapIterator[K, V]{
			direction: ds.Left,
			tn:        (*ds.RedBlackTree[Entry[K, V]])(m).Last(),
		}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e) {
				return
			}
		}
	}
}
