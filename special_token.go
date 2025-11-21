package streams

import (
	"io"
	"strings"
	"unicode/utf8"
)

type specialTokenParserStream struct {
	Stream[string]
	specialTokens []string
	minBufferLen  int

	lastToken string
	buffer    string
}

type LabeledChunk struct {
	Label string
	Chunk string
}

func NewSpecialTokenParserStream(src Stream[string], specialTokens []string) *specialTokenParserStream {
	minBufferLen := 0
	for _, lb := range specialTokens {
		if utf8.RuneCountInString(lb) > minBufferLen {
			minBufferLen = len(lb)
		}
	}

	return &specialTokenParserStream{
		Stream:        avoidNil(src),
		specialTokens: specialTokens,
		minBufferLen:  minBufferLen,
	}
}

func (s *specialTokenParserStream) cutOverflowBuffer() (label string, chunk string) {
	if utf8.RuneCountInString(s.buffer) <= s.minBufferLen {
		return "", ""
	}

	for _, st := range s.specialTokens {
		index := strings.Index(s.buffer, st)
		if index == -1 {
			continue
		}
		// 遇到了某个特殊标记
		prelastToken := s.lastToken
		s.lastToken = st
		chunkBeforeToken := s.buffer[:index]
		s.buffer = s.buffer[index:]
		if len(chunkBeforeToken) > 0 {
			return prelastToken, chunkBeforeToken
		} else {
			// buffer以这个标记开头
			s.buffer = s.buffer[len(st):]
			return st, ""
		}
	}

	index := utf8.RuneCountInString(s.buffer) - s.minBufferLen

	if index > 0 {
		bufferRunes := []rune(s.buffer)
		beforeChunk := string(bufferRunes[:index])
		s.buffer = string(bufferRunes[index:])
		return s.lastToken, beforeChunk
	}

	return "", ""
}

func (s *specialTokenParserStream) Recv() (LabeledChunk, error) {
	for {
		label, chunk := s.cutOverflowBuffer() // 先处理buffer中的数据
		if chunk != "" || label != "" {
			return LabeledChunk{
				Label: label,
				Chunk: chunk,
			}, nil
		}

		upChunk, err := s.Stream.Recv()
		if err == io.EOF { // 上游已经读完了，但是buffer中还有数据，需要处理
			s.minBufferLen = 0
			label, chunk := s.cutOverflowBuffer()
			if chunk == "" && label == "" {
				return LabeledChunk{}, io.EOF
			} else {
				return LabeledChunk{
					Label: label,
					Chunk: chunk,
				}, nil
			}
		} else if err != nil {
			return LabeledChunk{}, err
		}
		s.buffer += upChunk
	}
}

func (s *specialTokenParserStream) Demux() map[string]Stream[string] {
	labels := make([]string, len(s.specialTokens))
	for _, st := range s.specialTokens {
		labels = append(labels, st)
	}

	tmp := Demux(s, func(s LabeledChunk) string { return s.Label }, labels)
	res := make(map[string]Stream[string], len(labels))
	for k, v := range tmp {
		res[k] = Map(v, func(s LabeledChunk) string { return s.Chunk })
	}
	return res
}
