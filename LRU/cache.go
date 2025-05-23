package LRU

import (
	"container/list"
	"sync"
)

type CachedFunc[K comparable, V any] func(K) (V, error)

type LRUCache struct {
	capacity uint32
	data     map[interface{}]interface{}
	elements *list.List
	mutex    sync.RWMutex
}

func (lru *LRUCache) Put(key interface{}, value interface{}) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()
	if _, exists := lru.data[key]; exists || lru.elements.Len() == int(lru.capacity) {
		delete(lru.data, lru.elements.Back().Value.(interface{}))
		lru.elements.Remove(lru.elements.Back())
	}

	lru.elements.PushFront(key)
	lru.data[key] = value
}

func (lru *LRUCache) Get(key interface{}) interface{} {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()
	if value, exists := lru.data[key]; exists {
		lru.elements.MoveToFront(&list.Element{Value: value})
		return value
	} else {
		return nil
	}
}

func NewLRUDecorator[K comparable, V any](capacity uint32) func(CachedFunc[K, V]) CachedFunc[K, V] {
	cache := &LRUCache{
		capacity: capacity,
		data:     make(map[interface{}]interface{}),
		elements: list.New(),
		mutex:    sync.RWMutex{},
	}
	return func(f CachedFunc[K, V]) CachedFunc[K, V] {
		return func(parameter K) (V, error) {
			if val := cache.Get(parameter); val != nil {
				return val.(V), nil
			} else {
				result, err := f(parameter)
				if err != nil {
					return result, err
				}
				cache.Put(parameter, result)
				return result, err
			}
		}
	}
}

func New(capacity uint32) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		data:     make(map[interface{}]interface{}),
		elements: list.New(),
		mutex:    sync.RWMutex{},
	}
}
