package queue

import (
	"sync"

	"github.com/pkg/errors"
)

// Queue provides a basic queue ability with the peek method
type Queue struct {
	sync.Mutex
	pages []*queuePage
	size  int
}

// NewQueue returns a Queue
func NewQueue() *Queue {
	q := &Queue{
		pages: make([]*queuePage, 0, 256),
	}
	return q
}

// Push inserts a item at the bottom of the queue
func (q *Queue) Push(item interface{}) {
	q.Lock()
	defer q.Unlock()

	var page *queuePage
	if len(q.pages) == 0 {
		page = queuePagePool.Get().(*queuePage)
		q.pages = append(q.pages, page)
	} else {
		page = q.pages[len(q.pages)-1]
		if page.size == page.cap() {
			page = queuePagePool.Get().(*queuePage)
			q.pages = append(q.pages, page)
		}
	}
	page.push(item)
	q.size++
}

// Peek fetch the top item without removing it
func (q *Queue) Peek() interface{} {
	q.Lock()
	defer q.Unlock()

	if len(q.pages) == 0 {
		return nil
	}
	page := q.pages[0]
	item := page.peek()
	return item
}

// Pop returns a item at the top of the queue
func (q *Queue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	if len(q.pages) == 0 {
		return nil
	}
	page := q.pages[0]
	item := page.pop()
	if item != nil {
		if page.size == 0 && len(q.pages) > 1 {
			queuePagePool.Put(page)
			q.pages = q.pages[1:]
		}
		q.size--
	}
	return item
}

// Size returns the number of items
func (q *Queue) Size() int {
	q.Lock()
	defer q.Unlock()

	return q.size
}

var queuePagePool = sync.Pool{
	New: func() interface{} {
		return &queuePage{}
	},
}

type queuePage struct {
	queue [1024]interface{}
	head  int
	tail  int
	size  int
}

func (page *queuePage) push(item interface{}) error {
	if page.size >= page.cap() {
		return errors.New("full queue")
	}
	page.queue[page.tail] = item
	page.tail = (page.tail + 1) % page.cap()
	page.size++
	return nil
}

func (page *queuePage) peek() interface{} {
	if page.size == 0 {
		return nil
	}
	item := page.queue[page.head]
	return item
}

func (page *queuePage) pop() interface{} {
	if page.size == 0 {
		return nil
	}
	item := page.queue[page.head]
	page.head = (page.head + 1) % page.cap()
	page.size--
	return item
}

func (page *queuePage) cap() int {
	return len(page.queue)
}

// Iter iterates queue items
func (q *Queue) Iter(fn func(v interface{})) {
	q.Lock()
	defer q.Unlock()

	for _, page := range q.pages {
		page.iter(fn)
	}
}

// Iter iterates queue items
func (page *queuePage) iter(fn func(v interface{})) {
	for i := 0; i < page.size; i++ {
		item := page.queue[(page.head+i)%page.cap()]
		fn(item)
	}
}
