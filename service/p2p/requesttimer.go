package p2p

import (
	"sync"
	"time"
)

// RequestExpireHandler handles a request expire event
type RequestExpireHandler interface {
	OnTimerExpired(height uint32, value string)
}

// RequestTimer triggers a event when a request is expired
type RequestTimer struct {
	sync.Mutex
	timerMap map[uint32]*requestTimerItem
	valueMap map[string]map[uint32]bool
	handler  RequestExpireHandler
}

// NewRequestTimer returns a RequestTimer
func NewRequestTimer(handler RequestExpireHandler) *RequestTimer {
	rm := &RequestTimer{
		timerMap: map[uint32]*requestTimerItem{},
		valueMap: map[string]map[uint32]bool{},
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
func (rm *RequestTimer) Add(height uint32, t time.Duration, value string) {
	rm.Lock()
	defer rm.Unlock()

	rm.timerMap[height] = &requestTimerItem{
		Height:    height,
		ExpiredAt: uint64(time.Now().UnixNano()) + uint64(t),
		Value:     value,
	}
	heightMap, has := rm.valueMap[value]
	if !has {
		heightMap = map[uint32]bool{}
		rm.valueMap[value] = heightMap
	}
	heightMap[height] = true
}

// RemovesByValue removes requests by the value
func (rm *RequestTimer) RemovesByValue(value string) {
	rm.Lock()
	defer rm.Unlock()

	heightMap, has := rm.valueMap[value]
	if has {
		for height := range heightMap {
			delete(rm.timerMap, height)
		}
	}
	delete(rm.valueMap, value)
}

// Run is the main loop of RequestTimer
func (rm *RequestTimer) Run() {
	for {
		expired := []*requestTimerItem{}
		now := uint64(time.Now().UnixNano())
		remainMap := map[uint32]*requestTimerItem{}
		rm.Lock()
		for h, v := range rm.timerMap {
			if v.ExpiredAt <= now {
				expired = append(expired, v)

				heightMap, has := rm.valueMap[v.Value]
				if has {
					delete(heightMap, v.Height)
				}
				if len(heightMap) == 0 {
					delete(rm.valueMap, v.Value)
				}
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
		time.Sleep(200 * time.Millisecond)
	}
}

type requestTimerItem struct {
	Height    uint32
	ExpiredAt uint64
	Value     string
}
