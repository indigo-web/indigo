package pool

// ObjectPool is a generic analogue of sync.Pool, except it does not provide
// thread-safety. Concurrent access may lead to UB
type ObjectPool[T any] struct {
	queue []T
}

func NewObjectPool[T any](queueSize int) ObjectPool[T] {
	return ObjectPool[T]{
		queue: make([]T, 0, queueSize),
	}
}

func (o *ObjectPool[T]) Acquire() (obj T) {
	if len(o.queue) != 0 {
		obj = o.queue[len(o.queue)-1]
		o.queue = o.queue[:len(o.queue)-1]
	}

	return obj
}

func (o *ObjectPool[T]) Release(obj T) {
	if len(o.queue) == cap(o.queue) {
		return
	}

	o.queue = append(o.queue, obj)
}
