package streams

import (
	"errors"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

type SLabel struct {
	Name       string
	StartToken string
	EndToken   string
}

type labelStream struct {
	src    Stream[string]
	labels []SLabel

	buffer       string
	minBufferLen int
	currentLabel *SLabel
}

func NewLabelStream(src Stream[string], labels []SLabel) *labelStream {
	if err := checkLabels(labels); err != nil {
		panic(err) // labels是写在代码里的，所以panic相对安全，会在开发阶段暴露
	}

	minBufferLen := 0
	for _, lb := range labels {
		if utf8.RuneCountInString(lb.StartToken) > minBufferLen {
			minBufferLen = len(lb.StartToken)
		}
		if utf8.RuneCountInString(lb.EndToken) > minBufferLen {
			minBufferLen = len(lb.EndToken)
		}
	}

	return &labelStream{
		src:          avoidNil(src),
		labels:       labels,
		minBufferLen: minBufferLen,
	}
}

func (s *labelStream) cutOverflowBuffer() (label string, chunk string) {
	if utf8.RuneCountInString(s.buffer) <= s.minBufferLen {
		return "", ""
	}

	if s.currentLabel != nil {
		if s.currentLabel.EndToken == "" {
			chunk := s.buffer
			s.minBufferLen = 0
			s.buffer = ""
			return s.currentLabel.Name, chunk
		}
		index := strings.Index(s.buffer, s.currentLabel.EndToken)
		if index >= 0 {
			chunkBeforeToken := s.buffer[:index]
			s.buffer = s.buffer[index:]
			lastLabel := s.currentLabel
			s.currentLabel = nil
			return lastLabel.Name, chunkBeforeToken
		}
	}

	for i, lb := range s.labels {
		index := strings.Index(s.buffer, lb.StartToken)
		if index == -1 {
			continue
		}
		chunkBeforeToken := s.buffer[:index]
		s.buffer = s.buffer[index+len(lb.StartToken):]

		lbCopy := lb
		s.currentLabel = &lbCopy
		s.labels = slices.Delete(s.labels, i, i+1)
		if s.currentLabel.EndToken == "" { // 这个label没有endToken，所以后面所有的chunk都属于这个label
			s.minBufferLen = 0
		}
		if len(chunkBeforeToken) > 0 {
			return "", chunkBeforeToken
		} else {
			return s.cutOverflowBuffer()
		}
	}

	index := utf8.RuneCountInString(s.buffer) - s.minBufferLen

	if index > 0 {
		bufferRunes := []rune(s.buffer)
		beforeChunk := string(bufferRunes[:index])
		s.buffer = string(bufferRunes[index:])
		labelName := ""
		if s.currentLabel != nil {
			labelName = s.currentLabel.Name
		}
		return labelName, beforeChunk
	}

	return "", ""
}

// 需要注意的是:
// - labels是从前往后匹配，所以越前面的label优先级越高，因此传入labels的顺序应该尽量和流式文本中出现的顺序一致
// - 每个label只会使用一次，如果流式文本中需要2次，请使用不同的标签名，他们可以有相同的startToken和endToken，但是必须保证labels有序
// - 只有当labels的startToken之间没有重叠，labels才不用关注顺序
// - startToken 不会输出，但是endToken可能会，这样做的目的是可以保证下一个label的startToken可以从当前label的endToken之前进行匹配
func (s *labelStream) Recv() (LabeledChunk, error) {
	for {
		label, chunk := s.cutOverflowBuffer() // 先处理buffer中的数据
		if chunk != "" {
			return LabeledChunk{Label: label, Chunk: chunk}, nil
		}

		upChunk, err := s.src.Recv()
		if err == io.EOF { // 上游已经读完了，但是buffer中还有数据，需要处理
			s.minBufferLen = 0
			label, chunk := s.cutOverflowBuffer()
			if chunk != "" {
				return LabeledChunk{Label: label, Chunk: chunk}, nil
			} else {
				return LabeledChunk{}, io.EOF
			}
		} else if err != nil {
			return LabeledChunk{}, err
		}
		s.buffer += upChunk
	}
}

// checkLabels 检查label是否合法，比如label未命名（会和默认label冲突），或者不同label之间startToken有重叠，会导致label识别结果不确定
func checkLabels(labels []SLabel) error {
	if len(labels) == 0 {
		return errors.New("labels is empty")
	}
	return nil
}

// Split 将流式文本按照label切分成多个流
func (s *labelStream) Demux() map[string]Stream[string] {
	labels := make([]string, len(s.labels))
	for _, lb := range s.labels {
		labels = append(labels, lb.Name)
	}

	tmp := Demux(s, func(s LabeledChunk) string { return s.Label }, labels)
	res := make(map[string]Stream[string], len(labels))
	for k, v := range tmp {
		res[k] = Map(v, func(s LabeledChunk) string { return s.Chunk })
	}
	return res
}
