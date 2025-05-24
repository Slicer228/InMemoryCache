package LFU

import (
	"container/list"
	"fmt"
	"sync"
)

//In case of several equal access times values LFU algorithm uses MRU logic

type CachedFunc[K comparable, V any] func(K) (V, error)

type LFUCache struct {
	capacity               uint64                // max value of stored keys in cache
	data                   map[any]*list.Element // map for O(1) access time to cached value
	elements               *list.List            // Linked list for managing cached data
	mutex                  sync.RWMutex          // for synchronization
	maxAccessCountToUpdate uint64                // the maximum number of cached values received in a row before updating
}

type element struct {
	key      any
	value    any
	useCount uint64
}

func pushElemnetUp(elements *list.List, el *list.Element) {
	// service function to move up element in list, based of quantity of calls

	for prev := el.Prev(); ; prev = prev.Prev() {

		//check if cycle moved through all list
		if prev == nil {
			elements.MoveBefore(el, elements.Front())
			break
		}

		if el.Value.(*element).useCount > prev.Value.(*element).useCount {
			continue
		} else {
			elements.MoveBefore(el, prev.Next())
			break
		}
	}
}

func (lru *LFUCache) Put(key any, value any) {

	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	//check if key already exists to update value
	if el, exists := lru.data[key]; exists {
		el.Value.(*element).useCount++
		pushElemnetUp(lru.elements, el)
		el.Value.(*element).value = value
		fmt.Println("After PUT")
		for e := lru.elements.Front(); e != nil; e = e.Next() {
			fmt.Print(e.Value, " ")
		}
		fmt.Print("\n")
		return
	}

	//check if cache need to delete less used cache to set new
	if lru.elements.Len() == int(lru.capacity) {
		delete(lru.data, lru.elements.Back().Value.(*element).key)
		lru.elements.Remove(lru.elements.Back())
	}

	lru.elements.PushBack(&element{key: key, value: value, useCount: 1})
	lru.data[key] = lru.elements.Back()

}

func (lru *LFUCache) Get(key any) any {

	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	//check if key in cache exists to process receiving
	if el, exists := lru.data[key]; exists {

		//handling update of elem value
		if el.Value.(*element).useCount%lru.maxAccessCountToUpdate == 0 {
			return nil
		}

		el.Value.(*element).useCount++

		pushElemnetUp(lru.elements, el)

		return el.Value.(*element).value
	} else {
		return nil
	}
}

func NewLFUDecorator[K comparable, V any](capacity, maxAccessCountToUpdate uint64) func(CachedFunc[K, V]) CachedFunc[K, V] {
	cache := &LFUCache{
		capacity:               capacity,
		data:                   make(map[any]*list.Element),
		elements:               list.New(),
		mutex:                  sync.RWMutex{},
		maxAccessCountToUpdate: maxAccessCountToUpdate,
	}
	return func(f CachedFunc[K, V]) CachedFunc[K, V] {
		return func(parameter K) (V, error) {

			//check if key exists, in case when it time to update value GET method returns nil
			if val := cache.Get(parameter); val != nil {
				return val.(V), nil
			} else {
				result, err := f(parameter)

				//check if result has error, to avoid caching error-raising key
				if err != nil {
					return result, err
				}

				cache.Put(parameter, result)
				return result, err
			}
		}
	}
}

func New(capacity, maxAccessCountToUpdate uint64) *LFUCache {
	return &LFUCache{
		capacity:               capacity,
		data:                   make(map[any]*list.Element),
		elements:               list.New(),
		mutex:                  sync.RWMutex{},
		maxAccessCountToUpdate: maxAccessCountToUpdate,
	}
}
