package streams

import (
	"io"
)

type Stream[T any] interface {
	Recv() (T, error)
}

type BaseStream[T any] struct {
	Stream[T]
}

func NewBaseStream[T any](stream Stream[T]) BaseStream[T] {
	if stream, ok := stream.(BaseStream[T]); ok { // 避免冗余包装
		return stream
	}
	return BaseStream[T]{Stream: avoidNil(stream)}
}

func (c BaseStream[T]) Consume(handle func(T) error) error {
	for {
		v, err := c.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if err := handle(v); err != nil {
			return err
		}
	}
}

func avoidNil[T any](stream Stream[T]) Stream[T] {
	if stream == nil {
		return Empty[T]()
	}
	return stream
}

func avoidNils[T any](ss []Stream[T]) []Stream[T] {
	for i, s := range ss {
		ss[i] = avoidNil(s)
	}
	return ss
}

// func (c BaseStream[T]) Iter() iter.Seq[T] {
// 	if c.Stream == nil {
// 		return func(yield func(T) bool) {}
// 	}
// 	return func(yield func(T) bool) {
// 		for {
// 			v, err := c.Recv()
// 			if err != nil {
// 				break
// 			}
// 			if !yield(v) {
// 				return
// 			}
// 		}
// 	}
// }

// Iter 返回一个迭代器，遇到错误或者流结束时返回
// 注意：这会忽略流中的错误，如果需要处理错误，请使用Consume
// func Iter[T any](stream Stream[T]) iter.Seq[T] {
// 	return NewBaseStream(stream).Iter()
// }

// Consume 阻塞接收流中的数据，并对每个数据调用handle函数
func Consume[T any](stream Stream[T], handle func(T) error) error {
	return NewBaseStream(stream).Consume(handle)
}

// ToChan 将流中的数据转换为channel，如果流中没有数据或者异常，则channel会立即关闭
// 注意事项：
// 1. 这会忽略流中的错误，如果需要处理错误，请使用Consume
// 2. 调用方需要确保channel被消费完毕，否则会导致goroutine泄漏
func ToChan[T any](stream Stream[T]) <-chan T {
	ch := make(chan T, 16)
	go func() {
		defer close(ch)
		Consume(stream, func(v T) error {
			ch <- v
			return nil
		})
	}()
	return ch
}

func Empty[T any]() Stream[T] {
	return FromFunc(func() (T, error) {
		var zero T
		return zero, io.EOF
	})
}

var EmptyStringStream = Empty[string]()
