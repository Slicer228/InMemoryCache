package LFU

import (
	"container/list"
	"inmemorycache/abstract"
	"sync"
)

//In case of several equal access times values LFU algorithm uses MRU logic

type CachedFunc[K comparable, V any] func(K) (V, error)

type LFU interface {
	abstract.Cache
	migrateElementToNewBucket(el *list.Element)
	deleteLFUElement()
}

type LFUCache struct {
	capacity               uint64                // max value of stored keys in cache
	data                   map[any]*list.Element // map for O(1) access time to cached value
	elements               map[uint64]*list.List // Map of linked list for managing cached data key is count of usages, value is list of element structs
	mutex                  sync.RWMutex          // for synchronization
	maxAccessCountToUpdate uint64                // the maximum number of cached values received in a row before updating
	minKeyUsages           uint64                // minimum key of usage bucket
}

type element struct {
	key      any
	value    any
	useCount uint64
}

func (lfu *LFUCache) migrateElementToNewBucket(el *list.Element) {
	internalEl := el.Value.(*element)

	lfu.elements[internalEl.useCount].Remove(el)
	delete(lfu.data, internalEl.key)
	internalEl.useCount++

	// Create a new list if key doesn't exist
	if lfu.elements[internalEl.useCount] == nil {
		lfu.elements[internalEl.useCount] = list.New()
	}

	lfu.data[internalEl.key] = lfu.elements[internalEl.useCount].PushFront(internalEl)

	// Change minimal key of usages if last is empty now
	if lfu.elements[internalEl.useCount-1].Len() == 0 {
		delete(lfu.elements, internalEl.useCount-1)
		lfu.minKeyUsages = internalEl.useCount
	}
}

func (lfu *LFUCache) deleteLFUElement() {
	toDel := lfu.elements[lfu.minKeyUsages].Front()
	delete(lfu.data, toDel.Value.(*element).key)
	lfu.elements[lfu.minKeyUsages].Remove(toDel)
}

func (lfu *LFUCache) Put(key any, value any) {

	lfu.mutex.Lock()
	defer lfu.mutex.Unlock()

	//check if key already exists to update value
	if el, exists := lfu.data[key]; exists {
		el.Value.(*element).value = value
		lfu.migrateElementToNewBucket(el)
		return
	}

	//check if cache need to delete less used cache to set new
	if len(lfu.data) == int(lfu.capacity) {
		lfu.deleteLFUElement()
	}

	//check if bucket 1 exists else create
	if lfu.elements[1] == nil {
		lfu.elements[1] = list.New()
	}

	lfu.data[key] = lfu.elements[1].PushFront(&element{key: key, value: value, useCount: 1})
}

func (lfu *LFUCache) Get(key any) any {
	//return nil in case key not founded or value require update else return founded value

	lfu.mutex.RLock()
	defer lfu.mutex.RUnlock()

	//check if key in cache exists to process receiving
	if el, exists := lfu.data[key]; exists {

		lfu.migrateElementToNewBucket(el)

		//handling update of elem value
		if el.Value.(*element).useCount%lfu.maxAccessCountToUpdate == 0 {
			return nil
		}

		return el.Value.(*element).value
	} else {
		return nil
	}
}

func (lfu *LFUCache) Delete(key any) {
	lfu.mutex.Lock()
	defer lfu.mutex.Unlock()

	if el, exists := lfu.data[key]; exists {
		lfu.elements[el.Value.(*element).useCount].Remove(el)
		delete(lfu.data, el.Value.(*element).key)
	}
}

func (lfu *LFUCache) Size() int {
	return len(lfu.data)
}

func (lfu *LFUCache) Contains(key any) bool {
	lfu.mutex.RLock()
	defer lfu.mutex.RUnlock()

	_, exists := lfu.data[key]

	return exists
}

func NewLFUDecorator[K comparable, V any](capacity, maxAccessCountToUpdate uint64) func(CachedFunc[K, V]) CachedFunc[K, V] {
	cache := &LFUCache{
		capacity:               capacity,
		data:                   make(map[any]*list.Element),
		elements:               make(map[uint64]*list.List),
		mutex:                  sync.RWMutex{},
		maxAccessCountToUpdate: maxAccessCountToUpdate,
		minKeyUsages:           1,
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
		elements:               make(map[uint64]*list.List),
		mutex:                  sync.RWMutex{},
		maxAccessCountToUpdate: maxAccessCountToUpdate,
		minKeyUsages:           1,
	}
}
