package kvmap

import (
	"hash"
	"hash/maphash"

	"github.org/jccarlson/collections/compare"
)

// A MapHasher is a function which hashes map keys.
type MapHasher[K any] func(key *K) uint64

// HashableKey is a compare.Equalable with a HashBytes() method. HashBytes()
// should return a byte-slice representation of the wrapped value which is
// consistent with Equals().
//
// Specifically, for two values v1, v2:
// v1.Equals(v2) -> bytes.Equal(v1.HashBytes(), v2.HashBytes())
type HashableKey[T any] interface {
	compare.Equalable[T]

	HashBytes() []byte
}

// HashableKeyMapHasher returns a MapHasher for HashableKey types.
func HashableKeyMapHasher[K HashableKey[K]]() MapHasher[K] {
	seed := maphash.MakeSeed()
	return func(key *K) uint64 {
		return maphash.Bytes(seed, (*key).HashBytes())
	}
}

// BytesMapHasher returns a MapHasher for any key type. Users must provide a
// serialization function that takes a pointer to a key and returns a
// byte-slice representation of the key which is consistent with the comparison
// implementation used.
func BytesMapHasher[K any](toBytes func(*K) []byte) MapHasher[K] {
	seed := maphash.MakeSeed()
	return func(key *K) uint64 {
		return maphash.Bytes(seed, toBytes(key))
	}
}

// MapHasherFromHash64 allows users to easily convert a hash.Hash64 and a key
// serialization function toBytes into a MapHasher.
func MapHasherFromHash64[K any](hash64 hash.Hash64, toBytes func(*K) []byte) MapHasher[K] {
	return func(key *K) uint64 {
		hash64.Reset()
		hash64.Write(toBytes(key))
		return hash64.Sum64()
	}
}
