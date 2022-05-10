package queue

import (
	"sync"

	"github.com/meverselabs/meverse/common/hash"
)

// LinkedQueue is designed to allow users to remove the item by the key
type LinkedQueue struct {
	sync.Mutex
	Head   *linkedItem
	Tail   *linkedItem
	keyMap map[hash.Hash256]*linkedItem
}

// NewLinkedQueue make a LinkedQueue
func NewLinkedQueue() *LinkedQueue {
	q := &LinkedQueue{
		keyMap: map[hash.Hash256]*linkedItem{},
	}
	return q
}

// Size returns the number of items
func (q *LinkedQueue) Size() int {
	q.Lock()
	defer q.Unlock()

	return len(q.keyMap)
}

// Push inserts the item with the key at the bottom of the queue
func (q *LinkedQueue) Push(Key hash.Hash256, item interface{}) bool {
	q.Lock()
	defer q.Unlock()

	if _, has := q.keyMap[Key]; has {
		return false
	}

	nd := &linkedItem{
		Key:  Key,
		Item: item,
	}
	if q.Head == nil {
		q.Head = nd
		q.Tail = nd
	} else {
		nd.Prev = q.Tail
		q.Tail.Next = nd
		q.Tail = nd
	}
	q.keyMap[Key] = nd
	return true
}

// Pop returns a item at the top of the queue
func (q *LinkedQueue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	if q.Head == nil {
		return nil
	}
	nd := q.Head
	if nd == q.Tail {
		q.Head = nil
		q.Tail = nil
	} else {
		q.Head = nd.Next
		if q.Head != nil {
			q.Head.Prev = nil
		}
		nd.Next.Prev = nil
	}
	nd.Prev = nil
	nd.Next = nil
	delete(q.keyMap, nd.Key)
	return nd.Item
}

// Remove deletes a item by the key
func (q *LinkedQueue) Remove(Key hash.Hash256) interface{} {
	q.Lock()
	defer q.Unlock()

	if q.Head == nil {
		return nil
	}
	nd, has := q.keyMap[Key]
	if !has {
		return nil
	}
	if nd.Next != nil {
		nd.Next.Prev = nd.Prev
	}
	if nd.Prev != nil {
		nd.Prev.Next = nd.Next
	}
	if nd == q.Head {
		q.Head = nd.Next
		if q.Head != nil {
			q.Head.Prev = nil
		}
	}
	if nd == q.Tail {
		q.Tail = nd.Prev
		if q.Tail != nil {
			q.Tail.Next = nil
		}
	}
	nd.Prev = nil
	nd.Next = nil
	return nd.Item
}

type linkedItem struct {
	Prev *linkedItem
	Key  hash.Hash256
	Item interface{}
	Next *linkedItem
}

// Iter iterates queue items
func (q *LinkedQueue) Iter(fn func(keyMap hash.Hash256, v interface{})) {
	q.Lock()
	defer q.Unlock()

	cur := q.Head
	for cur != nil {
		fn(cur.Key, cur.Item)
		cur = cur.Next
	}
}
