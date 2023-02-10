package pipe_test

import (
	"testing"

	"github.com/hsfzxjy/pipe"
)

func TestController(t *testing.T) {
	c := pipe.NewController[string]()
	if c.Send("noop") {
		t.Fatal()
	}
	l, _ := c.Listen()
	c.Sink() <- "foo"
	if !c.Send("bar") {
		t.Fatal()
	}
	close(c.Sink())
	if <-l != "foo" || <-l != "bar" {
		t.Fatal()
	}
	if _, ok := <-l; ok {
		t.Fatal()
	}
}

func TestControllerC(t *testing.T) {
	c := pipe.NewControllerC[string]()
	if c.Send("noop") {
		t.Fatal()
	}
	l, _ := c.Listen()
	c.Sink() <- "foo"
	if !c.Send("bar") {
		t.Fatal()
	}
	close(c.Sink())
	if <-l != "foo" || <-l != "bar" {
		t.Fatal()
	}
	if _, ok := <-l; ok {
		t.Fatal()
	}
}

func TestControllerM(t *testing.T) {
	c := pipe.NewControllerM("0")
	if c.Send("noop") {
		t.Fatal()
	}
	l, _ := c.Listen()
	c.Sink() <- "foo"
	if !c.Send("bar") {
		t.Fatal()
	}
	close(c.Sink())
	if <-l != "0" || <-l != "foo" || <-l != "bar" {
		t.Fatal()
	}
	if _, ok := <-l; ok {
		t.Fatal()
	}
}

func TestControllerCM(t *testing.T) {
	c := pipe.NewControllerCM("0")
	if c.Send("noop") {
		t.Fatal()
	}
	l, _ := c.Listen()
	c.Sink() <- "foo"
	if !c.Send("bar") {
		t.Fatal()
	}
	close(c.Sink())
	if <-l != "0" || <-l != "foo" || <-l != "bar" {
		t.Fatal()
	}
	if _, ok := <-l; ok {
		t.Fatal()
	}
}
