package TTL

import "inmemorycache/abstract"

type CachedFunc[K comparable, V any] func(K) (V, error)

type TTL interface {
	abstract.Cache
	launchWorker() // Launch goroutine to manage outdated cache
}
