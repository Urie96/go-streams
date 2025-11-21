package streams

import (
	"io"
	"testing"
)

func TestConcatStream_Recv(t *testing.T) {
	t.Run("empty streams slice returns EOF immediately", func(t *testing.T) {
		stream := Concat[int]()
		expectStream(t, stream, []int{}, io.EOF)
	})

	t.Run("single stream with values", func(t *testing.T) {
		mock := FromSlice([]int{1, 2, 3})
		stream := Concat(mock)
		expectStream(t, stream, []int{1, 2, 3}, io.EOF)
	})

	t.Run("multiple streams with values", func(t *testing.T) {
		mock1 := FromSlice([]int{1, 2})
		mock2 := FromSlice([]int{3, 4})
		mock3 := FromSlice([]int{5})
		stream := Concat(mock1, mock2, mock3)
		expectStream(t, stream, []int{1, 2, 3, 4, 5}, io.EOF)
	})

	t.Run("mixed streams with some empty", func(t *testing.T) {
		empty := Empty[int]()
		mock1 := FromSlice([]int{1})
		empty2 := Empty[int]()
		mock2 := FromSlice([]int{2, 3})
		stream := Concat[int](empty, mock1, empty2, mock2)

		expectStream(t, stream, []int{1, 2, 3}, io.EOF)
	})
}

func TestFirstNonEmptyStream_Recv(t *testing.T) {
	t.Run("empty streams slice returns EOF immediately", func(t *testing.T) {
		stream := FirstNonEmpty[int]()
		expectStream(t, stream, []int{}, io.EOF)
	})

	t.Run("single stream with values", func(t *testing.T) {
		mock := FromSlice([]int{1, 2, 3})
		stream := FirstNonEmpty(mock)
		expectStream(t, stream, []int{1, 2, 3}, io.EOF)
	})

	t.Run("multiple streams with values", func(t *testing.T) {
		mock1 := FromSlice([]int{})
		mock2 := FromSlice([]int{3, 4})
		mock3 := FromSlice([]int{5})

		stream := FirstNonEmpty(mock1, mock2, mock3)

		expectStream(t, stream, []int{3, 4}, io.EOF)
	})

	t.Run("mixed streams with some empty", func(t *testing.T) {
		empty := Empty[int]()
		mock1 := FromSlice([]int{1})
		empty2 := Empty[int]()
		mock2 := FromSlice([]int{2, 3})
		stream := FirstNonEmpty[int](empty, mock1, empty2, mock2)
		expectStream(t, stream, []int{1}, io.EOF)
	})
}
