package streams

import "io"

type concatStream[T any] struct {
	streams            []Stream[T]
	index              int
	nonEmpty           bool
	firstNonEmptyUsage bool // true: 切换为首个非空流模式
}

func (m *concatStream[T]) Recv() (T, error) {
	for {
		if m.index >= len(m.streams) {
			var zero T
			return zero, io.EOF
		}
		stream := m.streams[m.index]
		value, err := stream.Recv()
		if err == io.EOF {
			if m.firstNonEmptyUsage && m.nonEmpty { // 首个非空流模式：第一个非空流已结束，整体可以结束了
				var zero T
				return zero, io.EOF
			}
			m.index++
			continue
		} else if err != nil {
			var zero T
			return zero, err
		}
		if m.firstNonEmptyUsage {
			m.nonEmpty = true
		}
		return value, nil
	}
}

// Concat 将多个流合并为一个流
func Concat[T any](streams ...Stream[T]) Stream[T] {
	return &concatStream[T]{
		streams: avoidNils(streams),
	}
}

// FirstNonEmpty 返回首个非空流，注意不是并发读，只有第一个流关闭了才会切换到下一个流
func FirstNonEmpty[T any](streams ...Stream[T]) Stream[T] {
	return &concatStream[T]{
		streams:            avoidNils(streams),
		firstNonEmptyUsage: true,
	}
}
