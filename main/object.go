package main

type Object struct {
	ID      string
	Group   string
	Pool    *ObjectPool
	anchors map[ListIndex]*ObjectAnchor
}

func NewObject(id string, group string, pool *ObjectPool) *Object {
	return &Object{
		ID:      id,
		Group:   group,
		Pool:    pool,
		anchors: make(map[ListIndex]*ObjectAnchor),
	}
}

func (object *Object) GetAnchor(list *ObjectList) (anchor *ObjectAnchor) {
	var ok bool
	if anchor, ok = object.anchors[list.index]; ok {
		return
	}
	anchor = NewAnchor()
	object.anchors[list.index] = anchor
	return anchor
}

func (object *Object) HasAnchor(list *ObjectList) (has bool) {
	_, has = object.anchors[list.index]
	return
}

func (object *Object) RemoveAnchor(list *ObjectList) {
	delete(object.anchors, list.index)
}
