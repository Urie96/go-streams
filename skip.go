package streams

// SkipUntil 返回某个包之后的尾部流，尾部流的首包一定是那个满足条件的包
func SkipUntil[T any](s Stream[T], f func(T) bool) Stream[T] {
	hasSkipped := false
	return FromFunc(func() (T, error) {
		var zero T
		if hasSkipped {
			return s.Recv()
		}
		for {
			v, err := s.Recv()
			if err != nil {
				return zero, err
			}
			if f(v) {
				hasSkipped = true
				return v, nil
			}
		}
	})
}

// SkipN 返回跳过n个包之后的流
func SkipN[T any](s Stream[T], n int) Stream[T] {
	i := 0
	return FromFunc(func() (T, error) {
		var zero T
		for {
			v, err := s.Recv()
			if err != nil {
				return zero, err
			}
			if i >= n {
				return v, nil
			}
			i++
		}
	})
}
