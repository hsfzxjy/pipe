package pipe

import (
	"sync"
	"sync/atomic"
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
	buf atomic.Pointer[bufNode[T]]
	// memorized indicates whether send previous received value
	// to newly registered listener
	memorized bool

	initOnce once
}

func (b *broadcaster[T]) init(in <-chan T, initial *T) {
	b.inCh = in
	b.activeList.init()
	b.starvedList.init()
	if initial == nil {
		b.buf.Store(nil)
	} else {
		b.buf.Store(&bufNode[T]{value: *initial})
	}
	b.memorized = initial != nil
}

func (b *broadcaster[T]) initialized() bool {
	return b.initOnce.Done()
}

func (b *broadcaster[T]) ensureInit() {
	b.initOnce.Do(func() {
		b.diedCh = make(chan struct{})
		b.listenerCh = make(chan *listener[T])
		go b.loop()
	})
}

func (b *broadcaster[T]) doSelect(
	activeBuf, starvedBuf []*listener[T],
	barrier chan struct{},
	reply chan selectResult[T],
	isCleaning bool,
) (isDied bool) {
	type _listenerType = listener[T]
	var listener *_listenerType
	var value T
	var ok bool
	var recvEntry bool
	var recvValue bool
	var nWaiting = 0

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
	head := b.activeList.root
	if head != nil {
		n := 1
		p := head
		activeBuf[0] = p
	ENUMERATE:
		for {
			for n < 8 && p.next != head {
				p = p.next
				activeBuf[n] = p
				n++
			}
			switch {
			case n == 8:
				b := [8]*_listenerType(activeBuf)
				selectorPond.tasks <- func() {
					select8(b, reply, barrier)
				}
				nWaiting++
			default:
				buf := activeBuf
				if n >= 4 {
					b := [4]*_listenerType(activeBuf)
					buf = buf[4:]
					selectorPond.tasks <- func() {
						select4(b, reply, barrier)
					}
					nWaiting++
					n -= 4
				}

				if n >= 2 {
					e0, e1 := buf[0], buf[1]
					buf = buf[2:]
					selectorPond.tasks <- func() {
						select2(e0, e1, reply, barrier)
					}
					nWaiting++
					n -= 2
				}

				if n == 1 {
					e := buf[0]
					selectorPond.tasks <- func() {
						select1(e, reply, barrier)
					}
					nWaiting++
					n -= 1
				}

				break ENUMERATE
			}
			n = 0
		}
	}
	head = b.starvedList.root
	if head != nil {
		n := 1
		p := head
		starvedBuf[0] = p
	ENUMERATE_STARVE:
		for {
			for n < 8 && p.next != head {
				p = p.next
				starvedBuf[n] = p
				n++
			}
			switch {
			case n == 8:
				b := [8]*_listenerType(starvedBuf)
				selectorPond.tasks <- func() {
					selectStarved8(b, reply, barrier)
				}
				nWaiting++
			default:
				buf := starvedBuf
				if n >= 4 {
					b := [4]*_listenerType(starvedBuf)
					buf = buf[4:]
					selectorPond.tasks <- func() {
						selectStarved4(b, reply, barrier)
					}
					nWaiting++
					n -= 4
				}

				if n >= 2 {
					e0, e1 := buf[0], buf[1]
					buf = buf[2:]
					selectorPond.tasks <- func() {
						selectStarved2(e0, e1, reply, barrier)
					}
					nWaiting++
					n -= 2
				}

				if n == 1 {
					e := buf[0]
					selectorPond.tasks <- func() {
						selectStarved1(e, reply, barrier)
					}
					nWaiting++
					n -= 1
				}

				break ENUMERATE_STARVE
			}
			n = 0
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
			if r.starved != nil {
				b.starvedList.drop(r.dead)
			} else {
				b.activeList.drop(r.dead)
			}
			r.dead.finalize()
		case r.starved != nil:
			if isCleaning {
				b.activeList.drop(r.starved)
				r.starved.finalize()
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
			// signal to die
			return true
		}
		if listener.outCh == nil {
			// registering nil chan, noop
			return false
		}
		select {
		case <-listener.cancelCh:
			// listener was canceled before registered
			listener.finalize()
			return false
		default:
		}
		if b.memorized {
			listener.buf = b.buf.Load()
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
		{
			buf := b.buf.Load()
			if buf != nil {
				buf.next = node
			}
			b.buf.Store(node)
		}
		b.starvedList.spliceTo(&b.activeList)
	}
	return false
}

func (b *broadcaster[T]) loop() {
	var activeBuf, starvedBuf [8]*listener[T]
	reply := make(chan selectResult[T])
	barrier := barrierPool.Get().(chan struct{})
	for {
		if died := b.doSelect(
			activeBuf[:], starvedBuf[:],
			barrier, reply,
			false); died {
			break
		}
	}
	close(b.diedCh)
	{
		p := b.starvedList.root
		sentinel := p
		if p == nil {
			goto DONE
		}
	LOOP:
		close(p.outCh)
		p = p.next
		if p != sentinel {
			goto LOOP
		}
	DONE:
	}
	b.starvedList.init()
	for !b.activeList.isEmpty() {
		b.doSelect(
			activeBuf[:], starvedBuf[:],
			barrier, reply,
			true)
	}
	barrierPool.Put(barrier)
}

func (b *broadcaster[T]) current() T {
	buf := b.buf.Load()
	return buf.value
}

func (b *broadcaster[T]) detach() {
	if !b.initialized() {
		return
	}
	select {
	case <-b.diedCh:
	case b.listenerCh <- nil:
	}
}

// Bind registers out as a new listener, which receives subsequent values from the upstream channel.
// If the input channel closed or the broadcaster detached, out will be closed immediately.
// A canceller is returned for canceling the subscription. When called, out will be
// unregistered and closed.
func (b *broadcaster[T]) Bind(out chan<- T) (cancel func()) {
	b.ensureInit()
	var once sync.Once
	cancelCh := make(chan struct{})
	entry := newListener[T]()
	entry.outCh = out
	entry.cancelCh = cancelCh
	select {
	case <-b.diedCh:
		return noop
	case b.listenerCh <- entry:
	}
	return func() {
		once.Do(func() {
			close(cancelCh)
		})
	}
}

// Listen creates a new output channel and registers it as a new listener.
// The output channel and corresponding canceller is returned.
func (b *broadcaster[T]) Listen() (<-chan T, func()) {
	out := make(chan T)
	return out, b.Bind(out)
}
