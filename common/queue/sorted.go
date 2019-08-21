package queue

import (
	"sort"
	"sync"
)

// SortedQueue sorts items by the priority
type SortedQueue struct {
	sync.Mutex
	items []*sortedQueueItem
	head  int
	size  int
}

// NewSortedQueue returns a SortedQueue
func NewSortedQueue() *SortedQueue {
	q := &SortedQueue{
		items: make([]*sortedQueueItem, 0, 256),
	}
	return q
}

// FindOrInsert finds the item of the priority if it exists or inserts the item by the priority
func (q *SortedQueue) FindOrInsert(value interface{}, Priority uint64) interface{} {
	q.Lock()
	defer q.Unlock()

	if item := q.findInternal(Priority); item != nil {
		return item
	}
	q.insert(value, Priority)
	return nil
}

// Insert inserts the item by the priority
func (q *SortedQueue) Insert(value interface{}, Priority uint64) {
	q.Lock()
	defer q.Unlock()

	q.insert(value, Priority)
}

func (q *SortedQueue) insert(value interface{}, Priority uint64) {
	item := sortedQueueItemPool.Get().(*sortedQueueItem)
	item.value = value
	item.priority = Priority
	idx := sort.Search(q.size, func(i int) bool {
		return Priority < q.items[q.head+i].priority
	})
	idx += q.head
	if q.head <= idx && idx < q.head+q.size && q.items[idx].priority == Priority {
		q.items[idx] = item
	} else {
		if q.size == len(q.items) {
			q.items = append(q.items, item)
			copy(q.items[idx+1:], q.items[idx:])
		} else {
			last := q.head + q.size
			if last == len(q.items) {
				copy(q.items[q.head-1:], q.items[q.head:])
				q.head--
				idx--
				if idx != last-1 {
					copy(q.items[idx+1:], q.items[idx:last])
				}
			} else {
				copy(q.items[idx+1:last+1], q.items[idx:last])
			}
		}
		q.items[idx] = item
		q.size++
	}
}

// Peek fetch the top item without removing it
func (q *SortedQueue) Peek() (interface{}, uint64) {
	q.Lock()
	defer q.Unlock()

	return q.peekInternal()
}

func (q *SortedQueue) peekInternal() (interface{}, uint64) {
	if q.size == 0 {
		return nil, 0
	}
	item := q.items[q.head]
	return item.value, item.priority
}

// Find fetch the target priority item without removing it
func (q *SortedQueue) Find(Priority uint64) interface{} {
	q.Lock()
	defer q.Unlock()

	return q.findInternal(Priority)
}

func (q *SortedQueue) findInternal(Priority uint64) interface{} {
	if q.size == 0 {
		return nil
	}
	for i := q.head; i < q.head+q.size; i++ {
		item := q.items[i]
		if item.priority == Priority {
			return item.value
		} else if item.priority > Priority {
			break
		}
	}
	return nil
}

// PopUntil returns a item at the top of the queue
func (q *SortedQueue) PopUntil(Priority uint64) interface{} {
	q.Lock()
	defer q.Unlock()

	for {
		item, itemPriority := q.peekInternal()
		if item == nil {
			return nil
		}
		if itemPriority < Priority {
			q.popInternal()
		} else if itemPriority == Priority {
			q.popInternal()
			return item
		} else {
			return nil
		}
	}
}

// Pop returns a item at the top of the queue
func (q *SortedQueue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	return q.popInternal()
}

func (q *SortedQueue) popInternal() interface{} {
	if q.size == 0 {
		return nil
	}
	item := q.items[q.head]
	q.items[q.head] = nil
	q.head++
	q.size--
	if len(q.items) > 4096 {
		if q.size < 1024 {
			items := make([]*sortedQueueItem, 2048)
			copy(items, q.items[q.head:q.head+q.size])
			q.head = 0
			q.items = items
		} else if q.head > len(q.items)/2+1 {
			copy(q.items, q.items[q.head:q.head+q.size])
			q.head = 0
		}
	}
	return item.value
}

// Size returns the number of items
func (q *SortedQueue) Size() int {
	q.Lock()
	defer q.Unlock()

	return q.size
}

// Iter iterates queue items
func (q *SortedQueue) Iter(fn func(v interface{}, priority uint64)) {
	q.Lock()
	defer q.Unlock()

	for i := 0; i < q.size; i++ {
		item := q.items[q.head+i]
		fn(item.value, item.priority)
	}
}

type sortedQueueItem struct {
	value    interface{}
	priority uint64
}

var sortedQueueItemPool = sync.Pool{
	New: func() interface{} {
		return &sortedQueueItem{}
	},
}
