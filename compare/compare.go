package compare

import "constraints"

// An Ordering returns true if t1 comes strictly before t2.
//
// For any Ordering O, and elements t1, t2, and t3, the following must hold:
//  - O(t1, t1) == false
//  - If O(t1,t2) == true and O(t2, t3) == true then O(t1, t3) == true
//  - If O(t1,t2) == false and O(t2, t3) == false then O(t1, t3) == false
//  - If O(t1, t2) == true then O(t2, t1) == false
//  - If O(t1, t2) == false and O(t2, t1) == false then t1 is equal to t2 for
//    ordering purposes.
type Ordering[T any] func(t1, t2 T) bool

// Less is the standard Ordering for constraints.Ordered types, using the '<'
// operator.
func Less[T constraints.Ordered](t1, t2 T) bool {
	return t1 < t2
}

// Reverse returns the reverse Ordering of o.
func Reverse[T any](o Ordering[T]) Ordering[T] {
	return func(t1, t2 T) bool {
		// Don't use !o(t1, t2), in the case that t1 == t2
		return o(t2, t1)
	}
}

// Orderer is an interface defining an ordering on elements of type T.
// Before(t) returns true if the receiver comes before t.
type Orderer[T any] interface {
	Before(t T) bool
}

// DefaultOrdering is the standard Ordering for Ordered types.
func DefaultOrdering[T Orderer[T]](t1, t2 T) bool {
	return t1.Before(t2)
}

// A Comparator returns true if t1 == t2.
//
// For any Comparator C, and elements t1, t2, and t3, the following must hold:
//  - C(t1, t1) == true
//  - If C(t1,t2) == true then C(t2,t1) == true
//  - If C(t1,t2) == true and C(t2, t3) == true then C(t1, t3) == true
type Comparator[T any] func(t1, t2 T) bool

// Equals is the default Comparator for comparable types.
func Equals[T comparable](t1, t2 T) bool {
	return t1 == t2
}

// Equater is an interface that wraps the Equals() method.
type Equater[T any] interface {
	Equals(t T) bool
}

// DefaultComparator is the default Comparator for Equalers.
func DefaultComparator[T Equater[T]](t1, t2 T) bool {
	return t1.Equals(t2)
}
