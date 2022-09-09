package pipe

type bufNode[T any] struct {
	value T
	next  *bufNode[T]
}

type listener[T any] struct {
	outCh      chan<- T
	cancelCh   chan struct{}
	buf        *bufNode[T]
	newBuf     **bufNode[T]
	prev, next *listener[T]
}

func (e *listener[T]) finalize() {
	if e.outCh != nil {
		close(e.outCh)
		e.outCh = nil
	}
	e.newBuf = nil
	e.buf = nil
	e.cancelCh = nil
}

func (e *listener[T]) curItem() T {
	if e.buf == nil {
		e.buf = *e.newBuf
	}
	return e.buf.value
}

func (e *listener[T]) advanceItem() (starved bool) {
	e.buf = e.buf.next
	return e.buf == nil
}

type listenerList[T any] struct {
	len  int
	root listener[T]
}

func (l *listenerList[T]) init() {
	root := &l.root
	root.prev, root.next = root, root
}

func (l *listenerList[T]) append(e *listener[T]) {
	root := &l.root
	oldBack := root.prev
	oldBack.next, e.prev = e, oldBack
	root.prev, e.next = e, root
	l.len++
}

func (l *listenerList[T]) drop(e *listener[T]) {
	if e.prev == nil {
		return
	}
	prev, next := e.prev, e.next
	prev.next, next.prev = next, prev
	e.prev, e.next = nil, nil
	l.len--
}

func (l *listenerList[T]) spliceTo(l2 *listenerList[T]) {
	if l.len == 0 {
		return
	}
	root, root2 := &l.root, &l2.root
	front, back, back2 := root.next, root.prev, root2.prev
	back2.next = front
	front.prev = back2
	root2.prev = back
	back.next = root2
	l2.len += l.len
	l.len = 0
	root.prev, root.next = root, root
}
