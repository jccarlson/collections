package kvmap

import (
	"golang.org/x/exp/constraints"

	"github.org/jccarlson/collections"
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

func (m *OrderedMap[K, V]) Put(key K, value V) {
	(*ds.RedBlackTree[Entry[K, V]])(m).Put(&orderedMapEntry[K, V]{key: key, value: &value})
}

func (m *OrderedMap[K, V]) Get(key K) (value V, ok bool) {
	entry, ok := (*ds.RedBlackTree[Entry[K, V]])(m).Get(&orderedMapEntry[K, V]{key: key})
	if ok {
		value = entry.Value()
	}
	return value, ok
}

func (m *OrderedMap[K, V]) Has(key K) bool {
	return (*ds.RedBlackTree[Entry[K, V]])(m).Has(&orderedMapEntry[K, V]{key: key})
}

func (m *OrderedMap[K, V]) Delete(key K) {
	(*ds.RedBlackTree[Entry[K, V]])(m).Delete(&orderedMapEntry[K, V]{key: key})
}

func (m *OrderedMap[K, V]) Len() int {
	return (*ds.RedBlackTree[Entry[K, V]])(m).Len()
}

func (m *OrderedMap[K, V]) String() string {
	return IterableMapToString[K, V](m)
}

func (m *OrderedMap[K, V]) GoString() string {
	return IterableMapToGoString[K, V](m)
}

type orderedMapIterator[K, V any] struct {
	direction ds.Direction
	tn        *ds.TreeNode[Entry[K, V]]
}

func (i *orderedMapIterator[K, V]) Next() (e Entry[K, V], ok bool) {
	if i.tn == nil {
		return
	}
	e = i.tn.Elem
	i.tn = i.tn.Walk(i.direction)
	return e, true
}

func (m *OrderedMap[K, V]) Iterator() collections.Iterator[Entry[K, V]] {
	return &orderedMapIterator[K, V]{direction: ds.Right, tn: (*ds.RedBlackTree[Entry[K, V]])(m).First()}
}

func (m *OrderedMap[K, V]) ReverseIterator() collections.Iterator[Entry[K, V]] {
	return &orderedMapIterator[K, V]{direction: ds.Left, tn: (*ds.RedBlackTree[Entry[K, V]])(m).Last()}
}
