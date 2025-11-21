package streams

import (
	"io"
	"testing"
)

func TestRemoveTokensStream_Recv(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		tokens   []string
		want     string
		wantErr  error
		inputErr error
	}{
		{
			name:    "no tokens",
			input:   []string{"hello", "world"},
			tokens:  []string{},
			want:    "helloworld",
			wantErr: io.EOF,
		},
		{
			name:    "single token removed",
			input:   []string{"foo", "bar", "baz"},
			tokens:  []string{"bar"},
			want:    "foobaz",
			wantErr: io.EOF,
		},
		{
			name:    "token across chunks",
			input:   []string{"a", "b", "c", "d"},
			tokens:  []string{"bc"},
			want:    "ad",
			wantErr: io.EOF,
		},
		{
			name:    "token across chunks 1",
			input:   []string{"00000", "ab", "cd"},
			tokens:  []string{"bc"},
			want:    "00000ad",
			wantErr: io.EOF,
		},
		{
			name:    "token across chunks 2",
			input:   []string{"a", "bc", "d"},
			tokens:  []string{"bc", "ab"},
			want:    "cd",
			wantErr: io.EOF,
		},
		{
			name:    "token across chunks 3",
			input:   []string{"00", "ab", "cd"},
			tokens:  []string{"bc", "ab"},
			want:    "00cd",
			wantErr: io.EOF,
		},
		{
			name:    "token across chunks 4",
			input:   []string{"ab", "cd"},
			tokens:  []string{"bc", "ab"},
			want:    "cd",
			wantErr: io.EOF,
		},
		{
			name:    "token across chunks 5",
			input:   []string{"ab", "cd"},
			tokens:  []string{"bc", "ad"},
			want:    "ad",
			wantErr: io.EOF,
		},
		{
			name:    "token across chunks 6",
			input:   []string{"abcd"},
			tokens:  []string{"bc", "ad"},
			want:    "ad",
			wantErr: io.EOF,
		},
		{
			name:    "multiple tokens",
			input:   []string{"prefix", "token1", "middle", "token2", "suffix"},
			tokens:  []string{"token1", "token2"},
			want:    "prefixmiddlesuffix",
			wantErr: io.EOF,
		},
		{
			name:    "EOF with leftover buffer",
			input:   []string{"left"},
			tokens:  []string{"foo"},
			want:    "left",
			wantErr: io.EOF,
		},
		{
			name: "",
			input: []string{
				`鉴于您的当前情况，`,
				`可以考虑使用一些缓解痉挛的药物，比如<`,
				`med>颠茄片<`,
				`/med>或<med>消旋山莨菪碱片</med>`,
				"<med>阿莫西林胶囊</med><inquiry>",
				"您出现这种症状已经持续多久了？</inquiry>",
			},
			tokens:  []string{"<med>", "</med>"},
			want:    "鉴于您的当前情况，可以考虑使用一些缓解痉挛的药物，比如颠茄片或消旋山莨菪碱片阿莫西林胶囊<inquiry>您出现这种症状已经持续多久了？</inquiry>",
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := FromSlice(tt.input)
			stream := NewRemoveTokensStream(mock, tt.tokens)
			expectStringStream(t, stream, tt.want, tt.wantErr)
		})
	}
}
