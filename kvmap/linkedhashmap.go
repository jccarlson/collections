package kvmap

import (
	"fmt"
	"math"

	"github.org/jccarlson/collections"
	"github.org/jccarlson/collections/compare"
)

// linkedHashMapEntry is a struct wrapping a Key-Value pair in a LinkedHashMap.
type linkedHashMapEntry[K any, V any] struct {
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

func (e *linkedHashMapEntry[K, V]) SetValue(v V) {
	*(e.value) = v
}

func initLinkedHashMapOptions(opts []Option) kvMapOpts {
	r := kvMapOpts{
		capacity:   defaultCap,
		loadFactor: defaultLoadFactor,
	}

	for _, opt := range opts {
		opt.setOpt(&r)
	}

	// Round capacity up to a power of 2 (otherwise quadratic probing fails),
	// with a min cap of 8.
	n := r.capacity
	for cap := minCap; cap > 0; cap <<= 1 {
		if cap >= n {
			r.capacity, n = cap, -1
			break
		}
	}
	if n >= 0 {
		panic(fmt.Sprintf("LinkedHashMap initial capacity %d out of range", n))
	}
	return r
}

const minCap = 1 << 3     // 8
const defaultCap = 1 << 5 // 32
const defaultLoadFactor = 0.75

// stepCheckProbabilityAtLoadFactor is the probability that adding an entry
// to the table will take stepCheck probes when the table is at loadFactor
// capacity.
const stepCheckProbabilityAtLoadFactor = 0.25

// NewComparableLinkedHashMap returns a pointer to a new LinkedHashMap with
// comparable keys, and uses the == operator to compare keys.
func NewComparableLinkedHashMap[K comparable, V any](opts ...Option) *LinkedHashMap[K, V] {
	o := initLinkedHashMapOptions(opts)

	return &LinkedHashMap[K, V]{
		comparator: compare.Equal[K],
		hasher:     ComparableMapHasher[K](),

		loadFactor: o.loadFactor,
		stepCheck:  int(math.Round(math.Log(stepCheckProbabilityAtLoadFactor) / math.Log(float64(o.loadFactor)))),

		cap: o.capacity,
	}
}

// NewHashableKeyLinkedHashMap returns a pointer to a new LinkedHashMap with
// HashableKey keys. This can be used to create maps with non-comparable keys
// or which don't use the == operator for comparison.
func NewHashableKeyLinkedHashMap[K HashableKey[K], V any](opts ...Option) *LinkedHashMap[K, V] {
	o := initLinkedHashMapOptions(opts)
	return &LinkedHashMap[K, V]{
		comparator: compare.EqualableComparator[K],
		hasher:     HashableKeyMapHasher[K](),
		loadFactor: o.loadFactor,
		stepCheck:  int(math.Round(math.Log(stepCheckProbabilityAtLoadFactor) / math.Log(float64(o.loadFactor)))),

		cap: o.capacity,
	}
}

// LinkedHashMap is a hash map which can store keys and values of any type, and
// can iterate over inserted key-value pairs in insertion-order. LinkedHashMap
// supports the Capacity() (default: 32) and the LoadFactor() (default: 0.75)
// Options; other Options will panic.
type LinkedHashMap[K any, V any] struct {
	comparator compare.Comparator[K]
	hasher     MapHasher[K]

	// loadFactor is the desired key density of the hash table before rehashing
	// occurs. Valid values are in the range (0, 1]
	loadFactor float32
	// stepCheck is the number of probes an insertion will make before checking
	// to see if the table should be rehashed.
	stepCheck int

	entries []*linkedHashMapEntry[K, V]

	// size is the number of valid entries (keys with values) in the map.
	size int
	// cap is the maximum number of keys the map can currently hold.
	cap int
	// nkeys is the number of keys (including tombstones) in the map.
	nkeys int

	head, tail *linkedHashMapEntry[K, V]
}

func (m *LinkedHashMap[K, V]) maybeResizeAndRehash() {
	if float32(m.nkeys)/float32(m.cap) >= m.loadFactor {
		// If most of the space is taken by tombstones, keep the same capacity
		// and rehash to clear the tombstones. Otherwise, double the capacity.
		if m.nkeys < m.size*2 {
			if m.cap<<1 < minCap {
				panic("LinkedHashMap capacity out-of-range")
			}
			m.cap <<= 1
		}

		tmpEntries := m.entries
		m.entries = make([]*linkedHashMapEntry[K, V], m.cap)
		m.size, m.nkeys = 0, 0
		for _, e := range tmpEntries {
			if e == nil || e.key == nil || e.value == nil {
				continue
			}
			m.emplace(e, false /*canReplace=*/)
		}
	}
}

func (m *LinkedHashMap[K, V]) emplace(entry *linkedHashMapEntry[K, V], canReplace bool) {
	if m.cap == m.nkeys {
		m.maybeResizeAndRehash()
	}

	capMask := m.cap - 1
	step := 0

	for hIdx := int(entry.hashCache) & capMask; ; hIdx = (hIdx + step) & capMask {
		currEntry := m.entries[hIdx]
		if currEntry == nil {
			// We are not replacing any existing entry or tombstone.
			m.entries[hIdx] = entry
			m.size++
			m.nkeys++
			break
		}

		// currEntry is an existing entry or a tombstone. If the keys are equal
		// we will replace it with the new entry, otherwise we have a hash
		// collision and we iterate again. Note that within a call to
		// maybeResizeAndRehash(), this is always a collision, and existing
		// entries are never replaced.
		if canReplace && entry.hashCache == currEntry.hashCache && m.comparator(*currEntry.key, *entry.key) {
			if currEntry.value != nil {
				// currEntry is not a tombstone, so we need to remove it from
				// the linked list.
				if currEntry.prev == nil {
					// currEntry was head.
					m.head = currEntry.next
				} else {
					currEntry.prev.next = currEntry.next
				}
				// currEntry.next cannot be nil because we've already added the
				// replacing element as the tail.
				currEntry.next.prev = currEntry.prev
				m.size--
			}

			m.entries[hIdx] = entry
			m.size++

			// We successfully found a place for the new element, so exit the
			// loop.
			break
		}
		step++
	}
	if step >= m.stepCheck {
		// Lots of collisions; check if rehash is needed.
		m.maybeResizeAndRehash()
	}
}

func (m *LinkedHashMap[K, V]) Put(key K, val V) {
	if m.entries == nil {
		m.entries = make([]*linkedHashMapEntry[K, V], m.cap)
	}
	e := &linkedHashMapEntry[K, V]{key: &key, value: &val, hashCache: m.hasher.Hash(&key), prev: m.tail}
	if m.head == nil {
		m.head = e
	}
	if e.prev != nil {
		e.prev.next = e
	}
	m.tail = e
	m.emplace(e, true /*canReplace=*/)
}

func (m *LinkedHashMap[K, V]) Get(key K) (val V, ok bool) {
	capMask := m.cap - 1
	h := m.hasher.Hash(&key)
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
	h := m.hasher.Hash(&key)
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
	h := m.hasher.Hash(&key)
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
	return IterableMapToString[K, V](m)
}

func (m *LinkedHashMap[K, V]) GoString() string {
	return IterableMapToGoString[K, V](m)
}

func (m *LinkedHashMap[K, V]) Iterator() collections.Iterator[Entry[K, V]] {
	return &linkedHashMapEntryIterator[K, V]{m.head}
}

func (m *LinkedHashMap[K, V]) ReverseIterator() collections.Iterator[Entry[K, V]] {
	return &linkedHashMapEntryReverseIterator[K, V]{m.tail}
}

type linkedHashMapEntryIterator[K, V any] struct {
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

type linkedHashMapEntryReverseIterator[K, V any] linkedHashMapEntryIterator[K, V]

func (i *linkedHashMapEntryReverseIterator[K, V]) Next() (entry Entry[K, V], ok bool) {
	if i.current == nil {
		return
	}
	entry, ok = i.current, true
	i.current = i.current.prev
	return
}
