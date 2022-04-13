package common

import "sync"

type SyncChan struct {
	Chan       chan interface{}
	closeMutex sync.Mutex
	isClose    bool
}

func NewSyncChan() *SyncChan {
	return &SyncChan{
		Chan: make(chan interface{}, 2),
	}
}

func (t *SyncChan) Close() {
	t.closeMutex.Lock()
	t.isClose = true
	t.closeMutex.Unlock()
	close(t.Chan)
}

func (t *SyncChan) Send(v interface{}) {
	t.closeMutex.Lock()
	if !t.isClose {
		t.Chan <- v
		t.closeMutex.Unlock()
	} else {
		t.closeMutex.Unlock()
	}
}
