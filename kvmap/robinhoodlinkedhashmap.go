//go:build robinhoodprobing

package kvmap

import (
	"fmt"
	"iter"
	"math"

	"github.org/jccarlson/collections/compare"
)

// linkedHashMapEntry is a struct wrapping a Key-Value pair in a LinkedHashMap.
type linkedHashMapEntry[K any, V any] struct {
	key   *K
	value *V

	hashCache uint64
	psl       int

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

	// Round capacity up to a power of 2 with a min cap of 8.
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

// logstepCheckProbabilityAtLoadFactor is the log of the probability (0.25) that
// adding an entry to the table will take stepCheck probes when the table is at
// loadFactor capacity.
const logStepCheckProbabilityAtLoadFactor = -1.38629436112

// NewComparableLinkedHashMap returns a pointer to a new LinkedHashMap with
// comparable keys, and uses the == operator to compare keys.
func NewComparableLinkedHashMap[K comparable, V any](opts ...Option) *LinkedHashMap[K, V] {
	o := initLinkedHashMapOptions(opts)
	stepCheck := math.MaxInt
	if o.loadFactor < 1 {
		stepCheck = int(math.Round(logStepCheckProbabilityAtLoadFactor / math.Log(float64(o.loadFactor))))
	}

	return &LinkedHashMap[K, V]{
		comparator: compare.Equal[K],
		hash:       ComparableMapHasher[K](),

		loadFactor: o.loadFactor,
		stepCheck:  stepCheck,

		cap: o.capacity,
	}
}

// NewHashableKeyLinkedHashMap returns a pointer to a new LinkedHashMap with
// HashableKey keys. This can be used to create maps with non-comparable keys
// or which don't use the == operator for comparison.
func NewHashableKeyLinkedHashMap[K HashableKey[K], V any](opts ...Option) *LinkedHashMap[K, V] {
	o := initLinkedHashMapOptions(opts)
	stepCheck := math.MaxInt
	if o.loadFactor < 1 {
		stepCheck = int(math.Round(logStepCheckProbabilityAtLoadFactor / math.Log(float64(o.loadFactor))))
	}
	return &LinkedHashMap[K, V]{
		comparator: compare.EqualableComparator[K],
		hash:       HashableKeyMapHasher[K](),

		loadFactor: o.loadFactor,
		stepCheck:  stepCheck,

		cap: o.capacity,
	}
}

// NewCustomLinkedHashMap returns a pointer to a new LinkedHashMap with
// a user-provided Comparator and MapHasher. This can be used for Maps which
// require a hash function other than what is provided in hash/maphash. The
// MapHasher provided should be consistent with the Comparator.
func NewCustomLinkedHashMap[K any, V any](comparator compare.Comparator[K], mapHasher MapHasher[K], opts ...Option) *LinkedHashMap[K, V] {
	o := initLinkedHashMapOptions(opts)
	stepCheck := math.MaxInt
	if o.loadFactor < 1 {
		stepCheck = int(math.Round(logStepCheckProbabilityAtLoadFactor / math.Log(float64(o.loadFactor))))
	}
	return &LinkedHashMap[K, V]{
		comparator: comparator,
		hash:       mapHasher,

		loadFactor: o.loadFactor,
		stepCheck:  stepCheck,

		cap: o.capacity,
	}
}

// LinkedHashMap is a hash map which can store keys and values of any type, and
// can iterate over inserted key-value pairs in insertion-order (and reverse).
// LinkedHashMap supports the Capacity() (default: 32) and the LoadFactor()
// (default: 0.75) Options; other Options are ignored.
type LinkedHashMap[K any, V any] struct {
	comparator compare.Comparator[K]
	hash       MapHasher[K]

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
	e := &linkedHashMapEntry[K, V]{key: &key, value: &val, hashCache: m.hash(&key), prev: m.tail}
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
	h := m.hash(&key)
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
	h := m.hash(&key)
	step := 0
	for hIdx := int(h) & capMask; ; hIdx = (hIdx + step) & capMask {
		currEntry := m.entries[hIdx]
		if currEntry == nil {
			// nothing to delete.
			return
		}
		if h == currEntry.hashCache && m.comparator(*currEntry.key, key) {
			if currEntry.value == nil {
				// tombstone, nothing to delete.
				return
			}
			if currEntry.prev != nil {
				currEntry.prev.next = currEntry.next
			} else {
				// currEntry was head.
				m.head = currEntry.next
			}
			if currEntry.next != nil {
				currEntry.next.prev = currEntry.prev
			} else {
				// currEntry was tail.
				m.tail = currEntry.prev
			}

			// make currEntry a tombstone.
			currEntry.value = nil
			currEntry.next, currEntry.prev = nil, nil
			m.size--
			return
		}
		step++
	}
}

func (m *LinkedHashMap[K, V]) Has(key K) bool {
	capMask := m.cap - 1
	h := m.hash(&key)
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
	return IterableMapToString(m)
}

func (m *LinkedHashMap[K, V]) GoString() string {
	return IterableMapToGoString(m)
}

type linkedHashMapEntryIterator[K, V any] struct {
	current *linkedHashMapEntry[K, V]
}

func (i *linkedHashMapEntryIterator[K, V]) next() (entry Entry[K, V], ok bool) {
	if i.current == nil {
		return
	}
	entry, ok = i.current, true
	i.current = i.current.next
	return
}

func (i *linkedHashMapEntryIterator[K, V]) prev() (entry Entry[K, V], ok bool) {
	if i.current == nil {
		return
	}
	entry, ok = i.current, true
	i.current = i.current.prev
	return
}

func (m *LinkedHashMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		it := &linkedHashMapEntryIterator[K, V]{m.head}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e.Key(), e.Value()) {
				return
			}
		}
	}
}

func (m *LinkedHashMap[K, V]) Backwards() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		it := &linkedHashMapEntryIterator[K, V]{m.tail}
		for e, ok := it.prev(); ok; e, ok = it.prev() {
			if !yield(e.Key(), e.Value()) {
				return
			}
		}
	}
}

func (m *LinkedHashMap[K, V]) Entries() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		it := &linkedHashMapEntryIterator[K, V]{m.head}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e) {
				return
			}
		}
	}
}

func (m *LinkedHashMap[K, V]) EntriesBackwards() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		it := &linkedHashMapEntryIterator[K, V]{m.tail}
		for e, ok := it.prev(); ok; e, ok = it.prev() {
			if !yield(e) {
				return
			}
		}
	}
}
