package kvmap

import (
	"fmt"
	"strings"

	"github.org/jccarlson/collections"
)

// Interface is the interface common to all key-value maps in package kvmap.
// Users can implement this Interface so their types can use the provided
// utility functions.
type Interface[K, V any] interface {
	Put(K, V)
	Get(K) (V, bool)
	Delete(K)
	Has(K) bool
	Len() int
}

// Entry is the interface wrapping the key-value pairs of a map.
type Entry[K, V any] interface {
	Key() K
	Value() V
	SetValue(V)
}

// IterableMap is the interface wrapping a type that is both a map
// (implementing Interface) and Iterable over its entries.
type IterableMap[K, V any] interface {
	Interface[K, V]
	Iterator() collections.Iterator[Entry[K, V]]
}

type kvMapOpts struct {
	capacity   int
	loadFactor float32
}

// Option is an interface which wraps an adjustable parameter for a map at
// creation. An Option should only be created via one of the functions below.
type Option interface {
	setOpt(*kvMapOpts)
	String() string
}

type capOpt int

func (o capOpt) setOpt(opts *kvMapOpts) {
	opts.capacity = int(o)
}

func (o capOpt) String() string { return fmt.Sprintf("Capacity(%v)", int(o)) }

// Returns an Option which sets the desired initial capacity of the map. Note
// that it is only guaranteed that the capacity will be greater than or equal
// to n.
func Capacity(n int) Option {
	if n < 0 {
		panic("Capacity must be >= 0")
	}
	return capOpt(n)
}

type loadFactorOpt float32

func (o loadFactorOpt) setOpt(opts *kvMapOpts) {
	opts.loadFactor = float32(o)
}

func (o loadFactorOpt) String() string { return fmt.Sprintf("LoadFactor(%v)", float32(o)) }

// Returns an Option which sets the desired load factor of the map. The load
// factor must be in the range (0, 1].
func LoadFactor(loadFactor float32) Option {
	if loadFactor <= 0 || loadFactor > 1 {
		panic(fmt.Sprintf("load factor %f out of range (0.0, 1.0]", loadFactor))
	}
	return loadFactorOpt(loadFactor)
}

// ForEach calls f(key, value) for each key-value pair in m.
func ForEach[K, V any](m IterableMap[K, V], f func(key K, val V)) {
	it := m.Iterator()
	for e, ok := it.Next(); ok; e, ok = it.Next() {
		f(e.Key(), e.Value())
	}
}

// Prints the provided IterableMap to a string. Can be used to easily implement
// the String() method for IterableMap types.
func IterableMapToString[K, V any](m IterableMap[K, V]) string {
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

// Prints the provided IterableMap with type information to a string. Can be
// used to easily implement the GoString() method for IterableMap types.
func IterableMapToGoString[K, V any](m IterableMap[K, V]) string {
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
