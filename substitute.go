package streams

import "sync"

type substitutingStream[T any] struct {
	Stream[T]
	replace func(T) Stream[T]

	mu sync.Mutex
}

// SubstituteStream 对流中每个数据执行replace函数，如果replace响应的流不为空，则用响应的新流替换原流
// 注意事项：
// 1. 当流被替换之后，replace函数不会再被调用
// 2. 当流被替换之后，原流的数据不会撤销，因为下游已经接收到了
// 3. 当流被替换之后，原流的剩余的数据会被丢弃，如果上游是FromChan创建的流，则可能导致协程泄露，为了避免这种情况，需要业务方在replace函数内起一个协程来消费完原流的所有数据
func SubstituteStream[T any](src Stream[T], replace func(T) Stream[T]) Stream[T] {
	return &substitutingStream[T]{
		Stream:  avoidNil(src),
		replace: replace,
	}
}

func (s *substitutingStream[T]) Recv() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frame, err := s.Stream.Recv()
	if err != nil {
		return frame, err
	}

	if s.replace == nil {
		return frame, err
	}

	replacements := s.replace(frame)
	if replacements != nil {
		s.Stream = replacements
		return s.Stream.Recv()
	} else {
		return frame, err
	}
}
