package pipe

import (
	"sync"
	"sync/atomic"
)

type once struct {
	done uint32
	m    sync.Mutex
}

func (o *once) Do(f func()) {

	if atomic.LoadUint32(&o.done) == 0 {
		o.doSlow(f)
	}
}

func (o *once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}

func (o *once) Done() bool {
	return atomic.LoadUint32(&o.done) == 1
}
