package pipe

import (
	"fmt"
	"reflect"
)

// Converge2 converges values from ch1 and ch2 into returned channel.
func Converge2[A, B any](ch1 <-chan A, ch2 <-chan B) <-chan any {
	out := make(chan any)
	go func() {
		defer close(out)
		n := 2
		for n > 0 {
			select {
			case x, ok := <-ch1:
				if !ok {
					n--
					ch1 = nil
				} else {
					out <- x
				}
			case x, ok := <-ch2:
				if !ok {
					n--
					ch2 = nil
				} else {
					out <- x
				}
			}
		}
	}()
	return out
}

// Converge2 converges values from ch1, ch3 and ch2 into returned channel.
func Converge3[A, B, C any](ch1 <-chan A, ch2 <-chan B, ch3 <-chan C) <-chan any {
	out := make(chan any)
	go func() {
		defer close(out)
		n := 3
		for n > 0 {
			select {
			case x, ok := <-ch1:
				if !ok {
					n--
					ch1 = nil
				} else {
					out <- x
				}
			case x, ok := <-ch2:
				if !ok {
					n--
					ch2 = nil
				} else {
					out <- x
				}
			case x, ok := <-ch3:
				if !ok {
					n--
					ch3 = nil
				} else {
					out <- x
				}
			}

		}
	}()
	return out
}

// ConvergeN converges values from arbitary number of channels.
// Each of chans should be of type <-chan T for some T.
func ConvergeN(chans ...any) <-chan any {
	out := make(chan any)
	cases := make([]reflect.SelectCase, len(chans))
	for i, ch := range chans {
		value := reflect.ValueOf(ch)
		typ := value.Type()
		if typ.Kind() != reflect.Chan || (typ.ChanDir() != reflect.BothDir && typ.ChanDir() != reflect.RecvDir) {
			panic(fmt.Sprintf("expect a <-chan * for %d-th argument, got %s", i, typ.String()))
		}
		cases[i] = reflect.SelectCase{
			Chan: value,
			Dir:  reflect.SelectRecv,
		}
	}
	go func() {
		defer close(out)
		n := len(cases)
		for n > 0 {
			i, x, ok := reflect.Select(cases)
			if !ok {
				n--
				cases[i].Chan = reflect.Zero(cases[i].Chan.Type())
				continue
			}
			out <- x.Interface()
		}
	}()
	return out
}
