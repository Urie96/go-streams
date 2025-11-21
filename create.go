package streams

import (
	"io"
	"sync"
)

type funcStream[T any] struct {
	fn func() (T, error)
}

func (f *funcStream[T]) Recv() (T, error) {
	return f.fn()
}

// FromFunc 通过一个Recv函数创建流
func FromFunc[T any](fn func() (T, error)) Stream[T] {
	return &funcStream[T]{fn: fn}
}

// FromChan 从chan中创建一个流
func FromChan[T any](ch <-chan T) Stream[T] {
	return FromFunc(func() (T, error) {
		v, ok := <-ch
		if !ok {
			var zero T
			return zero, io.EOF
		}
		return v, nil
	})
}

// FromSlice 从slice中创建一个流
func FromSlice[T any](slice []T) Stream[T] {
	return FromFunc(func() (T, error) {
		if len(slice) == 0 {
			var zero T
			return zero, io.EOF
		}
		v := slice[0]
		slice = slice[1:]
		return v, nil
	})
}

func FromErr[T any](err error) Stream[T] {
	return FromFunc(func() (T, error) {
		var zero T
		return zero, err
	})
}

// FromFutureStream 接收一个Stream的chan，在未来消费时，才会从chan中取出这个Stream并消费。对于创建流耗时较长的场景十分有用。
// 注意，只会从这个chan中取一个Stream。强烈建议chan用make([]Stream[T], 1)定义，避免生产方协程泄露。
func FromFutureStream[T any](ch <-chan Stream[T]) Stream[T] {
	var mu sync.Mutex
	var src Stream[T]

	return FromFunc(func() (T, error) {
		mu.Lock()
		defer mu.Unlock()
		if src == nil {
			src = avoidNil(<-ch)
		}
		return src.Recv()
	})
}
