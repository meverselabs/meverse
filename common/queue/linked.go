package queue

import (
	"github.com/fletaio/fleta/common/hash"
)

// LinkedQueue is designed to allow users to remove the item by the key
type LinkedQueue struct {
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

// Push inserts the item with the key at the bottom of the queue
func (q *LinkedQueue) Push(Key hash.Hash256, item interface{}) {
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
}

// Pop returns a item at the top of the queue
func (q *LinkedQueue) Pop() interface{} {
	if q.Head == nil {
		return nil
	}
	nd := q.Head
	if nd == q.Tail {
		q.Head = nil
		q.Tail = nil
	} else {
		q.Head = nd.Next
		nd.Next.Prev = nil
	}
	nd.Prev = nil
	nd.Next = nil
	delete(q.keyMap, nd.Key)
	return nd.Item
}

// Remove deletes a item by the key
func (q *LinkedQueue) Remove(Key hash.Hash256) interface{} {
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
	}
	if nd == q.Tail {
		q.Tail = nd.Prev
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
