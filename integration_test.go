package streams

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func expectStringStream(t *testing.T, src Stream[string], expected string, wantErr error) {
	var got []string
	for {
		val, err := src.Recv()
		if err != nil {
			if !errors.Is(err, wantErr) {
				t.Errorf("Expected error %q, got %q", wantErr, err.Error())
			}
			break
		}
		got = append(got, val)
	}
	if strings.Join(got, "") != expected {
		t.Errorf("Expected %s, got %s", expected, strings.Join(got, ""))
	}
}

func expectStream[T comparable](t *testing.T, src Stream[T], expected []T, wantErr error) {
	var got []T
	for {
		val, err := src.Recv()
		if err != nil {
			if !errors.Is(err, wantErr) {
				t.Errorf("Expected error %q, got %q", wantErr, err.Error())
			}
			break
		}
		got = append(got, val)
	}
	if len(got) != len(expected) {
		t.Errorf("Expected %d items, got %d: %v vs %v", len(expected), len(got), expected, got)
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("At index %d: expected %v, got %v", i, expected[i], v)
		}
	}
}

func TestIntegration1(t *testing.T) {
	t.Parallel()

	src := FromSlice([]string{"sta", "rt<|inq", "uiry|>h", "ello", "world", "<|/inq", "uiry|>", "end", "sdf<|diag", "nosis|>", "stop"})
	demux := NewSpecialTokenParserStream(src, []string{"<|inquiry|>", "<|/inquiry|>", "<|diagnosis|>"}).Demux()
	t.Run("test inquiry label", func(t *testing.T) {
		t.Parallel()
		expectStringStream(t, demux["<|inquiry|>"], "helloworld", io.EOF)
	})
	t.Run("test inquiry end label", func(t *testing.T) {
		t.Parallel()
		expectStringStream(t, demux["<|/inquiry|>"], "endsdf", io.EOF)
	})
	t.Run("test diagnosis label", func(t *testing.T) {
		t.Parallel()
		expectStringStream(t, demux["<|diagnosis|>"], "stop", io.EOF)
	})
	t.Run("default label", func(t *testing.T) {
		t.Parallel()
		expectStringStream(t, demux[""], "start", io.EOF)
	})

	src = FromSlice([]string{"start<|diag", "nosis|>", "unreach"})
	specialTokenStream := NewSpecialTokenParserStream(src, []string{"<|diagnosis|>"})
	subStream := SubstituteStream(specialTokenStream, func(frame LabeledChunk) Stream[LabeledChunk] {
		if frame.Label == "<|diagnosis|>" {
			return Map(FromSlice([]string{"我是", "咨询小结"}), func(s string) LabeledChunk { return LabeledChunk{Chunk: s} })
		}
		return nil
	})
	final := Map(subStream, func(frame LabeledChunk) string { return frame.Chunk })
	expectStringStream(t, final, "start我是咨询小结", io.EOF)
}
