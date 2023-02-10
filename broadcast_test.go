package pipe_test

import (
	"testing"
	"time"

	"github.com/hsfzxjy/pipe"
)

func closed[T any](ch <-chan T) {
	for range ch {
	}
}

func TestBroadcastNonBlocking(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	defer b.Detach()
	select {
	case ch <- 1:
	case <-time.After(1 * time.Second):
		t.Error("timeout")
	}
}

func TestBroadcastBindAfterSend(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	defer b.Detach()
	ch <- 1
	out, _ := b.Listen()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expect no value")
		}
	case <-time.After(10 * time.Millisecond):
	}
}

func TestBroadcastBind2(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	defer b.Detach()
	out1, _ := b.Listen()
	out2, _ := b.Listen()
	ch <- 1
	if <-out1 != 1 || <-out2 != 1 {
		t.Fatal()
	}
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
				if x != j {
					t.Fatalf("n=%d: <-outs[%d] != %d, but %d", n, i, j, x)
				}
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
	if <-out1 != 1 || <-out2 != 1 {
		t.Fatal()
	}
	close(ch)
	closed(out1)
	closed(out2)
}

func TestBroadcastBindCancel(t *testing.T) {
	ch := make(chan int)
	b := pipe.Broadcast(ch)
	out1, cancel1 := b.Listen()
	out2, cancel2 := b.Listen()
	ch <- 1
	cancel1()
	cancel2()
	closed(out1)
	closed(out2)
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
	closed(out1)
	closed(out2)
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
	closed(out1)
	closed(out2)
}

func TestBroadcastMBind(t *testing.T) {
	ch := make(chan int)
	b := pipe.BroadcastM(ch, 42)
	out, cancel := b.Listen()
	defer cancel()
	if <-out != 42 {
		t.Fatal()
	}
	ch <- 1
	if <-out != 1 {
		t.Fatal()
	}
	out2, cancel2 := b.Listen()
	defer cancel2()
	if <-out2 != 1 {
		t.Fatal()
	}
	close(ch)
}

func TestBroadcastCMUntil(t *testing.T) {
	ch := make(chan int)
	b := pipe.BroadcastCM(ch, 42)
	go func() { ch <- 1 }()
	b.Until(1)
}

// func TestBroadcastCMUntilCh(t *testing.T) {
// 	ch := make(chan int)
// 	b := pipe.BroadcastCM(ch, 42)
// 	wait, _ := b.UntilCh(1)
// 	ch <- 1
// 	<-wait
// }

func TestBroadcastCMUntilChCancel(t *testing.T) {
	ch := make(chan int)
	b := pipe.BroadcastCM(ch, 42)
	wait, cancel := b.UntilCh(1)
	go cancel()
	<-wait
}
