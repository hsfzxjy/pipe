package pipe

type sink[T any] struct{ ch chan T }

// Sink returns the upstream channel of the controller.
func (s sink[T]) Sink() chan<- T { return s.ch }

// A Controller bundles a upstream channel and a broadcaster.
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

// Current returns the latest value that the broadcaster memorizes.
func (c *ControllerM[T]) Current() T { return c.broadcaster.current() }

type ControllerCM[T comparable] struct {
	sink[T]
	broadcasterc[T]
}

// A Controller with a comparable element type and memorizable broadcaster.
func NewControllerCM[T comparable](initial T) *ControllerCM[T] {
	c := new(ControllerCM[T])
	c.sink.ch = make(chan T)
	c.broadcaster.init(c.sink.ch, &initial)
	return c
}

// Current returns the latest value that the broadcaster memorizes.
func (c *ControllerCM[T]) Current() T { return c.broadcaster.current() }
