package main

import "sync"

type ObjectPool struct {
	objects              ObjectSet
	queuedObjects        ObjectList
	groupedQueuedObjects GroupedObjectList
	mu                   sync.Mutex
	NotifyChannel        chan bool
	MarkChannel          chan bool
	References           uint16
	size                 uint32
	mark                 uint32
	marker               *Session
}

func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		objects:              *NewObjectSet(),
		queuedObjects:        *NewObjectList(),
		groupedQueuedObjects: *NewGroupedObjectList(),
		mu:                   sync.Mutex{},
		NotifyChannel:        make(chan bool, 1),
		MarkChannel:          make(chan bool, 1),
		References:           0,
		size:                 0,
		mark:                 0,
		marker:               nil,
	}
}

func (pool *ObjectPool) Lock() {
	pool.mu.Lock()
}

func (pool *ObjectPool) Unlock() {
	select {
	case <-pool.NotifyChannel:
	default:
	}
	if pool.queuedObjects.length > 0 {
		pool.NotifyChannel <- true
	}
	select {
	case <-pool.MarkChannel:
	default:
	}
	if pool.marker == nil && pool.size <= pool.mark {
		pool.MarkChannel <- true
	}
	pool.mu.Unlock()
}

func (pool *ObjectPool) NewObject(id string, group string) *Object {
	return NewObject(id, group, pool)
}

func (pool *ObjectPool) Queue(object *Object) (queued bool) {
	if pool.objects.Has(object) {
		queued = false
		return
	}
	pool.objects.Add(object)
	pool.queuedObjects.Append(object)
	groupedQueue := pool.groupedQueuedObjects.GetList(object.Group)
	groupedQueue.Append(object)
	queued = true
	pool.size++
	return
}

func (pool *ObjectPool) Requeue(object *Object) {
	pool.queuedObjects.Append(object)
	groupedQueue := pool.groupedQueuedObjects.GetList(object.Group)
	groupedQueue.Append(object)
}

func (pool *ObjectPool) Release(object *Object) {
	pool.objects.Remove(object)
	pool.size--
}

func (pool *ObjectPool) Dequeue(limit uint16) (objects []*Object) {
	objects = make([]*Object, limit)[0:0]
	if limit == 0 {
		return
	}
	firstObject := pool.queuedObjects.RemoveFirst()
	if firstObject == nil {
		return objects
	}
	objects = append(objects, firstObject)
	group := firstObject.Group
	groupedList := pool.groupedQueuedObjects.GetList(group)
	groupedList.Remove(firstObject)
	if limit > 1 {
		for i := uint16(1); i < limit; i++ {
			object := groupedList.RemoveFirst()
			if object == nil {
				break
			}
			objects = append(objects, object)
			pool.queuedObjects.Remove(object)
		}
	}
	if groupedList.length == 0 {
		pool.groupedQueuedObjects.RemoveList(group)
	}
	return
}
