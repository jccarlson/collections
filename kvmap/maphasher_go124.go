//go:build go1.24 && !forcepolyfill

package kvmap

import "hash/maphash"

// ComparableMapHasher returns a MapHasher for comparable keys which is
// consistent with the == operator.
//
// Specifically, with
// hash := ComparableMapHasher[T]()
// for two values v1, v2 of type T:
// v1 == v2 -> hash(v1) == hash(v2)
func ComparableMapHasher[K comparable]() MapHasher[K] {
	seed := maphash.MakeSeed()
	return func(key *K) uint64 {
		return maphash.Comparable(seed, *key)
	}
}
