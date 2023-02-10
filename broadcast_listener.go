package pipe

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type bufNode[T any] struct {
	value T
	next  *bufNode[T]
}

type listener[T any] struct {
	outCh      chan<- T
	cancelCh   chan struct{}
	buf        *bufNode[T]
	newBuf     *atomic.Pointer[bufNode[T]]
	prev, next *listener[T]
}

type untypedListener = listener[struct{}]

var listenerPool = sync.Pool{
	New: func() any { return &untypedListener{} },
}

func newListener[T any]() *listener[T] {
	l := listenerPool.Get().(*untypedListener)
	return (*listener[T])(unsafe.Pointer(l))
}

func (e *listener[T]) finalize() {
	if e.outCh != nil {
		close(e.outCh)
		e.outCh = nil
	}
	e.newBuf = nil
	e.buf = nil
	e.cancelCh = nil
	listenerPool.Put((*untypedListener)(unsafe.Pointer(e)))
}

func (e *listener[T]) curItem() T {
	if e.buf == nil {
		e.buf = e.newBuf.Load()
	}
	return e.buf.value
}

func (e *listener[T]) advanceItem() (starved bool) {
	e.buf = e.buf.next
	return e.buf == nil
}

type listenerList[T any] struct {
	root *listener[T]
}

func (l *listenerList[T]) init() {
	l.root = nil
}

func (l *listenerList[T]) append(e *listener[T]) {
	root := l.root
	if root == nil {
		l.root = e
		e.next = e
		e.prev = e
	} else {
		oldBack := root.prev
		oldBack.next, root.prev = e, e
		e.prev, e.next = oldBack, root
	}
}

func (l *listenerList[T]) drop(e *listener[T]) {
	if e.prev == nil {
		return
	}
	prev, next := e.prev, e.next
	if prev == next && next == e {
		l.root = nil
	} else {
		prev.next, next.prev = next, prev
		if e == l.root {
			l.root = next
		}
	}
	e.prev, e.next = nil, nil
}

func (l *listenerList[T]) spliceTo(l2 *listenerList[T]) {
	if l.root == nil {
		return
	}
	if l2.root == nil {
		l2.root = l.root
		l.root = nil
		return
	}
	front, front2 := l.root, l2.root
	back, back2 := front.prev, front2.prev
	front.prev = back2
	back2.next = front
	front2.prev = back
	back.next = front2

	l.root = nil
}

func (l *listenerList[T]) isEmpty() bool { return l.root == nil }
