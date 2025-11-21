package streams

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestSpecialTokenParserStream_Recv(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		tokens      []string
		want        []LabeledChunk
		wantErr     error
		description string
	}{
		{
			name:   "single token in one chunk",
			input:  []string{"hello<|end|>world"},
			tokens: []string{"<|end|>"},
			want: []LabeledChunk{
				{Label: "", Chunk: "hello"},
				{Label: "<|end|>", Chunk: ""},
				{Label: "<|end|>", Chunk: "world"},
			},
			wantErr: io.EOF,
		},
		{
			name:   "token split across chunks",
			input:  []string{"hello<|", "end|>world"},
			tokens: []string{"<|end|>"},
			want: []LabeledChunk{
				{Label: "", Chunk: "hello"},
				{Label: "<|end|>", Chunk: ""},
				{Label: "<|end|>", Chunk: "world"},
			},
			wantErr: io.EOF,
		},
		{
			name:   "multiple tokens",
			input:  []string{"a<|mid|>b<|end|>c"},
			tokens: []string{"<|mid|>", "<|end|>"},
			want: []LabeledChunk{
				{Label: "", Chunk: "a"},
				{Label: "<|mid|>", Chunk: ""},
				{Label: "<|mid|>", Chunk: "b"},
				{Label: "<|end|>", Chunk: ""},
				{Label: "<|end|>", Chunk: "c"},
			},
			wantErr: io.EOF,
		},
		{
			name:   "only one token",
			input:  []string{"<|end|>"},
			tokens: []string{"<|end|>"},
			want: []LabeledChunk{
				{Label: "<|end|>", Chunk: ""},
			},
			wantErr: io.EOF,
		},
		{
			name:    "empty input",
			input:   []string{},
			tokens:  []string{"<|end|>"},
			want:    nil,
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := FromSlice(tt.input)
			parser := NewSpecialTokenParserStream(src, tt.tokens)
			expectStream(t, parser, tt.want, tt.wantErr)
		})
	}
	t.Run("no token", func(t *testing.T) {
		input := []string{"hello", "world"}
		src := FromSlice(input)
		parser := NewSpecialTokenParserStream(src, []string{"<|end|>"})
		var got []string
		for {
			lc, err := parser.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					t.Fatalf("unexpected error: got %v, want %v", err, io.EOF)
				}
				break
			}
			if lc.Label != "" {
				t.Fatalf("unexpected label: got %v, want %v", lc.Label, "")
			}
			got = append(got, lc.Chunk)
		}

		if strings.Join(input, "") != strings.Join(got, "") {
			t.Fatalf("unexpected label: got %v, want %v", strings.Join(got, ""), strings.Join(input, ""))
		}
	})
}
