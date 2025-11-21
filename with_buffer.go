package streams

import (
	"sync"
)

func WithBuffer[T any](src Stream[T]) Stream[T] {
	type valWithErr struct {
		val T
		err error
	}
	queue := &ConcurrentQueue[valWithErr]{}
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	go func() {
		for {
			val, err := src.Recv()
			queue.Push(valWithErr{val: val, err: err})
			cond.Signal()
			if err != nil {
				return
			}
		}
	}()

	return FromFunc(func() (T, error) {
		mu.Lock()
		defer mu.Unlock()

		for {
			val, ok := queue.Pop()
			if ok {
				return val.val, val.err
			}
			cond.Wait()
		}
	})
}
