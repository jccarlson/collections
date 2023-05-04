package internal

import (
	"runtime"
)

// ChanIterator is a wrapper type around a receiving channel which implements
// the Iterator interface. A ChanIterator is itself Iterable, and calls to
// Iterator() return itself.
type ChanIterator[V any] struct {
	c    <-chan V
	done chan struct{}
}

func (i *ChanIterator[V]) Close() {
	select {
	case <-i.done:
		// Already closed, Do nothing.
		return
	default:
		close(i.done)
	}
}

func finalize[V any](i *ChanIterator[V]) {
	i.Close()
}

func (i *ChanIterator[V]) Next() (val V, ok bool) {
	val, ok = <-i.c
	return
}

// ChanIteratorSender is created as a pair to a ChanIterator and sends values to ChanIterator.
type ChanIteratorSender[V any] struct {
	c    chan<- V
	done <-chan struct{}
}

func (i *ChanIteratorSender[V]) Send(val V) (ok bool) {
	select {
	case i.c <- val:
		return true
	case <-i.done:
		return false
	}
}

func (i *ChanIteratorSender[V]) Close() {
	close(i.c)
}

func NewChanIteratorPair[V any]() (*ChanIteratorSender[V], *ChanIterator[V]) {
	c, done := make(chan V), make(chan struct{})
	sender := &ChanIteratorSender[V]{c: c, done: done}
	it := &ChanIterator[V]{c: c, done: done}
	runtime.SetFinalizer(it, finalize[V])
	return sender, it
}
