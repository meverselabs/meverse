package queue

type IKeyQueueItem interface {
	Key() interface{}
	iPriorityQueueItem
}

type kQueueItem struct {
	IPriorityQueueItem
}

func (k *kQueueItem) Key() interface{} {
	ik := k._unsafeSelf().(IKeyQueueItem)
	return ik.Key()
}

func NewKeyPriorityQueueItem(i IKeyQueueItem) IKeyQueueItem {
	ki := &kQueueItem{
		IPriorityQueueItem: NewPriorityQueueItem(i),
	}
	return ki
}

type IKeyPriorityQueue interface {
	Push(v IKeyQueueItem)
	Pop() IKeyQueueItem
	Peek() IKeyQueueItem
	Get(key interface{}) IKeyQueueItem
	RemoveKey(key interface{})
	Iter(f func(IKeyQueueItem) bool)
	Size() int
	Clear()
}

type keyPriorityQueue struct {
	que IPriorityQueue
	m   map[interface{}]IPriorityQueueItem
}

func NewKeyPriorityQueue(highest PriorityType) IKeyPriorityQueue {
	kq := &keyPriorityQueue{
		que: NewPriorityQueue(highest),
		m:   map[interface{}]IPriorityQueueItem{},
	}
	return kq
}

func (kq *keyPriorityQueue) Push(v IKeyQueueItem) {
	if _, has := kq.m[v.Key()]; has {
		return
	}
	pki := NewKeyPriorityQueueItem(v)
	pqi := pki.(IPriorityQueueItem)
	kq.que.Push(pqi)
	kq.m[v.Key()] = pqi
}

func (kq *keyPriorityQueue) Pop() IKeyQueueItem {
	if len(kq.m) == 0 {
		return nil
	}
	i := kq.que.Pop()
	item := i.(IKeyQueueItem)
	delete(kq.m, item.Key())

	return i._unsafeSelf().(IKeyQueueItem)
}

func (kq *keyPriorityQueue) Peek() IKeyQueueItem {
	if len(kq.m) == 0 {
		return nil
	}

	i := kq.que.Peek()
	return i._unsafeSelf().(IKeyQueueItem)
}

// Get item from map
func (kq *keyPriorityQueue) Get(key interface{}) IKeyQueueItem {
	pos, has := kq.m[key]
	if has {
		return pos._unsafeSelf().(IKeyQueueItem)
	}
	return nil
}

// RemoveKey is remove queue item by key
func (kq *keyPriorityQueue) RemoveKey(key interface{}) {
	item := kq.m[key]
	kq.que.Remove(item.Index())
	delete(kq.m, key)
}

// PopIter Iter iterates queue items Pop
func (kq *keyPriorityQueue) Iter(f func(IKeyQueueItem) bool) {
	for {
		item := kq.Pop()
		if item == nil {
			break
		}
		if !f(item) {
			break
		}
	}
}

// clear is remove all
func (kq *keyPriorityQueue) Size() int {
	return kq.que.Size()
}

// clear is remove all
func (kq *keyPriorityQueue) Clear() {
	kq.que.Clear()
	for i := range kq.m {
		delete(kq.m, i)
	}
}
