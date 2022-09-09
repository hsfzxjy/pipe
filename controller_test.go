package pipe_test

import (
	"testing"

	"github.com/hsfzxjy/pipe"
)

func TestController(t *testing.T) {
	c := pipe.NewController[string]()
	l, _ := c.Listen()
	c.Sink() <- "foo"
	c.Sink() <- "bar"
	close(c.Sink())
	if <-l != "foo" || <-l != "bar" {
		t.Fatal()
	}
	if _, ok := <-l; ok {
		t.Fatal()
	}
}
