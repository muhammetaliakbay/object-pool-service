package main

type ObjectSet struct {
	objects map[string]*Object
}

func NewObjectSet() *ObjectSet {
	return &ObjectSet{
		objects: make(map[string]*Object),
	}
}

func (set *ObjectSet) Add(object *Object) {
	set.objects[object.ID] = object
}

func (set *ObjectSet) Remove(object *Object) {
	delete(set.objects, object.ID)
}

func (set *ObjectSet) Has(object *Object) (has bool) {
	_, has = set.objects[object.ID]
	return
}

func (set *ObjectSet) Get(id string) (object *Object) {
	var has bool
	if object, has = set.objects[id]; !has {
		return nil
	}
	return
}
