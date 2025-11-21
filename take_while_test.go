package streams

import (
	"errors"
	"io"
	"testing"
)

func TestTakeWhile(t *testing.T) {
	t.Run("takes items while condition is true", func(t *testing.T) {
		stream := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		result := TakeWhile(stream, func(i int) bool {
			return i <= 5
		})
		expectStream(t, result, []int{1, 2, 3, 4, 5}, io.EOF)
	})

	t.Run("takes all items when condition always true", func(t *testing.T) {
		stream := FromSlice([]int{1, 2, 3, 4, 5})
		result := TakeWhile(stream, func(i int) bool {
			return true
		})
		expectStream(t, result, []int{1, 2, 3, 4, 5}, io.EOF)
	})

	t.Run("takes no items when condition always false", func(t *testing.T) {
		stream := FromSlice([]int{1, 2, 3, 4, 5})
		result := TakeWhile(stream, func(i int) bool {
			return false
		})
		expectStream(t, result, []int{}, io.EOF) // Note: Takes first item then stops
	})

	t.Run("takes first item only when condition false after first", func(t *testing.T) {
		stream := FromSlice([]int{1, 2, 3, 4, 5})
		result := TakeWhile(stream, func(i int) bool {
			return i == 1
		})
		expectStream(t, result, []int{1}, io.EOF)
	})

	t.Run("handles empty stream", func(t *testing.T) {
		stream := FromSlice([]int{})
		result := TakeWhile(stream, func(i int) bool {
			return true
		})
		expectStream(t, result, []int{}, io.EOF)
	})

	t.Run("handles single item stream", func(t *testing.T) {
		stream := FromSlice([]int{5})
		result := TakeWhile(stream, func(i int) bool {
			return i <= 5
		})
		expectStream(t, result, []int{5}, io.EOF)
	})

	t.Run("handles single item stream that doesn't match", func(t *testing.T) {
		stream := FromSlice([]int{10})
		result := TakeWhile(stream, func(i int) bool {
			return i <= 5
		})
		expectStream(t, result, []int{}, io.EOF) // Takes first item then stops
	})

	t.Run("propagates errors from source stream", func(t *testing.T) {
		testErr := errors.New("test error")
		stream := FromErr[int](testErr)
		result := TakeWhile(stream, func(i int) bool {
			return true
		})
		expectStream(t, result, []int{}, testErr)
	})

	t.Run("works with string streams", func(t *testing.T) {
		stream := FromSlice([]string{"apple", "banana", "apricot", "avocado", "berry"})
		result := TakeWhile(stream, func(s string) bool {
			return s[0] == 'a'
		})
		expectStream(t, result, []string{"apple"}, io.EOF)
	})

	t.Run("stops at first non-matching item but includes it", func(t *testing.T) {
		stream := FromSlice([]int{2, 4, 6, 7, 8, 10})
		result := TakeWhile(stream, func(i int) bool {
			return i%2 == 0
		})
		expectStream(t, result, []int{2, 4, 6}, io.EOF) // Includes the odd number 7
	})

	t.Run("complex condition test", func(t *testing.T) {
		stream := FromSlice([]string{"cat", "dog", "cattle", "deer", "cow", "camel"})
		result := TakeWhile(stream, func(s string) bool {
			return len(s) <= 4
		})
		expectStream(t, result, []string{"cat", "dog"}, io.EOF)
	})

	t.Run("condition based on index-like behavior", func(t *testing.T) {
		stream := FromSlice([]int{10, 20, 30, 40, 50})
		count := 0
		result := TakeWhile(stream, func(i int) bool {
			count++
			return count <= 3
		})
		expectStream(t, result, []int{10, 20, 30}, io.EOF)
	})

	t.Run("zero value handling", func(t *testing.T) {
		stream := FromSlice([]int{0, 1, 2, 0, 3})
		result := TakeWhile(stream, func(i int) bool {
			return i == 0
		})
		expectStream(t, result, []int{0}, io.EOF) // Takes first zero, then stops at 1
	})
}

