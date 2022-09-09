package pipe

import (
	"sync"
)

func noop() {}

// broadcaster[T] multiplexes values from inCh to several listeners.
// It gaurantees that sending to inCh won't block for long.
type broadcaster[T any] struct {
	// the channel to multiplex
	inCh <-chan T
	// a channel signaling inCh was closed or broadcaster was detached
	diedCh chan struct{}
	// a channel for manipulating listeners.
	listenerCh chan *listener[T]
	// activeList holds listeners that have remaining values to flush out
	activeList listenerList[T]
	// starvedList holds listeners that is waiting for new values
	starvedList listenerList[T]
	// buf stores previous received value
	buf *bufNode[T]
	// memorized indicates whether send previous received value
	// to newly registered listener
	memorized bool
}

func (b *broadcaster[T]) init(in <-chan T, initial *T) {
	b.inCh = in
	b.diedCh = make(chan struct{})
	b.listenerCh = make(chan *listener[T])
	b.activeList.init()
	b.starvedList.init()
	if initial == nil {
		b.buf = nil
	} else {
		b.buf = &bufNode[T]{value: *initial}
	}
	b.memorized = initial != nil
	go b.loop()
}

func (b *broadcaster[T]) doSelect(isCleaning bool) (isDied bool) {
	reply := make(chan selectResult[T])
	var listener *listener[T]
	var value T
	var ok bool
	var recvEntry bool
	var recvValue bool
	var nWaiting = 0

	barrier := make(chan struct{})
	if !isCleaning {
		nWaiting++
		selectorPond.tasks <- func() {
			select {
			case listener = <-b.listenerCh:
				recvEntry = true
			case value, ok = <-b.inCh:
				recvValue = true
			case <-barrier:
			}
			reply <- selectResult[T]{}
		}
	}
	n := b.activeList.len
	head := b.activeList.root.next
	for n > 0 {
		switch {
		case n >= 8:
			h := head
			head = head.next.next.next.next.next.next.next.next
			selectorPond.tasks <- func() {
				select8(h, reply, barrier)
			}
			nWaiting++
			n -= 8
		case n >= 4:
			h := head
			head = head.next.next.next.next
			selectorPond.tasks <- func() {
				select4(h, reply, barrier)
			}
			nWaiting++
			n -= 4
		case n >= 2:
			h := head
			head = head.next.next
			selectorPond.tasks <- func() {
				select2(h, reply, barrier)
			}
			nWaiting++
			n -= 2
		case n >= 1:
			h := head
			head = head.next
			selectorPond.tasks <- func() {
				select1(h, reply, barrier)
			}
			nWaiting++
			n -= 1
		}
	}
	notResolvedYet := true
	for nWaiting > 0 {
		var r selectResult[T]
		if notResolvedYet {
			r = <-reply
			notResolvedYet = false
			goto HANDLE_REPLY
		}
		select {
		case barrier <- struct{}{}:
		case r = <-reply:
			goto HANDLE_REPLY
		}
		continue
	HANDLE_REPLY:
		switch {
		case r.dead != nil:
			r.dead.finalize()
			b.activeList.drop(r.dead)
		case r.starved != nil:
			if isCleaning {
				r.starved.finalize()
				b.activeList.drop(r.starved)
			} else {
				b.activeList.drop(r.starved)
				b.starvedList.append(r.starved)
			}
		}
		nWaiting--
		continue
	}
	switch {
	case recvEntry:
		if listener == nil {
			return true
		}
		if listener.outCh == nil {
			return false
		}
		select {
		case <-listener.cancelCh:
			if listener.newBuf != nil {
				// listener was registered on one of the two lists
				if listener.buf != nil {
					b.activeList.drop(listener)
				} else {
					b.starvedList.drop(listener)
				}
				listener.finalize()
			} else {
				// listener was canceled before registered
				listener.finalize()
			}
			return false
		default:
		}
		if b.memorized {
			listener.buf = b.buf
		}
		listener.newBuf = &b.buf
		if listener.buf == nil {
			b.starvedList.append(listener)
		} else {
			b.activeList.append(listener)
		}
	case recvValue:
		if !ok {
			return true
		}
		node := &bufNode[T]{value: value}
		if b.buf != nil {
			b.buf.next = node
		}
		b.buf = node
		b.starvedList.spliceTo(&b.activeList)
	}
	return false
}

func (b *broadcaster[T]) loop() {
	for {
		if died := b.doSelect(false); died {
			break
		}
	}
	close(b.diedCh)
	for h := b.starvedList.root.next; h != &b.starvedList.root; h = h.next {
		close(h.outCh)
	}
	b.starvedList.init()
	for b.activeList.len != 0 {
		b.doSelect(true)
	}
}

func (b *broadcaster[T]) current() T { return b.buf.value }

// Detach prematurely detaches the broadcaster from the upstream channel.
// No more values from upstream channel would be broadcasted, and no more new listeners
// should be registered.
func (b *broadcaster[T]) Detach() {
	select {
	case <-b.diedCh:
	case b.listenerCh <- nil:
	}
}

// Bind registers out as a new listener, which receives subsequent values from the upstream channel.
// If the input channel closed or the broadcaster detached, out will be closed immediately.
// A canceller is returned for canceling the subscription. When called, out will be
// unregistered and closed.
func (b *broadcaster[T]) Bind(out chan<- T) func() {
	var once sync.Once
	cancelCh := make(chan struct{})
	entry := &listener[T]{outCh: out, cancelCh: cancelCh}
	select {
	case <-b.diedCh:
		return noop
	case b.listenerCh <- entry:
	}
	return func() {
		once.Do(func() {
			close(cancelCh)
			select {
			case b.listenerCh <- entry:
			case <-b.diedCh:
			}
		})
	}
}

// Listen creates a new output channel and registers it as a new listener.
// The output channel and corresponding canceller is returned.
func (b *broadcaster[T]) Listen() (<-chan T, func()) {
	out := make(chan T)
	return out, b.Bind(out)
}
