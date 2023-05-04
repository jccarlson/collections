package collections

import "github.org/jccarlson/collections/internal"

// An Iterator iterates through a sequence of values. Upon creation, users can
// repeatedly call Next() to retrieve the next value in the sequence, until
// ok == false.
type Iterator[V any] interface {
	Next() (val V, ok bool)
}

type closeable interface {
	Close()
}

func maybeClose(v any) {
	vCloseable, ok := v.(closeable)
	if ok {
		vCloseable.Close()
	}
}

// Any returns true if the predicate returns true for any value in values.
func Any[V any](it Iterator[V], predicate func(V) bool) bool {
	if it == nil {
		return false
	}

	for val, ok := it.Next(); ok; val, ok = it.Next() {
		if predicate(val) {
			maybeClose(it)
			return true
		}
	}
	return false
}

// All returns true if the predicate returns true for all values in values.
func All[V any](iterator Iterator[V], predicate func(V) bool) bool {
	if iterator == nil {
		return true
	}

	for val, ok := iterator.Next(); ok; val, ok = iterator.Next() {
		if !predicate(val) {
			maybeClose(iterator)
			return false
		}
	}
	return true
}

// Filter returns an Iterator with only values for which predicate is true.
func Filter[V any](iterator Iterator[V], predicate func(V) bool) Iterator[V] {
	if iterator == nil {
		return nil
	}

	sender, ci := internal.NewChanIteratorPair[V]()

	go func() {
		for val, ok := iterator.Next(); ok; val, ok = iterator.Next() {
			if predicate(val) && !sender.Send(val) {
				break
			}
		}
		sender.Close()
	}()
	return ci
}

// Map consumes values of type V1, transforms them to type V2 via mapper, then
// returns them in order via a new Iterator.
func Map[V1, V2 any](iterator Iterator[V1], mapper func(V1) V2) Iterator[V2] {
	if iterator == nil {
		return nil
	}

	sender, ci := internal.NewChanIteratorPair[V2]()

	go func() {
		for val, ok := iterator.Next(); ok; val, ok = iterator.Next() {
			if !sender.Send(mapper(val)) {
				break
			}
		}
		sender.Close()
	}()
	return ci
}

// Reduce aggregates all values in iterator into a single result of type V2 via
// the reducer function. reducer takes a base value of type V2 and a value of
// type V1 and returns a new base value which represents the aggregation of
// both. It is called repeatedly with each value from iterator and the result
// of the previous call, starting with initial. The result of the last call is
// returned.
func Reduce[V1, V2 any](iterator Iterator[V1], initial V2, reducer func(V2, V1) V2) V2 {
	if iterator == nil {
		return initial
	}

	for val, ok := iterator.Next(); ok; val, ok = iterator.Next() {
		initial = reducer(initial, val)
	}
	return initial
}

// ToSlice collects all values in iterator to a slice.
func ToSlice[V any](iterator Iterator[V]) []V {
	if iterator == nil {
		return nil
	}

	result := []V{}
	for val, ok := iterator.Next(); ok; val, ok = iterator.Next() {
		result = append(result, val)
	}
	return result
}
