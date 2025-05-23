package main

import (
	"InMemoryCache/LRU"
	. "fmt"
)

func sum(a int) (int, error) {
	return a + a, nil
}

func main() {
	Println("Hello from cache!")
	decorator := LRU.NewLRUDecorator[int, int](5)
	sumCached := decorator(sum)
	Println(sumCached(1))
	Println(sumCached(2))
	Println(sumCached(3))

	return
}
