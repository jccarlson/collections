package ds

import "github.org/jccarlson/collections/compare"

type BinaryHeap[E any] struct {
	tree   []E
	before compare.Ordering[E]
	size   int
}
