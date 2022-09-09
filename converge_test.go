package pipe_test

import (
	"testing"

	"github.com/hsfzxjy/pipe"
)

func TestConverge2(t *testing.T) {
	a := make(chan int)
	b := make(chan string)
	c := pipe.Converge2(a, b)
	go func() {
		a <- 42
		b <- "foo"
		close(a)
		close(b)
	}()
	if <-c != 42 {
		t.Fatal()
	}
	if <-c != "foo" {
		t.Fatal()
	}
	if _, ok := <-c; ok {
		t.Fatal()
	}
}

func TestConverge3(t *testing.T) {
	a := make(chan int)
	b := make(chan string)
	c := make(chan float32)
	d := pipe.Converge3(a, b, c)
	go func() {
		a <- 42
		b <- "foo"
		c <- 3.14
		close(a)
		close(b)
		close(c)
	}()
	if <-d != 42 || <-d != "foo" || <-d != float32(3.14) {
		t.Fatal()
	}
	if _, ok := <-d; ok {
		t.Fatal()
	}
}

func TestConvergeN(t *testing.T) {
	a := make(chan int)
	b := make(chan string)
	c := make(chan float32)
	d := make(chan byte)
	r := pipe.ConvergeN(a, b, c, d)
	go func() {
		a <- 42
		b <- "foo"
		c <- 3.14
		d <- 1
		close(a)
		close(b)
		close(c)
		close(d)
	}()
	if <-r != 42 || <-r != "foo" || <-r != float32(3.14) || <-r != byte(1) {
		t.Fatal()
	}
	if _, ok := <-r; ok {
		t.Fatal()
	}
}

func TestConverge0(t *testing.T) {
	r := pipe.ConvergeN()
	if _, ok := <-r; ok {
		t.Fatal()
	}
}
