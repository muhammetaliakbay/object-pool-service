package main

import "context"

type Session struct {
	instance  *Instance
	pool      *ObjectPool
	objects   *ObjectList
	limitLeft uint16
}

func NewSession(instance *Instance) *Session {
	return &Session{
		instance:  instance,
		pool:      nil,
		objects:   NewObjectList(),
		limitLeft: 0,
	}
}

var never = make(chan bool)

func (session *Session) Start(
	poolName string,
	messageChannel chan interface{},
	claimChannel chan []string,
	markerChannel chan uint32,
	ctx context.Context,
	limit uint16,
) {
	session.pool = session.instance.RefPool(poolName)
	session.limitLeft = limit
	go func() {
		defer session.end(poolName)
	loop:
		for {
			notifyChannel := never
			if session.limitLeft > 0 {
				ids := session.dequeue()
				if len(ids) > 0 {
					claimChannel <- ids
				}
				if session.limitLeft > 0 {
					notifyChannel = session.pool.NotifyChannel
				}
			}
			gotMarker, poolSize := session.tryClaimMarker()
			if gotMarker {
				markerChannel <- poolSize
			}
			select {
			case <-ctx.Done():
				break loop
			case msg := <-messageChannel:
				switch message := msg.(type) {
				case *QueueMessage:
					session.queue(message.Group, message.Objects)
				case *RequeueMessage:
					session.requeue(message.Objects)
				case *ReleaseMessage:
					session.release(message.Objects)
				case *MarkMessage:
					session.mark(message.Size)
				}
			case <-session.pool.MarkChannel:
			case <-notifyChannel:
			}
		}
	}()
}

func (session *Session) tryClaimMarker() (gotMarker bool, size uint32) {
	session.pool.Lock()
	if session.pool.marker == nil && session.pool.size <= session.pool.mark {
		session.pool.marker = session
		gotMarker = true
		size = session.pool.size
	} else {
		gotMarker = false
	}
	session.pool.Unlock()
	return
}

func (session *Session) end(poolName string) {
	objects := session.objects.MoveToArray()
	session.pool.Lock()
	for _, object := range objects {
		object.Pool.Requeue(object)
	}
	if session.pool.marker == session {
		session.pool.marker = nil
	}
	session.pool.Unlock()
	session.limitLeft = 0
	session.instance.UnrefPool(poolName)
}

func (session *Session) queue(group string, ids []string) (queuedIds []string) {
	if len(ids) == 0 {
		queuedIds = []string{}
		return
	}
	queuedIds = make([]string, len(ids))[0:0]
	session.pool.Lock()
	for _, id := range ids {
		object := session.pool.NewObject(id, group)
		if session.pool.Queue(object) {
			queuedIds = append(queuedIds, id)
		}
	}
	session.pool.Unlock()
	return
}

func (session *Session) dequeue() (ids []string) {
	session.pool.Lock()
	objects := session.pool.Dequeue(session.limitLeft)
	session.pool.Unlock()
	count := len(objects)
	session.limitLeft -= uint16(count)
	ids = make([]string, count)
	for i, object := range objects {
		ids[i] = object.ID
		session.objects.Append(object)
	}
	return
}

func (session *Session) requeue(ids []string) {
	if len(ids) == 0 {
		return
	}
	session.pool.Lock()
	for _, id := range ids {
		object := session.pool.objects.Get(id)
		if object != nil && session.objects.Has(object) {
			session.objects.Remove(object)
			session.pool.Requeue(object)
			session.limitLeft++
		}
	}
	session.pool.Unlock()
}

func (session *Session) release(ids []string) {
	if len(ids) == 0 {
		return
	}
	session.pool.Lock()
	for _, id := range ids {
		object := session.pool.objects.Get(id)
		if object != nil && session.objects.Has(object) {
			session.objects.Remove(object)
			session.pool.Release(object)
			session.limitLeft++
		}
	}
	session.pool.Unlock()
}

func (session *Session) mark(size uint32) {
	session.pool.Lock()
	if session.pool.marker == session {
		session.pool.mark = size
		session.pool.marker = nil
	}
	session.pool.Unlock()
}
