package MRU

import (
	"container/list"
	"inmemorycache/abstract"
	"sync"
)

type CachedFunc[K comparable, V any] func(K) (V, error)

type MRU interface {
	abstract.Cache
}

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

func (mru *MRUCache) Put(key any, value any) {
	mru.mutex.Lock()
	defer mru.mutex.Unlock()

	if el, exists := mru.data[key]; exists {
		mru.elements.Remove(el)
		delete(mru.data, key)
	}

	if mru.elements.Len() == int(mru.capacity) {
		delete(mru.data, mru.elements.Front().Value.(*element).key)
		mru.elements.Remove(mru.elements.Front())
	}

	mru.elements.PushFront(&element{key: key, value: value})
	mru.data[key] = mru.elements.Front()
}

func (mru *MRUCache) Get(key any) any {
	mru.mutex.RLock()
	defer mru.mutex.RUnlock()
	if value, exists := mru.data[key]; exists {
		mru.elements.MoveToFront(value)
		return value.Value.(*element).value
	} else {
		return nil
	}
}

func (mru *MRUCache) Delete(key any) {
	mru.mutex.Lock()
	defer mru.mutex.Unlock()

	if el, exists := mru.data[key]; exists {
		mru.elements.Remove(el)
		delete(mru.data, el.Value.(*element).key)
	}
}

func (mru *MRUCache) Size() int {
	return len(mru.data)
}

func (mru *MRUCache) Contains(key any) bool {
	mru.mutex.RLock()
	defer mru.mutex.RUnlock()

	_, exists := mru.data[key]

	return exists
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
