package kvmap

import (
	"hash/maphash"
	"sync/atomic"
	"unsafe"

	"github.org/jccarlson/collections/compare"
)

type atomicPointer[T any] struct {
	p unsafe.Pointer
}

func (p *atomicPointer[T]) Load() *T {
	return (*T)(atomic.LoadPointer(&p.p))
}

func (p *atomicPointer[T]) Store(val *T) {
	atomic.StorePointer(&p.p, unsafe.Pointer(val))
}

func (p *atomicPointer[T]) Swap(val *T) (old *T) {
	return (*T)(atomic.SwapPointer(&p.p, unsafe.Pointer(val)))
}

func (p *atomicPointer[T]) CompareAndSwap(old, new *T) (swapped bool) {
	return atomic.CompareAndSwapPointer(&p.p, unsafe.Pointer(old), unsafe.Pointer(new))
}

// lockFreeHashMapEntry is a struct wrapping a Key-Value pair in a LockFreeHashMap.
type lockFreeHashMapEntry[K Hashable, V any] struct {
	key   K
	value V
}

// LockFreeHashMap is a mutex-free hash map for concurrent use by multiple go
// routines.
type LockFreeHashMap[K Hashable, V any] struct {
	comparator compare.Comparator[K]

	seed      maphash.Seed
	entries   []atomicPointer[lockFreeHashMapEntry[K, V]]
	capIdx    int
	size      int
	tombstone *lockFreeHashMapEntry[K, V]
}

func (m *LockFreeHashMap[K, V]) emplace(entry *lockFreeHashMapEntry[K, V]) {
	capacity := int(5)
	hashf := &maphash.Hash{}
	hashf.SetSeed(m.seed)
	for hIdx := int(hash(hashf, entry.key)) % capacity; ; hIdx = (hIdx + 1) % capacity {
		currEntry := m.entries[hIdx].Load()
		if currEntry == nil || (currEntry != m.tombstone && m.comparator(currEntry.key, entry.key)) {
			if m.entries[hIdx].CompareAndSwap(currEntry, entry) {
				break
			}
		}
	}
}

func (m *LockFreeHashMap[K, V]) Put(key K, value V) {

}
