package pipe

import (
	"runtime"
	"sync/atomic"
	"time"
)

type pond struct {
	idleCount int32
	tasks     chan func()
	dispatch  chan func()
}

var selectorPond = newPond()

func newPond() *pond {
	b := new(pond)
	b.tasks = make(chan func())
	b.dispatch = make(chan func())
	go b.loop()
	go b.idleLoop()
	return b
}

func (b *pond) loop() {
	for task := range b.tasks {
		if atomic.LoadInt32(&b.idleCount) == 0 {
			go b.worker(task)
			continue
		}
		fails := 0
	INNER:
		for {
			if fails == 5 {
				go b.worker(task)
				break INNER
			}
			select {
			case b.dispatch <- task:
				break INNER
			default:
				runtime.Gosched()
				fails++
			}
		}
	}
}

func (b *pond) idleLoop() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		if atomic.LoadInt32(&b.idleCount) != 0 {
			select {
			case b.dispatch <- nil:
			case <-ticker.C:
			}
		}
	}
}

func (b *pond) worker(task func()) {
	task()
	atomic.AddInt32(&b.idleCount, 1)
	for task = range b.dispatch {
		atomic.AddInt32(&b.idleCount, -1)
		if task == nil {
			return
		} else {
			task()
			atomic.AddInt32(&b.idleCount, 1)
		}
	}
}
