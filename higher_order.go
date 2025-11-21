package streams

type transformedStream[T any, R any] struct {
	Stream[T]
	mapper func(T, error) (R, error)
}

// Recv implements Stream.
func (h *transformedStream[T, R]) Recv() (R, error) {
	return h.mapper(h.Stream.Recv())
}

// Map 将流中的元素映射为另一个类型
func Map[T any, R any](src Stream[T], mapper func(T) R) Stream[R] {
	return MapErr(src, func(t T, err error) (R, error) {
		if err != nil {
			var zero R
			return zero, err
		}
		return mapper(t), nil
	})
}

func MapErr[T any, R any](src Stream[T], mapper func(T, error) (R, error)) Stream[R] {
	s := &transformedStream[T, R]{
		Stream: avoidNil(src),
		mapper: mapper,
	}
	return s
}

type filterStream[T any] struct {
	Stream[T]
	validate func(T) bool
}

func (f *filterStream[T]) Recv() (T, error) {
	for {
		v, err := f.Stream.Recv()
		if err != nil {
			var zero T
			return zero, err
		}
		if f.validate(v) {
			return v, nil
		}
	}
}

func Filter[T any](src Stream[T], validate func(T) bool) Stream[T] {
	return &filterStream[T]{
		Stream:   avoidNil(src),
		validate: validate,
	}
}
