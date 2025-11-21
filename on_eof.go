package streams

import "io"

func OnEOF[T any](src Stream[T], f func()) Stream[T] {
	return FromFunc(func() (T, error) {
		p, err := src.Recv()
		if err == io.EOF {
			f()
		}
		return p, err
	})
}
