package pipe

import (
	"context"

	"golang.org/x/exp/slices"
)

// A listenable object that one can bind listeners to.
type Listenable[T any] interface {
	Bind(out chan<- T) func()
	Listen() (out <-chan T, cancel func())
}

// A listenable object that also memorizes the latest value.
type ListenableM[T any] interface {
	Listenable[T]
	Current() T
}

// A listenable object with comparable element type.
// This allows additional methods Until, UntilCh and UntilContext to be called.
type ListenableC[T comparable] interface {
	Listenable[T]
	Until(...T)
	UntilCh(...T) (<-chan struct{}, func())
	UntilContext(context.Context, ...T)
}

// A listenable object with comparable element type and memorizes the latest value.
type ListenableCM[T comparable] interface {
	ListenableC[T]
	Current() T
}

// Until blocks until one of the conditions satisfies:
// 1) one of the value from b shows up in targets;
// 2) b does not accept new listeners (either b is detached or upstream channel closed).
func Until[T comparable, P Listenable[T]](b P, targets ...T) {
	out, cancel := b.Listen()
	for x := range out {
		if slices.Contains(targets, x) {
			cancel()
			return
		}
	}
}

// UntilCh is the asynchronous version of Until.
// The returned signalCh will be closed when one of the conditions satisfies:
// 1) one of the value from b shows up in targets;
// 2) b does not accept new listeners (either b is detached or upstream channel closed);
// 3) canceller is called.
func UntilCh[T comparable, P Listenable[T]](b P, targets ...T) (signalCh <-chan struct{}, canceller func()) {
	out, cancel := b.Listen()
	signal := make(chan struct{})
	go func() {
		defer close(signal)
		defer cancel()
		for x := range out {
			if slices.Contains(targets, x) {
				return
			}
		}
	}()
	return signal, cancel
}

// UntilContext blocks until one of the conditions satisfies:
// 1) one of the value from b shows up in targets;
// 2) b does not accept new listeners (either b is detached or upstream channel closed);
// 3) ctx is canceled.
func UntilContext[T comparable, P Listenable[T]](ctx context.Context, b P, targets ...T) {
	out, cancel := b.Listen()
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case x, ok := <-out:
			if !ok || slices.Contains(targets, x) {
				return
			}
		}
	}
}

type Broadcaster[T any] struct{ broadcaster[T] }

// Broadcast returns a Broadcaster that pipes values from upstream channel into listeners.
// Broadcaster gaurantees upstream <- val from outside will NOT block, but if it's detached
// prematurely, upstream <- val will block again.
func Broadcast[T any](upstream <-chan T) *Broadcaster[T] {
	b := new(Broadcaster[T])
	b.init(upstream, nil)
	return b
}

type BroadcasterM[T any] struct{ broadcaster[T] }

// BroadcastM returns a broadcaster that memorizes the latest value from upstream.
// Newly registered listener will be firstly fed with the memorized latest value, then subsequent values from upstream.
// If no value coming out of upstream yet, initial is fed.
// The latest value is stored by value (instead of by reference).
func BroadcastM[T any](upstream <-chan T, initial T) *BroadcasterM[T] {
	b := new(BroadcasterM[T])
	b.init(upstream, &initial)
	return b
}

// Current returns the latest value that the broadcaster memorizes.
func (b *BroadcasterM[T]) Current() T { return b.current() }

type broadcasterc[T comparable] struct{ broadcaster[T] }

// Shorthand for Until(b, targets...)
func (b *broadcasterc[T]) Until(targets ...T) {
	Until(b, targets...)
}

// Shorthand for UntilCh(b, targets...)
func (b *broadcasterc[T]) UntilCh(targets ...T) (<-chan struct{}, func()) {
	return UntilCh(b, targets...)
}

// Shorthand for UntilContext(ctx, b, targets...)
func (b *broadcasterc[T]) UntilContext(ctx context.Context, targets ...T) {
	UntilContext(ctx, b, targets...)
}

type BroadcasterC[T comparable] struct{ broadcasterc[T] }

// BroadcastC returns a broadcaster with a comparable type T as element type.
// This allows methods like b.Until(targets...) to be called instead of Until(b, targets...),
// which helps auto type inference and sometimes saves the typing of type variables.
func BroadcastC[T comparable](in <-chan T) *BroadcasterC[T] {
	b := new(BroadcasterC[T])
	b.init(in, nil)
	return b
}

type BroadcasterCM[T comparable] struct{ broadcasterc[T] }

// BroadcastCM returns a broadcaster with comparable element type and also
// is able to memorize the latest value.
func BroadcastCM[T comparable](in <-chan T, initial T) *BroadcasterCM[T] {
	b := new(BroadcasterCM[T])
	b.init(in, &initial)
	return b
}

// Current returns the latest value that the broadcaster memorizes.
func (b *BroadcasterCM[T]) Current() T { return b.current() }
