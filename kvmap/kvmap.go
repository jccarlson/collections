package kvmap

import (
	"fmt"
	"hash/maphash"
	"strings"

	"github.org/jccarlson/collections"
)

type Interface[K, V any] interface {
	Put(K, V)
	Get(K) (V, bool)
	Delete(K)
	Has(K) bool
	Len() int
}

type Entry[K, V any] interface {
	Key() K
	Value() V
}

type IterableMap[K, V any] interface {
	Interface[K, V]
	collections.Iterable[Entry[K, V]]
}

type Hashable interface {
	HashBytes() []byte
}

type entryChanIterator[K, V any] <-chan Entry[K, V]

func (i entryChanIterator[K, V]) Next() (e Entry[K, V], ok bool) {
	e, ok = <-i
	return
}

func hash[K Hashable](hash *maphash.Hash, key K) uint64 {
	hash.Write(key.HashBytes())
	r := hash.Sum64()
	hash.Reset()
	return r
}

func iterableMapToString[K, V any](m IterableMap[K, V]) string {
	sb := &strings.Builder{}
	sb.WriteString("map[")
	it := m.Iterator()
	e, ok := it.Next()
	eToStr := func(e Entry[K, V]) string {
		return fmt.Sprintf("%v:%v", e.Key(), e.Value())
	}
	if ok {
		sb.WriteString(eToStr(e))
	}
	for e, ok = it.Next(); ok; e, ok = it.Next() {
		sb.WriteRune(' ')
		sb.WriteString(eToStr(e))
	}
	sb.WriteRune(']')
	return sb.String()
}

func iterableMapToGoString[K, V any](m IterableMap[K, V]) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("%T{", m))
	it := m.Iterator()
	e, ok := it.Next()
	eToStr := func(e Entry[K, V]) string {
		return fmt.Sprintf("%#v:%#v", e.Key(), e.Value())
	}
	if ok {
		sb.WriteString(eToStr(e))
	}
	for e, ok = it.Next(); ok; e, ok = it.Next() {
		sb.WriteString(", ")
		sb.WriteString(eToStr(e))
	}
	sb.WriteRune('}')
	return sb.String()
}
