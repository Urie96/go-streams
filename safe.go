package streams

import "sync"

type SafeStream[T any] struct {
	Stream[T]
	mu sync.Mutex
}

// ToSafe 将一个流转换为并发安全的流
func ToSafe[T any](stream Stream[T]) Stream[T] {
	if stream, ok := stream.(*SafeStream[T]); ok {
		return stream
	}
	return &SafeStream[T]{Stream: avoidNil(stream)}
}

func (s *SafeStream[T]) Recv() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Stream.Recv()
}
