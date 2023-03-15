package pipe_test

import (
	"testing"
	"time"

	"github.com/hsfzxjy/pipe"
	"github.com/stretchr/testify/assert"
)

func never(t *testing.T, cond func() bool) {
	assert.Never(t, cond,
		100*time.Millisecond,
		10*time.Millisecond)
}

func eventually(t *testing.T, cond func() bool) {
	assert.Eventually(t, cond,
		100*time.Millisecond,
		10*time.Millisecond)
}

func nonblocking(t *testing.T, routine func()) {
	var passed bool
	go func() { routine(); passed = true }()
	eventually(t, func() bool { return passed })
}

func blocking(t *testing.T, routine func()) {
	var passed bool
	go func() { routine(); passed = true }()
	never(t, func() bool { return passed })
}

func closed[T any](ch <-chan T) func() bool {
	return func() bool {
		select {
		case _, ok := <-ch:
			return !ok
		default:
			return false
		}
	}
}

func TestBroadcastNonBlocking(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	defer b.Detach()
	nonblocking(t, func() { ch <- 1 })
}

func TestBroadcastBindAfterSend(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	defer b.Detach()
	ch <- 1
	out, _ := b.Listen()
	blocking(t, func() { <-out })
}

func TestBroadcastBind2(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	defer b.Detach()
	out1, _ := b.Listen()
	out2, _ := b.Listen()
	ch <- 1
	assert.Equal(t, 1, <-out1)
	assert.Equal(t, 1, <-out2)
}

func TestBroadcastBind1to65(t *testing.T) {
	for n := 1; n <= 35; n++ {
		ch := make(chan int)
		b := pipe.Broadcast(ch)
		outs := make([]<-chan int, n)
		for i := 0; i < n; i++ {
			outs[i], _ = b.Listen()
		}
		for i := 0; i < n; i++ {
			ch <- i
		}
		for j := 0; j < n; j++ {
			for i := 0; i < n; i++ {
				x := <-outs[i]
				assert.Equal(t, x, j,
					"n=%d, i=%d, j=%d, x=%d",
					n, i, j, x)
			}
		}
		b.Detach()
	}
}

func TestBroadcastBind2Close(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	out1, _ := b.Listen()
	out2, _ := b.Listen()
	ch <- 1
	assert.Equal(t, 1, <-out1)
	assert.Equal(t, 1, <-out2)
	close(ch)
	eventually(t, closed(out1))
	eventually(t, closed(out2))
}

func TestBroadcastBindCancel(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	out1, cancel1 := b.Listen()
	out2, cancel2 := b.Listen()
	ch <- 1
	cancel1()
	cancel2()
	eventually(t, closed(out1))
	eventually(t, closed(out2))
}

func TestBroadcastBindCancelBeforeClose(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	out1, cancel1 := b.Listen()
	out2, cancel2 := b.Listen()
	ch <- 1
	cancel1()
	cancel2()
	close(ch)
	eventually(t, closed(out1))
	eventually(t, closed(out2))
}

func TestBroadcastBindCancelAfterClose(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	out1, cancel1 := b.Listen()
	out2, cancel2 := b.Listen()
	ch <- 1
	close(ch)
	cancel1()
	cancel2()
	eventually(t, closed(out1))
	eventually(t, closed(out2))
}

func TestBroadcastMBind(t *testing.T) {
	ch := make(chan int)
	b := pipe.BroadcastM(ch, 42)
	out, cancel := b.Listen()
	defer cancel()
	assert.Equal(t, 42, <-out)
	ch <- 1
	assert.Equal(t, 1, <-out)
	out2, cancel2 := b.Listen()
	defer cancel2()
	assert.Equal(t, 1, <-out2)
	close(ch)
}

func TestBroadcastCMUntil(t *testing.T) {
	ch := make(chan int)
	b := pipe.BroadcastCM(ch, 42)
	go func() { ch <- 1 }()
	nonblocking(t, func() { b.Until(1) })
}

func TestBroadcastCMUntilChCancel(t *testing.T) {
	ch := make(chan int)
	b := pipe.BroadcastCM(ch, 42)
	wait, cancel := b.UntilCh(1)
	go cancel()
	nonblocking(t, func() { <-wait })
}
