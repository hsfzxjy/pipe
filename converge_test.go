package pipe_test

import (
	"testing"

	"github.com/hsfzxjy/pipe"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, 42, <-c)
	assert.Equal(t, "foo", <-c)
	_, ok := <-c
	assert.False(t, ok)
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
	assert.Equal(t, 42, <-d)
	assert.Equal(t, "foo", <-d)
	assert.Equal(t, float32(3.14), <-d)
	_, ok := <-d
	assert.False(t, ok)
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
	assert.Equal(t, 42, <-r)
	assert.Equal(t, "foo", <-r)
	assert.Equal(t, float32(3.14), <-r)
	assert.Equal(t, byte(1), <-r)
}

func TestConverge0(t *testing.T) {
	r := pipe.ConvergeN()
	_, ok := <-r
	assert.False(t, ok)
}
