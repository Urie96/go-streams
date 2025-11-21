package streams

import "io"

func TakeWhile[T any](src Stream[T], canTake func(packet T) bool) Stream[T] {
	var zero T
	end := false

	return FromFunc(func() (T, error) {
		for {
			if end {
				return zero, io.EOF
			}
			v, err := src.Recv()
			if err != nil {
				return v, err
			}
			if canTake(v) {
				return v, nil
			}
			end = true
		}
	})
}
