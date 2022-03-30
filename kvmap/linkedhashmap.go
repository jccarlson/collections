package kvmap

import (
	"hash/maphash"
	"math"

	"github.org/jccarlson/collections"
	"github.org/jccarlson/collections/compare"
)

// linkedHashMapEntry is a struct wrapping a Key-Value pair in a LinkedHashMap.
type linkedHashMapEntry[K Hashable, V any] struct {
	key   *K
	value *V

	hashCache uint64

	prev, next *linkedHashMapEntry[K, V]
}

func (e *linkedHashMapEntry[K, V]) Key() K {
	return *e.key
}

func (e *linkedHashMapEntry[K, V]) Value() V {
	return *e.value
}

var baseTableCap = 1 << 5 // 32

// loadFactor is the desired key density of the hash table before rehashing
// occurs.
const loadFactor = 0.75

// stepCheckProbabilityAtLoadFactor is the probability that adding an entry
// to the table will take stepCheck probes when the table is at loadFactor
// capacity.
const stepCheckProbabilityAtLoadFactor = 0.25

// stepCheck is the number of probes an insertion will make before checking
// to see if the table should be rehashed.
var stepCheck = int(math.Round(math.Log(stepCheckProbabilityAtLoadFactor) / math.Log(loadFactor)))

func NewComparableLinkedHashMap[K interface {
	Hashable
	comparable
}, V any]() *LinkedHashMap[K, V] {
	return &LinkedHashMap[K, V]{
		comparator: compare.Equals[K],
		cap:        baseTableCap,
	}
}

func NewCustomComparatorLinkedHashMap[K Hashable, V any](comparator compare.Comparator[K]) *LinkedHashMap[K, V] {
	return &LinkedHashMap[K, V]{
		comparator: comparator,
		cap:        baseTableCap,
	}
}

func NewEqualerLinkedHashMap[K interface {
	Hashable
	compare.Equater[K]
}, V any]() *LinkedHashMap[K, V] {
	return &LinkedHashMap[K, V]{
		comparator: compare.DefaultComparator[K],
		cap:        baseTableCap,
	}
}

// LinkedHashMap is a quadratic probe based hash map which can store keys
// satisfying the Hashable interface and values of any type, and can iterate
// over inserted key-value pairs in insertion-order.
type LinkedHashMap[K Hashable, V any] struct {
	comparator compare.Comparator[K]

	hash    maphash.Hash
	entries []linkedHashMapEntry[K, V]
	size    int
	cap     int
	nkeys   int

	head, tail *linkedHashMapEntry[K, V]
}

func (m *LinkedHashMap[K, V]) maybeResizeAndRehash() {
	if float64(m.nkeys)/float64(m.cap) > loadFactor {
		// If most of the space is taken by tombstones, keep the same capacity
		// and rehash to clear the tombstones. Otherwise, double the capacity.
		if m.nkeys < m.size*2 {
			if m.cap<<1 < baseTableCap {
				panic("LinkedHashMap capacity out-of-range")
			}
			m.cap <<= 1
		}

		tmpEntries := m.entries
		m.entries = make([]linkedHashMapEntry[K, V], m.cap)
		m.size, m.nkeys = 0, 0
		for _, e := range tmpEntries {
			if e.key == nil || e.value == nil {
				continue
			}
			m.emplace(e)
		}
	}
}

func (m *LinkedHashMap[K, V]) emplace(entry linkedHashMapEntry[K, V]) {
	if m.cap == m.nkeys {
		m.maybeResizeAndRehash()
	}

	capMask := m.cap - 1
	step := 0

	for hIdx := int(entry.hashCache) & capMask; ; hIdx = (hIdx + step) & capMask {
		currEntry := m.entries[hIdx]
		if currEntry.key == nil {
			m.entries[hIdx] = entry
			m.size++
			m.nkeys++
			break
		}
		if entry.hashCache == currEntry.hashCache && m.comparator(*currEntry.key, *entry.key) {
			m.entries[hIdx] = entry
			m.size++
			break
		}
		step++
	}
	if step >= stepCheck {
		// lots of collisions; check if rehash is needed
		m.maybeResizeAndRehash()
	}
}

func (m *LinkedHashMap[K, V]) Put(key K, val V) {
	if m.entries == nil {
		m.entries = make([]linkedHashMapEntry[K, V], m.cap)
	}
	e := linkedHashMapEntry[K, V]{key: &key, value: &val, hashCache: hash(&m.hash, key), prev: m.tail}
	if m.head == nil {
		m.head = &e
	}
	if e.prev != nil {
		e.prev.next = &e
	}
	m.tail = &e
	m.emplace(e)
}

func (m *LinkedHashMap[K, V]) Get(key K) (val V, ok bool) {
	capMask := m.cap - 1
	h := hash(&m.hash, key)
	step := 0
	for hIdx := int(h) & capMask; ; hIdx = (hIdx + step) & capMask {
		currEntry := m.entries[hIdx]
		if currEntry.key == nil {
			return
		}
		if h == currEntry.hashCache && m.comparator(*currEntry.key, key) {
			if currEntry.value == nil {
				return
			}
			return *currEntry.value, true
		}
		step++
	}
}

func (m *LinkedHashMap[K, V]) Delete(key K) {
	capMask := m.cap - 1
	h := hash(&m.hash, key)
	step := 0
	for hIdx := int(h) & capMask; ; hIdx = (hIdx + step) & capMask {
		currEntry := m.entries[hIdx]
		if currEntry.key == nil {
			return
		}
		if h == currEntry.hashCache && m.comparator(*currEntry.key, key) {
			if currEntry.prev != nil {
				currEntry.prev.next = currEntry.next
			}
			if currEntry.next != nil {
				currEntry.next.prev = currEntry.prev
			}
			m.entries[hIdx].value = nil
			m.entries[hIdx].next, m.entries[hIdx].prev = nil, nil
			m.size--
			return
		}
		step++
	}
}

func (m *LinkedHashMap[K, V]) Has(key K) bool {
	capMask := m.cap - 1
	h := hash(&m.hash, key)
	step := 0
	for hIdx := int(h) & capMask; ; hIdx = (hIdx + step) & capMask {
		currEntry := m.entries[hIdx]
		if currEntry.key == nil {
			return false
		}
		if h == currEntry.hashCache && m.comparator(*currEntry.key, key) {
			return currEntry.value != nil
		}
		step++
	}
}

func (m *LinkedHashMap[K, V]) Len() int {
	return m.size
}

func (m *LinkedHashMap[K, V]) String() string {
	return iterableMapToString[K, V](m)
}

func (m *LinkedHashMap[K, V]) GoString() string {
	return iterableMapToGoString[K, V](m)
}

func (m *LinkedHashMap[K, V]) Iterator() collections.Iterator[Entry[K, V]] {
	return &linkedHashMapEntryIterator[K, V]{m.head}
}

type linkedHashMapEntryIterator[K Hashable, V any] struct {
	current *linkedHashMapEntry[K, V]
}

func (i *linkedHashMapEntryIterator[K, V]) Next() (entry Entry[K, V], ok bool) {
	if i.current == nil {
		return
	}
	entry, ok = i.current, true
	i.current = i.current.next
	return
}
