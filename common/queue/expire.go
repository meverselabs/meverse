package queue

import (
	"sync"
	"time"
)

type TxExpiredType int

const (
	Resend TxExpiredType = iota
	Expired
	Remain
	Error
)

var txExpiredTypes = [...]string{
	"Expired",
	"Remain",
	"Error",
}

func (m TxExpiredType) String() string { return txExpiredTypes[m] }

// ItemExpireHandler handles a group expire event
type ItemExpireHandler interface {
	// OnItemExpired is called when the item is expired
	OnItemExpired(interval time.Duration, key string, item interface{}, IsLast bool) TxExpiredType
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

// AddGroupRepeat adds and runs a expire gruop n times
func (q *ExpireQueue) AddGroupRepeat(n int, in time.Duration) {
	for i := 0; i < n; i++ {
		q.AddGroup(in)
	}
}

// AddGroup adds and runs a expire gruop
func (q *ExpireQueue) AddGroup(interval time.Duration) {
	q.Lock()
	defer q.Unlock()

	g := &group{
		Interval: interval,
		itemMap:  map[string]*groupItem{},
	}
	q.groups = append(q.groups, g)
	go q.manageExpireItem(g, len(q.groups))
}

func (q *ExpireQueue) manageExpireItem(g *group, idx int) {
	for {
		expiredMap := q.setupUpToDateItem(g)

		if len(expiredMap) > 0 {
			IsLast := q.isLast(idx)
			q.deleteExpiredItem(expiredMap, g, IsLast)
			q.sendItemNextGroup(IsLast, idx, expiredMap)
		}
		time.Sleep(time.Second)
	}
}

func (q *ExpireQueue) sendItemNextGroup(IsLast bool, idx int, expiredMap map[string]interface{}) {
	q.Lock()
	if !IsLast {
		next := q.groups[idx]
		for k, v := range expiredMap {
			next.itemMap[k] = &groupItem{
				expiredAt: time.Now().Add(next.Interval).UnixNano(),
				item:      v,
			}
		}
	}
	q.Unlock()
}

func (q *ExpireQueue) deleteExpiredItem(expiredMap map[string]interface{}, g *group, IsLast bool) {
	deletes := []string{}
	for _, h := range q.handlers {
		for k, v := range expiredMap {
			switch h.OnItemExpired(g.Interval, k, v, IsLast) {
			case Expired, Error:
				deletes = append(deletes, k)
			}
		}
	}
	q.Lock()
	for _, k := range deletes {
		delete(g.itemMap, k)
		delete(expiredMap, k)
	}
	q.Unlock()
}

func (q *ExpireQueue) setupUpToDateItem(g *group) map[string]interface{} {
	expiredMap := map[string]interface{}{}
	now := time.Now().UnixNano()

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

	return expiredMap
}

// Size returns a queue size
func (q *ExpireQueue) isLast(idx int) bool {
	q.Lock()
	defer q.Unlock()
	return idx == len(q.groups)
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
