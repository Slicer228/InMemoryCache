package MRU

import (
	"container/list"
	"sync"
)

type CachedFunc[K comparable, V any] func(K) (V, error)

type MRUCache struct {
	capacity uint32
	data     map[any]*list.Element
	elements *list.List
	mutex    sync.RWMutex
}

type element struct {
	key   any
	value any
}

func (lru *MRUCache) Put(key any, value any) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if el, exists := lru.data[key]; exists {
		lru.elements.Remove(el)
		delete(lru.data, key)
	}

	if lru.elements.Len() == int(lru.capacity) {
		delete(lru.data, lru.elements.Front().Value.(*element).key)
		lru.elements.Remove(lru.elements.Front())
	}

	lru.elements.PushFront(&element{key: key, value: value})
	lru.data[key] = lru.elements.Front()
}

func (lru *MRUCache) Get(key any) any {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()
	if value, exists := lru.data[key]; exists {
		lru.elements.MoveToFront(value)
		return value.Value.(*element).value
	} else {
		return nil
	}
}

func NewMRUDecorator[K comparable, V any](capacity uint32) func(CachedFunc[K, V]) CachedFunc[K, V] {
	cache := &MRUCache{
		capacity: capacity,
		data:     make(map[any]*list.Element),
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

func New(capacity uint32) *MRUCache {
	return &MRUCache{
		capacity: capacity,
		data:     make(map[any]*list.Element),
		elements: list.New(),
		mutex:    sync.RWMutex{},
	}
}
