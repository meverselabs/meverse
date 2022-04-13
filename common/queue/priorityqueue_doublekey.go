package queue

type IDoubleKeyQueueItem interface {
	IKeyQueueItem
	DKey() interface{}
	DPriority() uint64
}

type IDoubleKeyQueue interface {
	Push(v IDoubleKeyQueueItem)
	Pop() IDoubleKeyQueueItem
	Peek() IDoubleKeyQueueItem
	DPeek() IKeyPriorityQueue
	Get(key interface{}) IDoubleKeyQueueItem
	DGet(key interface{}) IKeyPriorityQueue
	RemoveKey(key interface{})
	List() []IDoubleKeyQueueItem
	Iter(f func(IDoubleKeyQueueItem) bool)
	Size() int
	Clear()
}

// A PriorityQueue implements heap.Interface and holds Items.
type doubleKeyPriorityQueue struct {
	IKeyPriorityQueue
	dKeyMap      map[interface{}]interface{}
	PriorityType PriorityType
}

type queItem struct {
	IKeyPriorityQueue
	IDoubleKeyQueueItem
}

func newQueItem(pt PriorityType, v IDoubleKeyQueueItem) IDoubleKeyQueueItem {
	return &queItem{
		NewKeyPriorityQueue(pt),
		v,
	}
}

func (pi *queItem) Key() interface{} {
	return pi.DKey()
}
func (pi *queItem) Priority() uint64 {
	return pi.DPriority()
}

func NewDoubleKeyPriorityQueue(pt PriorityType) IDoubleKeyQueue {
	kq := &doubleKeyPriorityQueue{
		IKeyPriorityQueue: NewKeyPriorityQueue(pt),
		dKeyMap:           map[interface{}]interface{}{},
		PriorityType:      pt,
	}
	return kq
}

func (kq *doubleKeyPriorityQueue) Push(v IDoubleKeyQueueItem) {
	q := kq.IKeyPriorityQueue.Get(v.DKey())
	if q == nil {
		q = newQueItem(kq.PriorityType, v)
		kq.IKeyPriorityQueue.Push(q)
	}
	iq := q.(IKeyPriorityQueue)
	iq.Push(v)
	kq.dKeyMap[v.Key()] = v.DKey()
}

func (kq *doubleKeyPriorityQueue) Pop() IDoubleKeyQueueItem {
	q := kq.IKeyPriorityQueue.Peek()
	if q == nil {
		return nil
	}
	que := q.(IKeyPriorityQueue)
	t := que.Pop()
	if que.Size() == 0 {
		kq.IKeyPriorityQueue.Pop()
	}
	item := t.(IDoubleKeyQueueItem)
	delete(kq.dKeyMap, item.Key())
	return item
}

func (kq *doubleKeyPriorityQueue) Peek() IDoubleKeyQueueItem {
	q := kq.IKeyPriorityQueue.Peek()
	if q == nil {
		return nil
	}
	return q.(*queItem).IKeyPriorityQueue.Peek().(IDoubleKeyQueueItem)
}

// Get item from map
func (kq *doubleKeyPriorityQueue) DPeek() IKeyPriorityQueue {
	return kq.IKeyPriorityQueue.Peek().(IKeyPriorityQueue)
}

// Get item from map
func (kq *doubleKeyPriorityQueue) Get(key interface{}) IDoubleKeyQueueItem {
	dkey, has := kq.dKeyMap[key]
	if !has {
		return nil
	}
	q := kq.IKeyPriorityQueue.Get(dkey)
	if q == nil {
		panic("q is nil")
	}
	ikq := q.(IKeyPriorityQueue)
	return ikq.Get(key).(IDoubleKeyQueueItem)
}

// Get item from map
func (kq *doubleKeyPriorityQueue) DGet(dkey interface{}) IKeyPriorityQueue {
	return kq.IKeyPriorityQueue.Get(dkey).(IKeyPriorityQueue)
}

// Delete key item from queue.
func (kq *doubleKeyPriorityQueue) RemoveKey(key interface{}) {
	dkey, has := kq.dKeyMap[key]
	if !has {
		return
	}
	idk := kq.Get(dkey)
	if idk == nil {
		return
	}
	ikq := idk.(IKeyPriorityQueue)
	ikq.RemoveKey(key)
	if ikq.Size() == 0 {
		kq.IKeyPriorityQueue.RemoveKey(dkey)
	}
	delete(kq.dKeyMap, key)
}

// Iter iterates queue items
func (kq *doubleKeyPriorityQueue) Iter(f func(IDoubleKeyQueueItem) bool) {
	for {
		item := kq.Peek()
		if item == nil {
			break
		}
		if !f(item) {
			break
		}
		kq.Pop()
	}
}

// Size return of map
func (kq *doubleKeyPriorityQueue) Size() int {
	return len(kq.dKeyMap)
}

// clear is remove all
func (kq *doubleKeyPriorityQueue) Clear() {
	for {
		q := kq.Pop().(*queItem)
		if q == nil {
			break
		}
		q.Clear()
	}
	kq.IKeyPriorityQueue.Clear()
	kq.dKeyMap = nil
}

// List returns all items
func (kq *doubleKeyPriorityQueue) List() []IDoubleKeyQueueItem {
	list := []IDoubleKeyQueueItem{}
	for k, _ := range kq.dKeyMap {
		list = append(list, kq.Get(k))
	}
	return list
}
