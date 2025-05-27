package abstract

type Cache interface {
	Get(key any) any
	Put(key any, value any)
	Delete(key any)
	Contains(key any) bool
	Size() int
}
