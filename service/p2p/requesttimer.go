package p2p

import (
	"sync"
	"time"
)

// RequestExpireHandler handles a request expire event
type RequestExpireHandler interface {
	OnTimerExpired(height uint32, value interface{})
}

// RequestTimer triggers a event when a request is expired
type RequestTimer struct {
	sync.Mutex
	timerMap map[uint32]*requestTimerItem
	handler  RequestExpireHandler
}

// NewRequestTimer returns a RequestTimer
func NewRequestTimer(handler RequestExpireHandler) *RequestTimer {
	rm := &RequestTimer{
		timerMap: map[uint32]*requestTimerItem{},
		handler:  handler,
	}
	return rm
}

// Exist returns the target height request exists or not
func (rm *RequestTimer) Exist(height uint32) bool {
	rm.Lock()
	defer rm.Unlock()

	_, has := rm.timerMap[height]
	return has
}

// Add adds the timer of the request
func (rm *RequestTimer) Add(height uint32, t time.Duration, value interface{}) {
	rm.Lock()
	defer rm.Unlock()

	rm.timerMap[height] = &requestTimerItem{
		Height:    height,
		ExpiredAt: uint64(time.Now().UnixNano()) + uint64(t),
		Value:     value,
	}
}

// Remove removes the timer of the request
func (rm *RequestTimer) Remove(height uint32) {
	rm.Lock()
	defer rm.Unlock()

	delete(rm.timerMap, height)
}

// Run is the main loop of RequestTimer
func (rm *RequestTimer) Run() {
	timer := time.NewTimer(100 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			expired := []*requestTimerItem{}
			now := uint64(time.Now().UnixNano())
			remainMap := map[uint32]*requestTimerItem{}
			rm.Lock()
			for h, v := range rm.timerMap {
				if v.ExpiredAt <= now {
					expired = append(expired, v)
				} else {
					remainMap[h] = v
				}
			}
			rm.timerMap = remainMap
			rm.Unlock()

			if rm.handler != nil {
				for _, v := range expired {
					rm.handler.OnTimerExpired(v.Height, v.Value)
				}
			}
			timer.Reset(100 * time.Millisecond)
		}
	}
}

type requestTimerItem struct {
	Height    uint32
	ExpiredAt uint64
	Value     interface{}
}
