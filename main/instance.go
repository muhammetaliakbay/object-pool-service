package main

import (
	"sync"
)

type Instance struct {
	pools map[string]*ObjectPool
	mu    sync.Mutex
}

func NewInstance() *Instance {
	return &Instance{
		pools: make(map[string]*ObjectPool),
		mu:    sync.Mutex{},
	}
}

func (instance *Instance) RefPool(name string) (pool *ObjectPool) {
	instance.mu.Lock()
	var ok bool
	if pool, ok = instance.pools[name]; !ok {
		pool = NewObjectPool()
		instance.pools[name] = pool
	}
	pool.References++
	instance.mu.Unlock()
	return
}

func (instance *Instance) UnrefPool(name string) (pool *ObjectPool) {
	instance.mu.Lock()
	defer instance.mu.Unlock()
	var ok bool
	if pool, ok = instance.pools[name]; !ok {
		return
	}
	pool.References--
	if pool.References == 0 {
		delete(instance.pools, name)
	}
	return
}
