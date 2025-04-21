package collections

import (
	"fmt"
	"iter"
)

const minSize = 16

// Deque is a double-ended queue, which can function as both a stack and a
// queue. Elements can be added and removed from the head and tail in O(1)
// time.
//
// Stack operations: Push, Pop, Peek
// Queue operations: Enqueue, Dequeue, Peek
// General operations: AddFirst, AddLast, RemoveFirst, RemoveLast, ElementAt,
// Size, All, Backwards
type Deque[E any] struct {
	elems      []E
	head, tail int
}

func (d *Deque[E]) maybeGrow() {
	l := len(d.elems)
	if l == 0 {
		d.elems = make([]E, minSize)
		d.head, d.tail = 0, 0
		return
	}

	if d.tail-d.head == l {
		old := d.elems
		d.elems = make([]E, l<<1)
		copy(d.elems[copy(d.elems, old[d.head:]):], old[:d.head])
		d.head, d.tail = 0, l
	}
}

// AddLast adds an element to the tail of the Deque.
func (d *Deque[E]) AddLast(elem E) {
	d.maybeGrow()
	d.elems[d.tail&(len(d.elems)-1)] = elem
	d.tail++
}

// RemoveLast removes the tail element of the Deque. It returns an error if the
// Deque is empty.
func (d *Deque[E]) RemoveLast() (elem E, err error) {
	if d.tail == d.head {
		return elem, fmt.Errorf("empty Deque")
	}
	d.tail--
	elem = d.elems[d.tail&(len(d.elems)-1)]
	return elem, nil
}

// AddFirst adds an element to the head of the Deque.
func (d *Deque[E]) AddFirst(elem E) {
	d.maybeGrow()
	d.head--
	if d.head < 0 {
		l := len(d.elems)
		d.head += l
		d.tail += l
	}
	d.elems[d.head] = elem
}

// RemoveFirst removes the head element of the Deque. It returns an error if
// the Deque is empty.
func (d *Deque[E]) RemoveFirst() (elem E, err error) {
	if d.tail == d.head {
		return elem, fmt.Errorf("empty Deque")
	}
	elem = d.elems[d.head]
	d.head++
	if l := len(d.elems); d.head >= l {

		d.head -= l
		d.tail -= l
	}
	return elem, nil
}

// Peek returns, but does not remove, the head element of the Deque. It returns
// an error if the Deque is empty.
func (d *Deque[E]) Peek() (elem E, err error) {
	if d.tail == d.head {
		return elem, fmt.Errorf("empty Deque")
	}
	return d.elems[d.head], nil
}

// PeekLast returns, but does not remove, the tail element of the Deque. It
// returns an error if the Deque is empty.
func (d *Deque[E]) PeekLast() (elem E, err error) {
	if d.tail == d.head {
		return elem, fmt.Errorf("empty Deque")
	}
	return d.elems[(d.tail-1)&(len(d.elems)-1)], nil
}

// Enqueue adds an element to the Deque when used as a Queue. It is an alias
// for AddLast.
func (d *Deque[E]) Enqueue(elem E) {
	d.AddLast(elem)
}

// Dequeue removes an element from the Deque when used as a Queue. It returns
// an error if the Deque is empty. It is an alias for RemoveFirst.
func (d *Deque[E]) Dequeue() (elem E, err error) {
	return d.RemoveFirst()
}

// Push adds an element to the Deque when used as a Stack. It is an alias for
// AddFirst
func (d *Deque[E]) Push(elem E) {
	d.AddFirst(elem)
}

// Pop removes an element from the Deque when used as a Stack. It returns an
// error if the Deque is empty. It is an alias for RemoveFirst
func (d *Deque[E]) Pop() (elem E, err error) {
	return d.RemoveFirst()
}

// Size returns the number of elements in the Deque.
func (d *Deque[E]) Size() int {
	return d.tail - d.head
}

// ElementAt returns the i'th element of the Deque, starting from the head
// element at 0.
func (d *Deque[E]) ElementAt(i int) (elem E, err error) {
	if i < 0 || i >= d.Size() {
		return elem, fmt.Errorf("index out of bounds: %d", i)
	}
	return d.elems[(d.head+i)&(len(d.elems)-1)], nil
}

// All returns an iterator over the elements of the Deque, in order from head
// to tail.
func (d *Deque[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		for i := d.head; i < d.tail; i++ {
			if !yield(d.elems[i&(len(d.elems)-1)]) {
				return
			}
		}
	}
}

// Backwards returns an iterator over the elements of the Deque, in reverse
// order from tail to head.
func (d *Deque[E]) Backwards() iter.Seq[E] {
	return func(yield func(E) bool) {
		for i := d.tail - 1; i >= d.head; i-- {
			if !yield(d.elems[i&(len(d.elems)-1)]) {
				return
			}
		}
	}
}
