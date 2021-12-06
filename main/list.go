package main

type ListIndex uint32

type ObjectAnchor struct {
	next *Object
	prev *Object
}

func NewAnchor() *ObjectAnchor {
	return &ObjectAnchor{
		next: nil,
		prev: nil,
	}
}

type ObjectList struct {
	index  ListIndex
	first  *Object
	last   *Object
	length uint32
}

var nextObjectListIndex ListIndex = 0

func NewObjectList() (list *ObjectList) {
	list = &ObjectList{
		index: nextObjectListIndex,
		first: nil,
		last:  nil,
	}
	nextObjectListIndex += 1
	return
}

func (list *ObjectList) Append(object *Object) {
	list.length++
	anchor := object.GetAnchor(list)
	if list.last == nil {
		list.first = object
		list.last = object
	} else {
		lastAnchor := list.last.GetAnchor(list)
		lastAnchor.next = object
		anchor.prev = list.last
		list.last = object
	}
}

func (list *ObjectList) Remove(object *Object) {
	list.length--
	anchor := object.GetAnchor(list)
	if list.first == object {
		list.first = anchor.next
	} else {
		anchor.prev.GetAnchor(list).next = anchor.next
	}
	if list.last == object {
		list.last = anchor.prev
	} else {
		anchor.next.GetAnchor(list).prev = anchor.prev
	}
	object.RemoveAnchor(list)
}

func (list *ObjectList) RemoveFirst() (object *Object) {
	object = list.first
	if object == nil {
		return nil
	}
	list.length--
	anchor := object.GetAnchor(list)
	list.first = anchor.next
	if list.last == object {
		list.last = nil
	} else {
		anchor.next.GetAnchor(list).prev = nil
	}
	object.RemoveAnchor(list)
	return
}

func (list *ObjectList) Next(prev *Object) (next *Object) {
	if prev == nil {
		next = list.first
		return
	}
	next = prev.GetAnchor(list).next
	return
}

func (list *ObjectList) Has(object *Object) (has bool) {
	has = object.HasAnchor(list)
	return
}

func (list *ObjectList) CopyToArray() (arr []*Object) {
	arr = make([]*Object, list.length)
	index := uint32(0)
	for next := list.first; next != nil; next = list.Next(next) {
		arr[index] = next
		index++
	}
	return
}

func (list *ObjectList) MoveToArray() (arr []*Object) {
	arr = make([]*Object, list.length)
	index := uint32(0)
	for next := list.first; next != nil; next = list.Next(next) {
		arr[index] = next
		index++
	}
	list.first = nil
	list.last = nil
	list.length = 0
	for _, object := range arr {
		object.RemoveAnchor(list)
	}
	return
}
