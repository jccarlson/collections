package kvmap

import (
	"constraints"

	"github.org/jccarlson/collections"
	"github.org/jccarlson/collections/compare"
)

type color byte
const (
	black color = iota
	red
)

// treeMapEntry is a struct wrapping a Key-Value pair in a
// TreeMap.
type treeMapEntry[K, V any] struct {
	key   K
	value V

	left, right *treeMapEntry[K, V]

	nodeColor color
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

func NewCustomOrderingTreeMap[K, V any](ordering compare.Ordering[K]) *TreeMap[K, V] {
	return &TreeMap[K, V]{ordering: ordering}
}

// TreeMap is a balanced binary tree mapping keys of type K to values of type
// V, which iterates over entries based on the Ordering.
type TreeMap[K, V any] struct {
	ordering compare.Ordering[K]

	root *treeMapEntry[K, V]
	size int
}

func (m *TreeMap[K, V]) Put(key K, value V) {
	m.size += putRecursive(&m.root, &treeMapEntry[K, V]{key: key, value: value}, key, m.ordering)

}

func putRecursive[K, V any](root **treeMapEntry[K, V], e *treeMapEntry[K, V], key K, before compare.Ordering[K]) int {
	if *root == nil {
		*root = e
		return 1
	}
	if before(key, (*root).key) {
		return putRecursive(&(*root).left, e, key, before)

	}
	if before((*root).key, key) {
		return putRecursive(&(*root).right, e, key, before)

	}
	(*root).value = e.value
	return 0
}

func (m *TreeMap[K, V]) Get(key K) (value V, ok bool) {
	return getRecursive(m.root, key, m.ordering)
}

func (m *TreeMap[K, V]) Has(key K) bool {
	_, ok := getRecursive(m.root, key, m.ordering)
	return ok
}

func getRecursive[K, V any](root *treeMapEntry[K, V], key K, before compare.Ordering[K]) (value V, ok bool) {
	if root == nil {
		return
	}
	if before(key, root.key) {
		return getRecursive(root.left, key, before)
	}
	if before(root.key, key) {
		return getRecursive(root.right, key, before)
	}
	return root.value, true
}

func (m *TreeMap[K, V]) Delete(key K) {
	if m.Has(key) {
		m.size -= deleteRecursive(&m.root, key, m.ordering)
	}
}

func deleteRecursive[K, V any](root **treeMapEntry[K, V], key K, before compare.Ordering[K]) int {
	if *root == nil {
		return 0
	}
	if before(key, (*root).key) {
		return deleteRecursive(&(*root).left, key, before)

	}
	if before((*root).key, key) {
		return deleteRecursive(&(*root).right, key, before)

	}
	if (*root).left == nil {
		*root = (*root).right
	} else {
		t := &(*root).left
		for (*t).right != nil {
			t = &(*t).right
		}
		(*root).key = (*t).key
		(*root).value = (*t).value
		*t = (*t).left
	}
	return 1
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

func itRecursive[K, V any](root *treeMapEntry[K, V], it chan<- Entry[K, V]) {
	if root == nil {
		return
	}
	itRecursive(root.left, it)
	it <- root
	itRecursive(root.right, it)
}
