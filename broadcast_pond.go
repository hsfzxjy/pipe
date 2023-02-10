package pipe

import (
	"runtime"
	"sync/atomic"
	"time"
)

type pond struct {
	idleCount atomic.Int32
	tasks     chan func()
	dispatch  chan func()
}

var selectorPond = newPond()

func newPond() *pond {
	b := new(pond)
	b.tasks = make(chan func())
	b.dispatch = make(chan func())
	go b.loop()
	return b
}

func (b *pond) loop() {
	var ticker *time.Ticker
	var tickerC <-chan time.Time
	var cancelC chan func()
	var total int32
	for {
		select {
		case task := <-b.tasks:
			if ticker == nil {
				ticker = time.NewTicker(time.Second * 5)
				tickerC = ticker.C
			}
			if b.idleCount.Load() == 0 {
				total += 1
				go b.worker(task)
				continue
			}
			fails := 0
		INNER:
			for {
				if fails == 10 {
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
		case <-tickerC:
			if total != 0 {
				cancelC = b.dispatch
			} else {
				ticker.Stop()
				ticker = nil
				tickerC = nil
			}
		case cancelC <- nil:
			total -= 1
			if b.idleCount.Load() < total/2 {
				cancelC = nil
			}
		}
	}

}

func (b *pond) worker(task func()) {
	task()
	b.idleCount.Add(1)
	for task = range b.dispatch {
		b.idleCount.Add(-1)
		if task == nil {
			return
		}
		task()
		b.idleCount.Add(1)
	}
}
