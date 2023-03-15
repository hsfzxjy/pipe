package pipe_test

import (
	"testing"

	"github.com/hsfzxjy/pipe"
	"github.com/stretchr/testify/assert"
)

func TestController(t *testing.T) {
	c := pipe.NewController[string]()
	assert.False(t, c.Send("noop"))
	l, _ := c.Listen()
	c.Sink() <- "foo"
	assert.True(t, c.Send("bar"))
	close(c.Sink())
	assert.Equal(t, "foo", <-l)
	assert.Equal(t, "bar", <-l)
	_, ok := <-l
	assert.False(t, ok)
}

func TestControllerC(t *testing.T) {
	c := pipe.NewControllerC[string]()
	assert.False(t, c.Send("noop"))
	l, _ := c.Listen()
	c.Sink() <- "foo"
	assert.True(t, c.Send("bar"))
	close(c.Sink())
	assert.Equal(t, "foo", <-l)
	assert.Equal(t, "bar", <-l)
	_, ok := <-l
	assert.False(t, ok)
}

func TestControllerM(t *testing.T) {
	c := pipe.NewControllerM("0")
	assert.False(t, c.Send("noop"))
	l, _ := c.Listen()
	c.Sink() <- "foo"
	assert.True(t, c.Send("bar"))
	close(c.Sink())
	assert.Equal(t, "noop", <-l)
	assert.Equal(t, "foo", <-l)
	assert.Equal(t, "bar", <-l)
	_, ok := <-l
	assert.False(t, ok)
}

func TestControllerCM(t *testing.T) {
	c := pipe.NewControllerCM("0", false)
	assert.False(t, c.Send("noop"))
	l, _ := c.Listen()
	c.Sink() <- "foo"
	assert.True(t, c.Send("bar"))
	close(c.Sink())
	assert.Equal(t, "noop", <-l)
	assert.Equal(t, "foo", <-l)
	assert.Equal(t, "bar", <-l)
	_, ok := <-l
	assert.False(t, ok)
}

func TestControllerCMDedup(t *testing.T) {
	c := pipe.NewControllerCM("0", true)
	assert.False(t, c.Send("noop"))
	l, _ := c.Listen()
	c.Sink() <- "foo"
	c.Sink() <- "foo"
	assert.True(t, c.Send("bar"))
	close(c.Sink())
	assert.Equal(t, "noop", <-l)
	assert.Equal(t, "foo", <-l)
	assert.Equal(t, "bar", <-l)
	_, ok := <-l
	assert.False(t, ok)
}
