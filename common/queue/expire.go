package queue

import (
	"sync"
	"time"
)

// ItemExpireHandler handles a group expire event
type ItemExpireHandler interface {
	// OnItemExpired is called when the item is expired
	OnItemExpired(interval time.Duration, key string, item interface{}, IsLast bool)
}

// ExpireQueue provides sequential expirations by groups
type ExpireQueue struct {
	sync.Mutex
	groups   []*group
	handlers []ItemExpireHandler
}

// NewExpireQueue returns ExpireQueue
func NewExpireQueue() *ExpireQueue {
	q := &ExpireQueue{}
	return q
}

// AddGroup adds and runs a expire gruop
func (q *ExpireQueue) AddGroup(interval time.Duration) {
	q.Lock()
	defer q.Unlock()

	g := &group{
		Interval: interval,
		itemMap:  map[string]*groupItem{},
	}
	idx := len(q.groups)
	q.groups = append(q.groups, g)

	go func() {
		timer := time.NewTimer(time.Millisecond)
		for {
			select {
			case <-timer.C:
				now := time.Now().UnixNano()
				expiredMap := map[string]interface{}{}
				q.Lock()
				deletes := []string{}
				for k, v := range g.itemMap {
					if v.expiredAt <= now {
						expiredMap[k] = v.item
						deletes = append(deletes, k)
					}
				}
				for _, k := range deletes {
					delete(g.itemMap, k)
				}
				q.Unlock()

				if len(expiredMap) > 0 {
					q.Lock()
					IsLast := idx+1 == len(q.groups)
					q.Unlock()

					for _, h := range q.handlers {
						for k, v := range expiredMap {
							h.OnItemExpired(g.Interval, k, v, IsLast)
						}
					}

					q.Lock()
					if !IsLast {
						next := q.groups[idx+1]
						for k, v := range expiredMap {
							next.itemMap[k] = &groupItem{
								expiredAt: time.Now().Add(next.Interval).UnixNano(),
								item:      v,
							}
						}
					}
					q.Unlock()
				}

				timer.Reset(time.Second)
			}
		}
	}()
}

// Size returns a queue size
func (q *ExpireQueue) Size() int {
	q.Lock()
	defer q.Unlock()

	Sum := 0
	for _, g := range q.groups {
		Sum += len(g.itemMap)
	}
	return Sum
}

// AddHandler adds a item expire handler
func (q *ExpireQueue) AddHandler(handler ItemExpireHandler) {
	q.Lock()
	defer q.Unlock()

	q.handlers = append(q.handlers, handler)
}

// Push adds a item to the expiration flow
func (q *ExpireQueue) Push(key string, item interface{}) {
	q.Lock()
	defer q.Unlock()

	g := q.groups[0]
	g.itemMap[key] = &groupItem{
		expiredAt: time.Now().Add(g.Interval).UnixNano(),
		item:      item,
	}
}

// Remove removes a item from the expiration flow
func (q *ExpireQueue) Remove(key string) {
	q.Lock()
	defer q.Unlock()

	for _, g := range q.groups {
		delete(g.itemMap, key)
	}
}

type group struct {
	Interval time.Duration
	itemMap  map[string]*groupItem
}

type groupItem struct {
	expiredAt int64
	item      interface{}
}
