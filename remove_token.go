package streams

import (
	"io"
	"strings"
	"unicode/utf8"
)

type removeTokensStream struct {
	src Stream[string]

	tokens       []string
	buffer       string
	minBufferLen int
}

// 注意事项:
//  1. 如果token之间有重叠，取最先遇到的token，比如：tokens是["bc","ab"]，流接收到的是["abcd"]，那么过滤之后会输出"cd"
//  2. 如果token过滤之后，前后刚好又形成新的过滤token，那么不会再次移除。比如：tokens是["bc","ad"]，流接收到的是["abcd"]，那么过滤之后仍会输出"ad"
func NewRemoveTokensStream(src Stream[string], tokens []string) Stream[string] {
	filterToken := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token != "" {
			filterToken = append(filterToken, token)
		}
	}
	tokens = filterToken
	if len(tokens) == 0 {
		return src
	}

	minBufferLen := 0
	for _, lb := range tokens {
		if utf8.RuneCountInString(lb) > minBufferLen {
			minBufferLen = len(lb)
		}
		if utf8.RuneCountInString(lb) > minBufferLen {
			minBufferLen = len(lb)
		}
	}

	return &removeTokensStream{
		src:          avoidNil(src),
		tokens:       tokens,
		minBufferLen: minBufferLen,
	}
}

func (s *removeTokensStream) cutOverflowBuffer() string {
	if utf8.RuneCountInString(s.buffer) <= s.minBufferLen {
		return ""
	}

	index := -1
	label := ""
	for _, l := range s.tokens {
		i := strings.Index(s.buffer, l)
		if i >= 0 {
			if index == -1 {
				index = i
				label = l
			} else if i < index {
				index = i
				label = l
			}
		}
	}

	if index >= 0 {
		res := s.buffer[:index]
		s.buffer = s.buffer[index+len(label):]
		return res
	}

	index = utf8.RuneCountInString(s.buffer) - s.minBufferLen

	if index > 0 {
		bufferRunes := []rune(s.buffer)
		beforeChunk := string(bufferRunes[:index])
		s.buffer = string(bufferRunes[index:])
		return beforeChunk
	}

	return ""
}

func (s *removeTokensStream) Recv() (string, error) {
	for {
		chunk := s.cutOverflowBuffer()
		if chunk != "" {
			return chunk, nil
		}

		upChunk, err := s.src.Recv()
		if err == io.EOF { // 上游已经读完了，但是buffer中还有数据，需要处理
			s.minBufferLen = 0
			chunk := s.cutOverflowBuffer()
			if chunk != "" {
				return chunk, nil
			} else {
				return "", io.EOF
			}
		} else if err != nil {
			return "", err
		}
		s.buffer += upChunk
	}
}

// RemoveLabels 移除流中的labels，返回一个新的流
func RemoveLabels(src Stream[string], labels []string) Stream[string] {
	return NewRemoveTokensStream(src, labels)
}
