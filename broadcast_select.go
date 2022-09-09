package pipe

type selectResult[T any] struct{ dead, starved *listener[T] }

func select1[T any](head *listener[T], reply chan<- selectResult[T], barrier <-chan struct{}) {
	select {
	case <-barrier:
		reply <- selectResult[T]{}
	case <-head.cancelCh:
		reply <- selectResult[T]{dead: head}
	case head.outCh <- head.curItem():
		if starved := head.advanceItem(); starved {
			reply <- selectResult[T]{starved: head}
		} else {
			reply <- selectResult[T]{}
		}
	}
}

func select2[T any](head *listener[T], reply chan<- selectResult[T], barrier <-chan struct{}) {
	entry0, entry1 := head, head.next
	select {
	case <-barrier:
		reply <- selectResult[T]{}
	case <-entry0.cancelCh:
		reply <- selectResult[T]{dead: entry0}
	case entry0.outCh <- entry0.curItem():
		if starved := entry0.advanceItem(); starved {
			reply <- selectResult[T]{starved: entry0}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entry1.cancelCh:
		reply <- selectResult[T]{dead: entry1}
	case entry1.outCh <- entry1.curItem():
		if starved := entry1.advanceItem(); starved {
			reply <- selectResult[T]{starved: entry1}
		} else {
			reply <- selectResult[T]{}
		}
	}
}

func select4[T any](head *listener[T], reply chan<- selectResult[T], barrier <-chan struct{}) {
	var entries [4]*listener[T]
	for i := 0; i < 4; i++ {
		entries[i] = head
		head = head.next
	}
	select {
	case <-barrier:
		reply <- selectResult[T]{}
	case <-entries[0].cancelCh:
		reply <- selectResult[T]{dead: entries[0]}
	case entries[0].outCh <- entries[0].curItem():
		if starved := entries[0].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[0]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[1].cancelCh:
		reply <- selectResult[T]{dead: entries[1]}
	case entries[1].outCh <- entries[1].curItem():
		if starved := entries[1].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[1]}
		} else {
			reply <- selectResult[T]{}
		}

	case <-entries[2].cancelCh:
		reply <- selectResult[T]{dead: entries[2]}
	case entries[2].outCh <- entries[2].curItem():
		if starved := entries[2].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[2]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[3].cancelCh:
		reply <- selectResult[T]{dead: entries[3]}
	case entries[3].outCh <- entries[3].curItem():
		if starved := entries[3].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[3]}
		} else {
			reply <- selectResult[T]{}
		}
	}
}

func select8[T any](head *listener[T], reply chan<- selectResult[T], barrier <-chan struct{}) {
	var entries [8]*listener[T]
	for i := 0; i < 8; i++ {
		entries[i] = head
		head = head.next
	}
	select {
	case <-barrier:
		reply <- selectResult[T]{}
	case <-entries[0].cancelCh:
		reply <- selectResult[T]{dead: entries[0]}
	case entries[0].outCh <- entries[0].curItem():
		if starved := entries[0].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[0]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[1].cancelCh:
		reply <- selectResult[T]{dead: entries[1]}
	case entries[1].outCh <- entries[1].curItem():
		if starved := entries[1].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[1]}
		} else {
			reply <- selectResult[T]{}
		}

	case <-entries[2].cancelCh:
		reply <- selectResult[T]{dead: entries[2]}
	case entries[2].outCh <- entries[2].curItem():
		if starved := entries[2].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[2]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[3].cancelCh:
		reply <- selectResult[T]{dead: entries[3]}
	case entries[3].outCh <- entries[3].curItem():
		if starved := entries[3].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[3]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[4].cancelCh:
		reply <- selectResult[T]{dead: entries[4]}
	case entries[4].outCh <- entries[4].curItem():
		if starved := entries[4].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[4]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[5].cancelCh:
		reply <- selectResult[T]{dead: entries[5]}
	case entries[5].outCh <- entries[5].curItem():
		if starved := entries[5].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[5]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[6].cancelCh:
		reply <- selectResult[T]{dead: entries[6]}
	case entries[6].outCh <- entries[6].curItem():
		if starved := entries[6].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[6]}
		} else {
			reply <- selectResult[T]{}
		}
	case <-entries[7].cancelCh:
		reply <- selectResult[T]{dead: entries[7]}
	case entries[7].outCh <- entries[7].curItem():
		if starved := entries[7].advanceItem(); starved {
			reply <- selectResult[T]{starved: entries[7]}
		} else {
			reply <- selectResult[T]{}
		}
	}

}
