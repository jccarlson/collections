package kvmap

import "iter"

type builtInLinkedEntry[K comparable, V any] struct {
	key   *K
	value *V

	next, prev *builtInLinkedEntry[K, V]
}

func (e *builtInLinkedEntry[K, V]) Key() K {
	return *e.key
}

func (e *builtInLinkedEntry[K, V]) Value() V {
	return *e.value
}

func (e *builtInLinkedEntry[K, V]) SetValue(v V) {
	*(e.value) = v
}

type BuiltInLinkedHashMap[K comparable, V any] struct {
	m          map[K]*builtInLinkedEntry[K, V]
	head, tail *builtInLinkedEntry[K, V]
}

func NewBuiltInLinkedHashMap[K comparable, V any](opts ...Option) *BuiltInLinkedHashMap[K, V] {
	o := initMapWrapperOptions(opts)
	if o.capacity >= 0 {
		return &BuiltInLinkedHashMap[K, V]{
			m: make(map[K]*builtInLinkedEntry[K, V], o.capacity),
		}
	}
	return &BuiltInLinkedHashMap[K, V]{
		m: make(map[K]*builtInLinkedEntry[K, V]),
	}
}

// Delete implements Interface.
func (b *BuiltInLinkedHashMap[K, V]) Delete(key K) {
	e := b.m[key]
	if e == nil {
		return
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		b.tail = e.prev
	}
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		b.head = e.next
	}
	delete(b.m, key)
}

// Get implements Interface.
func (b *BuiltInLinkedHashMap[K, V]) Get(key K) (val V, ok bool) {
	e, ok := b.m[key]
	if e == nil {
		return
	}
	return *e.value, true
}

// Has implements Interface.
func (b *BuiltInLinkedHashMap[K, V]) Has(key K) bool {
	e := b.m[key]
	return e != nil
}

// Len implements Interface.
func (b *BuiltInLinkedHashMap[K, V]) Len() int {
	return len(b.m)
}

// Put implements Interface.
func (b *BuiltInLinkedHashMap[K, V]) Put(key K, val V) {
	b.Delete(key)
	e := &builtInLinkedEntry[K, V]{
		key:   &key,
		value: &val,
		prev:  b.tail,
	}

	if b.tail != nil {
		b.tail.next = e
	} else {
		b.head = e
	}
	b.tail = e
	b.m[key] = e
}

type builtInLinkedEntryIterator[K comparable, V any] struct {
	current *builtInLinkedEntry[K, V]
}

func (i *builtInLinkedEntryIterator[K, V]) next() (entry Entry[K, V], ok bool) {
	if i.current == nil {
		return
	}
	entry, ok = i.current, true
	i.current = i.current.next
	return
}

func (i *builtInLinkedEntryIterator[K, V]) prev() (entry Entry[K, V], ok bool) {
	if i.current == nil {
		return
	}
	entry, ok = i.current, true
	i.current = i.current.prev
	return
}

func (m *BuiltInLinkedHashMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		it := &builtInLinkedEntryIterator[K, V]{m.head}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e.Key(), e.Value()) {
				return
			}
		}
	}
}

func (m *BuiltInLinkedHashMap[K, V]) Backwards() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		it := &builtInLinkedEntryIterator[K, V]{m.tail}
		for e, ok := it.prev(); ok; e, ok = it.prev() {
			if !yield(e.Key(), e.Value()) {
				return
			}
		}
	}
}

func (m *BuiltInLinkedHashMap[K, V]) Entries() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		it := &builtInLinkedEntryIterator[K, V]{m.head}
		for e, ok := it.next(); ok; e, ok = it.next() {
			if !yield(e) {
				return
			}
		}
	}
}

func (m *BuiltInLinkedHashMap[K, V]) EntriesBackwards() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		it := &builtInLinkedEntryIterator[K, V]{m.tail}
		for e, ok := it.prev(); ok; e, ok = it.prev() {
			if !yield(e) {
				return
			}
		}
	}
}

func (m *BuiltInLinkedHashMap[K, V]) String() string {
	return IterableMapToString(m)
}

func (m *BuiltInLinkedHashMap[K, V]) GoString() string {
	return IterableMapToGoString(m)
}
