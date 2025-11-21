package streams

import (
	"errors"
	"io"
	"testing"
)

func TestSubstituteStream_Recv(t *testing.T) {
	err1 := errors.New("test error")
	err2 := errors.New("replacement error")
	testCases := []struct {
		name     string
		src      Stream[int]
		replace  func(int) Stream[int]
		expected []int
		wantErr  error
		errMsg   string
	}{
		{
			name:     "no replacement",
			src:      FromSlice([]int{1, 2, 3}),
			replace:  func(item int) Stream[int] { return nil },
			expected: []int{1, 2, 3},
			wantErr:  io.EOF,
		},
		{
			name: "with replacement on second item",
			src:  FromSlice([]int{1, 2, 3}),
			replace: func(item int) Stream[int] {
				if item == 2 {
					return FromSlice([]int{20, 21, 22})
				}
				return nil
			},
			expected: []int{1, 20, 21, 22},
			wantErr:  io.EOF,
		},
		{
			name:     "original stream error",
			src:      FromErr[int](err1),
			replace:  func(item int) Stream[int] { return nil },
			expected: nil,
			wantErr:  err1,
			errMsg:   "test error",
		},
		{
			name: "replacement stream error",
			src:  FromSlice([]int{1}),
			replace: func(item int) Stream[int] {
				return FromErr[int](err2)
			},
			expected: nil,
			wantErr:  err2,
			errMsg:   "replacement error",
		},
		{
			name: "replacement stream error v2",
			src:  FromSlice([]int{1, 2, 3}),
			replace: func(item int) Stream[int] {
				if item == 2 {
					return FromErr[int](err2)
				}
				return nil
			},
			expected: []int{1},
			wantErr:  err2,
			errMsg:   "replacement error",
		},
		{
			name: "replacement on first item",
			src:  FromSlice([]int{1, 2, 3}),
			replace: func(item int) Stream[int] {
				if item == 1 {
					return FromSlice([]int{10, 11})
				}
				return nil
			},
			expected: []int{10, 11},
			wantErr:  io.EOF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stream := SubstituteStream(tc.src, tc.replace)
			expectStream(t, stream, tc.expected, tc.wantErr)
		})
	}

	t.Run("test", func(t *testing.T) {
		stream := SubstituteStream(FromSlice([]string{"start", "<|diagnosis|>", "end"}), func(item string) Stream[string] {
			if item == "<|diagnosis|>" {
				return FromSlice([]string{"我是", "咨询小结"})
			}
			return nil
		})
		expectStream(t, stream, []string{"start", "我是", "咨询小结"}, io.EOF)
	})
}
