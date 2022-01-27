package gopherbox

import "github.org/jccarlson/gopherbox/compare"

type Stack[T any] interface {
	Container[T]
	Push(t T)
	Pop() (T, error)
	Peek() (T, error)
}

type sliceStack[T any] struct {
	elements   []T
	comparator compare.Comparator[T]
}

func (s *sliceStack[T]) Has(t T) bool {
	for _, e := range s.elements {
		if s.comparator(t, e) {
			return true
		}
	}
	return false
}
