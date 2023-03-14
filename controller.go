package pipe

type sink[T any] struct{ ch chan T }

// A Controller bundles a sink channel and a broadcaster.
type Controller[T any] struct {
	sink[T]
	broadcaster[T]
}

func NewController[T any]() *Controller[T] {
	c := new(Controller[T])
	c.sink.ch = make(chan T)
	c.broadcaster.init(c.sink.ch, nil)
	return c
}

// Sink returns the sink channel of the controller.
func (c *Controller[T]) Sink() chan<- T {
	c.ensureInit()
	return c.ch
}

// Send sends value to the sink channel.
func (c *Controller[T]) Send(value T) (ok bool) {
	if c.initialized() {
		c.ch <- value
		return true
	}
	return false
}

// A Controller with a comparable element type.
type ControllerC[T comparable] struct {
	sink[T]
	broadcasterc[T]
}

func NewControllerC[T comparable]() *ControllerC[T] {
	c := new(ControllerC[T])
	c.sink.ch = make(chan T)
	c.broadcaster.init(c.sink.ch, nil)
	return c
}

// Sink returns the sink channel of the controller.
func (c *ControllerC[T]) Sink() chan<- T {
	c.ensureInit()
	return c.ch
}

// Send sends value to the sink channel.
func (c *ControllerC[T]) Send(value T) (ok bool) {
	if c.initialized() {
		c.ch <- value
		return true
	}
	return false
}

// A Controller with a memorizable broadcaster.
type ControllerM[T any] struct {
	sink[T]
	broadcaster[T]
}

func NewControllerM[T any](initial T) *ControllerM[T] {
	c := new(ControllerM[T])
	c.sink.ch = make(chan T)
	c.broadcaster.init(c.sink.ch, &initial)
	return c
}

// Sink returns the sink channel of the controller.
func (c *ControllerM[T]) Sink() chan<- T {
	c.ensureInit()
	return c.ch
}

// Send sends value to the sink channel.
func (c *ControllerM[T]) Send(value T) (ok bool) {
	if c.initialized() {
		c.ch <- value
		return true
	}
	c.replaceBuf(value)
	return false
}

// Current returns the latest value that the broadcaster memorizes.
func (c *ControllerM[T]) Current() T { return c.broadcaster.current() }

type ControllerCM[T comparable] struct {
	sink[T]
	broadcasterc[T]
}

// A Controller with a comparable element type and memorizable broadcaster.
func NewControllerCM[T comparable](initial T, dedup bool) *ControllerCM[T] {
	c := new(ControllerCM[T])
	c.sink.ch = make(chan T)
	c.broadcaster.init(c.sink.ch, &initial)
	c.dedup = dedup
	return c
}

// Sink returns the sink channel of the controller.
func (c *ControllerCM[T]) Sink() chan<- T {
	c.ensureInit()
	return c.ch
}

// Send sends value to the sink channel.
func (c *ControllerCM[T]) Send(value T) (ok bool) {
	if c.initialized() {
		c.ch <- value
		return true
	}
	c.replaceBuf(value)
	return false
}

// Current returns the latest value that the broadcaster memorizes.
func (c *ControllerCM[T]) Current() T { return c.broadcaster.current() }
