package main

type GroupedObjectList struct {
	listsByGroup map[string]*ObjectList
}

func NewGroupedObjectList() (grouped *GroupedObjectList) {
	return &GroupedObjectList{
		listsByGroup: make(map[string]*ObjectList),
	}
}

func (grouped *GroupedObjectList) GetList(group string) (list *ObjectList) {
	var ok bool
	if list, ok = grouped.listsByGroup[group]; ok {
		return
	}
	list = NewObjectList()
	grouped.listsByGroup[group] = list
	return list
}

func (grouped *GroupedObjectList) RemoveList(group string) {
	delete(grouped.listsByGroup, group)
}
