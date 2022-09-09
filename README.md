# pipe

pipe provides versatile channel transformers for Golang. Currently it can:

- Converge values from multiple upstream channels into single downstream channel;
- Broadcast values from single upstream channels to multiple downstream channels.

## Channel Broadcasting

### Simple Broadcaster

Use `Broadcast` to create a simple broadcaster

```go
upstream := make(chan int)
b := pipe.Broadcast(upstream)
```

Broadcaster provides `.Listen` and `.Bind` methods for registering downstream listeners

```go
l := make(chan int)
canceler := b.Bind(l)
// equivalent to
l2, canceler2 := b.Listen()
upstream <- 42
l3, canceler3 := b.Listen()
close(upstream)

// value 42 is multiplexed to all listeners
<-l  // 42
<-l2 // 42

// no more values after upstream was closed, broadcaster will pipe
// pending values out to all listeners and close them.
_, ok := <-l   // ok == false
_, ok2 := <-l2 // ok2 == false

// since l3 was registered after value 42 sent, it was closed without
// receiving any value
_, ok3 := <-l3 // ok3 == false
```

The returned `canceler` can be used for canceling subscription

```go
l, canceler := b.Listen()
upstream <- 42
<-l          // 42
canceler()
_, ok := <-l // ok == false, since l was closed after cancellation
```

Broadcaster gaurantees that sending to upstream channel will not block, even if there's no listeners

```go
upstream := make(chan int)
b := pipe.Broadcast(upstream)
upstream <- 42 // won't block
```

Use `.Detach` if you want to stop the broadcaster prematurely before the upstream closed, after which sending to upstream will block again

```go
upstream := make(chan int)
b := pipe.Broadcast(upstream)
b.Detach()
upstream <- 42 // will block
```

### Memorizable Broadcaster

Sometimes you may expect newly registered listener to be immediately fed with the latest value from upstream. For this scenario, we use the `BroadcastM` constructor

```go
upstream := make(chan int)
b := pipe.BroadcastM(upstream, 0) // 0 is an initial value
l, _ := b.Listen()
<-l  // 0
upstream <- 42
l2, _ := b.Listen()
<-l2 // 42
```

Memorizable broadcaster provides `.Current` method to retrieve the latest value

```go
upstream := make(chan int)
b := pipe.BroadcastM(upstream, 0)
b.Current() // 0
upstream <- 42
b.Current() // 42
```

### Broadcaster with comparable element type

For upstream with a comparable element type (`int`, `string`, etc.), we can use `BroadcastC` to create a broadcaster that provides additional useful methods. 

Currently, such broadcaster provides `.Until(targets...)` method, which will block until one of the values in `targets` shown up from the upstream.

```go
upstream := make(chan int)
b := pipe.BroadcastC(upstream)
go func() {
    for i := 0; i < 5; i++ {
        time.Sleep(1 * time.Second)
        b <- i
    }
}()
// block until value 3 or 4 shown up from the upstream, which should be
// approximately 4 seconds
b.Until(3, 4)
```

`.Until` has variants like `.UntilCh` and `.UntilContext`.

And also we have `BroadcastCM`, which combines the functionality of `BroadcastC` and `BroadcastM`. For more details please refer to [godoc](https://pkg.go.dev/github.com/hsfzxjy/pipe).

## Controller and Listener

A Controller bundles an upstream and a broadcaster, which is handy in some cases

```go
con := pipe.NewController[int]()
l, _ := con.Listen()

// con.Sink() exposes the upstream channel
con.Sink() <- 42
close(con.Sink())

<-l // 42
```

It's recommended to store controllers as private fields and expose them as Listenable interface, to which users can bind listeners. Consider an imaginary scenario

```go
// library-side
type State int

type Service struct {
    state *pipe.Controller[State]
}

func (ser *Service) State() Listenable[State] { return ser.state }
func (ser *Service) businessLogic() {
    // ...
    ser.state.Sink() <- SomeState
}

// user-side
listener, _ := service.State().Listen()
```

Similarly, there are variants like `Controller(C|M|CM)` and `Listenable(C|M|CM)`.

## Channel Converging

The method `Converge2`, `Converge3` and `ConvergeN` implements the channel converging logic

```go
// Converge 2 channels
a := make(chan int)
b := make(chan string)
r := pipe.Converge2(a, b)
go func() {
    a <- 42
    b <- "foo"
    close(a)
    close(b)
}()
for x := range r {
    println(x)
}
// Output:
// 42
// foo
```

```go
// Converge arbitary number of channels
a := make(chan int)
b := make(chan string)
c := make(chan float64)
d := make(chan byte)
r := pipe.Converge2(a, b)
go func() {
    a <- 42
    b <- "foo"
    c <- 3.14
    d <- byte(127)
    close(a)
    close(b)
    close(c)
    close(d)
}()
for x := range r {
    println(x)
}
// Output:
// 42
// foo
// 3.14
// 127
```

# License

The library is licensed under the MIT License.