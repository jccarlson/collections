package kvmap

import (
	"constraints"

	"github.org/jccarlson/collections"
	"github.org/jccarlson/collections/compare"
)

// treeMapEntry is a struct wrapping a Key-Value pair in a
// TreeMap.
type treeMapEntry[K, V any] struct {
	key   K
	value V

	left, right *treeMapEntry[K, V]
}

func (e *treeMapEntry[K, V]) Key() K {
	return e.key
}

func (e *treeMapEntry[K, V]) Value() V {
	return e.value
}

func NewOrderedTreeMap[K constraints.Ordered, V any]() *TreeMap[K, V] {
	return &TreeMap[K, V]{ordering: compare.Less[K]}
}

func NewOrdererTreeMap[K compare.Orderer[K], V any]() *TreeMap[K, V] {
	return &TreeMap[K, V]{ordering: compare.DefaultOrdering[K]}
}

func NewCustomOrderingTreeMap[K, V any](ordering compare.Ordering[K]) *TreeMap[K, V]{
	return &TreeMap[K,V]{ordering: ordering}
}

// TreeMap is a balanced binary tree mapping keys of type K to values of type
// V, which iterates over entries based on the Ordering.
type TreeMap[K, V any] struct {
	ordering compare.Ordering[K]

	root *treeMapEntry[K, V]
	size int
}

func (m *TreeMap[K, V]) Put(key K, value V) {
	putRecursive(&m.root, &treeMapEntry[K, V]{key: key, value: value}, key, m.ordering)
	m.size++
}

func putRecursive[K, V any](root **treeMapEntry[K, V], e *treeMapEntry[K, V], key K, before compare.Ordering[K]) {
	if *root == nil {
		*root = e
		return
	}
	if before(key, (*root).key) {
		putRecursive(&(*root).left, e, key, before)
		return
	}
	if before((*root).key, key) {
		putRecursive(&(*root).right, e, key, before)
		return
	}
	*root = e
}

func (m *TreeMap[K, V]) Get(key K) (value V, ok bool) {
	return getRecursive(m.root, key, m.ordering)
}

func (m *TreeMap[K, V]) Has(key K) bool {
	_, ok := getRecursive(m.root, key, m.ordering)
	return ok
}

func getRecursive[K, V any](root *treeMapEntry[K,V], key K, before compare.Ordering[K]) (value V, ok bool) {
	if root == nil {
		return
	}
	if before(key, root.key) {
		return getRecursive(root.left,key, before)
	}
	if before (root.key, key) {
		return getRecursive(root.right, key, before)
	}
	return root.value, true
}

func (m *TreeMap[K, V]) Delete(key K) {
	if m.Has(key) {
		putRecursive(&m.root, nil, key, m.ordering)
		m.size--
	}
}

func (m *TreeMap[K, V]) Len() int {
	return m.size
}

func (m *TreeMap[K, V]) String() string {
	return iterableMapToString[K, V](m)
}

func (m *TreeMap[K, V]) GoString() string {
	return iterableMapToGoString[K, V](m)
}

func (m *TreeMap[K, V]) Iterator() collections.Iterator[Entry[K, V]] {
	i := make(chan Entry[K, V])
	go func() {
		itRecursive(m.root, i)
		close(i)
	}()
	return entryChanIterator[K, V](i)
}

func itRecursive[K, V any](root *treeMapEntry[K,V], it chan<- Entry[K,V]) {
	if root == nil {
		return
	}
	itRecursive(root.left, it)
	it <- root
	itRecursive(root.right, it)
}

