package queue

import (
	"container/heap"
)

type PriorityType int

const (
	HIGHEST PriorityType = iota
	LOWEST
)

type iPriorityQueueItem interface {
	Priority() uint64
}

type IPriorityQueueItem interface {
	iPriorityQueueItem
	Index() int
	_unsafeSelf() interface{}
}

type priorityQueueItem struct {
	iPriorityQueueItem
	index int
}

func (q *priorityQueueItem) Index() int {
	return q.index
}

func (q *priorityQueueItem) _unsafeSelf() interface{} {
	return q.iPriorityQueueItem
}

func NewPriorityQueueItem(v iPriorityQueueItem) IPriorityQueueItem {
	return &priorityQueueItem{
		iPriorityQueueItem: v,
	}
}

// A PriorityQueue implements heap.Interface and holds Items.
type priorityQueueObj []*priorityQueueItem
type priorityQueueLowestObj struct {
	priorityQueueObj
}

func (pq priorityQueueObj) Len() int { return len(pq) }

func (pq priorityQueueObj) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority() > pq[j].Priority()
}
func (pq priorityQueueLowestObj) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq.priorityQueueObj[i].Priority() < pq.priorityQueueObj[j].Priority()
}

func (pq priorityQueueObj) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueueObj) Push(x interface{}) {
	n := len(*pq)
	item := x.(*priorityQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueueObj) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item._unsafeSelf()
}

func (pq *priorityQueueObj) Peek() interface{} {
	if len(*pq) == 0 {
		return nil
	}
	return (*pq)[0]._unsafeSelf()
}

type iPriorityQueue interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
	Push(x interface{}) // add x as element Len()
	Pop() interface{}   // remove and return element Len() - 1.
	Peek() interface{}
}

type IPriorityQueue interface {
	Push(v IPriorityQueueItem)
	Pop() IPriorityQueueItem
	Peek() IPriorityQueueItem
	Iter(f func(IPriorityQueueItem) bool)
	Remove(index int)
	Size() int
	Clear()
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue struct {
	iPriorityQueue
}

func NewPriorityQueue(pt PriorityType) IPriorityQueue {
	var i iPriorityQueue
	if pt == HIGHEST {
		i = &priorityQueueObj{}
	} else {
		i = &priorityQueueLowestObj{}
	}
	q := &PriorityQueue{
		i,
	}
	heap.Init(q.iPriorityQueue)
	return q
}

func (q *PriorityQueue) Push(v IPriorityQueueItem) {
	item := NewPriorityQueueItem(v)
	heap.Push(q.iPriorityQueue, item)
}

func (q *PriorityQueue) Pop() IPriorityQueueItem {
	i := heap.Pop(q.iPriorityQueue)
	if i == nil {
		return nil
	}
	return i.(IPriorityQueueItem)
}

func (q *PriorityQueue) Peek() IPriorityQueueItem {
	i := q.iPriorityQueue.Peek()
	if i == nil {
		return nil
	}
	return i.(IPriorityQueueItem)
}

// Iter iterates queue items
func (q *PriorityQueue) Remove(i int) {
	heap.Remove(q.iPriorityQueue, i)
}

// Iter iterates queue items
func (q *PriorityQueue) Iter(f func(IPriorityQueueItem) bool) {
	for {
		item := q.Pop()
		if item == nil {
			break
		}
		if !f(item) {
			break
		}
	}
}

// Size return of queue size
func (q *PriorityQueue) Size() int {
	return q.iPriorityQueue.Len()
}

// clear is remove all
func (q *PriorityQueue) Clear() {
	q.Iter(func(IPriorityQueueItem) bool { return true })
}
