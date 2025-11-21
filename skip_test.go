package streams

import (
	"io"
	"testing"
)

func TestSkipUntil(t *testing.T) {
	t.Run("在中间遇到", func(t *testing.T) {
		stream := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		stream = SkipUntil(stream, func(i int) bool {
			return i == 5
		})
		expectStream(t, stream, []int{5, 6, 7, 8, 9, 10}, io.EOF)
	})

	t.Run("一直没遇到", func(t *testing.T) {
		stream := FromSlice([]int{1, 2, 3, 4})
		stream = SkipUntil(stream, func(i int) bool {
			return i == 5
		})
		expectStream(t, stream, []int{}, io.EOF)
	})

	t.Run("首位就遇到", func(t *testing.T) {
		stream := FromSlice([]int{1, 2, 3, 4})
		stream = SkipUntil(stream, func(i int) bool {
			return i == 1
		})
		expectStream(t, stream, []int{1, 2, 3, 4}, io.EOF)
	})
}

func TestSkipN(t *testing.T) {
	t.Run("N = 0", func(t *testing.T) {
		stream := FromSlice([]string{"a", "b", "c", "d", "e"})
		stream = SkipN(stream, 0)
		expectStream(t, stream, []string{"a", "b", "c", "d", "e"}, io.EOF)
	})

	t.Run("N 在中间", func(t *testing.T) {
		stream := FromSlice([]string{"a", "b", "c", "d", "e"})
		stream = SkipN(stream, 2)
		expectStream(t, stream, []string{"c", "d", "e"}, io.EOF)
	})

	t.Run("N 在末尾", func(t *testing.T) {
		stream := FromSlice([]string{"a", "b", "c", "d", "e"})
		stream = SkipN(stream, 5)
		expectStream(t, stream, []string{}, io.EOF)
	})

	t.Run("N 大于长度", func(t *testing.T) {
		stream := FromSlice([]string{"a", "b", "c", "d", "e"})
		stream = SkipN(stream, 10)
		expectStream(t, stream, []string{}, io.EOF)
	})
}
