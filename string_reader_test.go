package streams

import (
	"io"
	"testing"
)

func TestStringReader_ReadUntil(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		delims  []string
		input   []string
		want    []string
		wantErr error
	}{
		{
			name:    "single delim hit",
			input:   []string{"hello", "world\nfoo"},
			delims:  []string{"\n"},
			want:    []string{"helloworld\n", "foo"},
			wantErr: io.EOF,
		},
		{
			name:    "empty delim rune split",
			input:   []string{"你好", "世界"},
			delims:  []string{""},
			want:    []string{"你", "好", "世", "界"},
			wantErr: io.EOF,
		},
		{
			name:    "multi delim",
			input:   []string{"foo&bar", "baz|end"},
			delims:  []string{"&", "|"},
			want:    []string{"foo&", "barbaz|", "end"},
			wantErr: io.EOF,
		},
		{
			name:    "no delim until EOF",
			input:   []string{"abc", "def"},
			delims:  []string{},
			want:    []string{"abcdef"},
			wantErr: io.EOF,
		},
		{
			name:    "EOF with leftover",
			input:   []string{"left1", "left2", "left3"},
			delims:  []string{"\n"},
			want:    []string{"left1left2left3"},
			wantErr: io.EOF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := NewStringReader(FromSlice(tt.input))
			expectStream(t, sr.ToSplitReader(tt.delims), tt.want, tt.wantErr)
		})
	}
}

func TestOnceStreamStream(t *testing.T) {
	input := OnceStringStream(FromSlice([]string{"a", "b", "c"}))
	expectStream(t, input, []string{"abc"}, io.EOF)
}
