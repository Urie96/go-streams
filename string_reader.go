package streams

import (
	"io"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

type StringReader struct {
	Stream[string]
	buffer strings.Builder
}

func NewStringReader(stream Stream[string]) *StringReader {
	return &StringReader{Stream: avoidNil(stream)}
}

// ReadUntil 合并流中的字符串直到遇到 delim 中的任意一个
// 注意事项：
// 1. 如果delims长度为0，则会阻塞合并流中的所有字符串
// 2. 如果有delim是""，则会将流切分为rune粒度
// 3. 如果上游流中一个chunk包含多个delim，不会按最先出现的delim进行切分，而是按delims的顺序进行切分(可优化)
func (b *StringReader) ReadUntil(delims []string) (string, error) {
	for {
		if b.buffer.Len() > 0 {
			buffer := b.buffer.String()
			for _, delim := range delims {
				index := strings.Index(buffer, delim)
				if index >= 0 {
					b.buffer.Reset()
					endIndex := index + len(delim)
					if delim == "" {
						_, size := utf8.DecodeRuneInString(buffer)
						endIndex = size
					}
					b.buffer.WriteString(buffer[endIndex:])
					return buffer[:endIndex], nil
				}
			}
		}

		v, err := b.Stream.Recv()
		if err == io.EOF {
			if b.buffer.Len() > 0 {
				buffer := b.buffer.String()
				b.buffer.Reset()
				return buffer, nil
			} else {
				return "", io.EOF
			}
		} else if err != nil {
			return "", err
		}
		if len(v) > 0 {
			b.buffer.WriteString(v)
		}
	}
}

// Recv 先将buffer返回，再读取上游流中的数据。主要用途是在ReadUntil方法消费到某个token之后，形成一个新的Stream
// 注意：由于清空了buffer，所以不可和ReadUntil方法并发调用(这样做也没意义)
func (b *StringReader) Recv() (string, error) {
	if b.buffer.Len() > 0 {
		buffer := b.buffer.String()
		b.buffer.Reset()
		return buffer, nil
	}
	return b.Stream.Recv()
}

// ReadLine 合并流中的字符串直到遇到换行符
func (b *StringReader) ReadLine() (string, error) {
	return b.ReadUntil([]string{"\n"})
}

// ToSplitReader 将ReadUtil方法封装为Stream
func (b *StringReader) ToSplitReader(delims []string) Stream[string] {
	return delimsStringReader{src: b, delims: delims}
}

// ToLineReader 将ReadLine方法封装为Stream
func (b *StringReader) ToLineReader() Stream[string] {
	return b.ToSplitReader([]string{"\n"})
}

func (b *StringReader) Collect() (string, error) {
	return OnceStringStream(b).Recv()
}

// CollectString 阻塞收集流中的数据，并合并成一个字符串返回
func CollectString(stream Stream[string]) (string, error) {
	return NewStringReader(stream).Collect()
}

type delimsStringReader struct {
	src    *StringReader
	delims []string
}

func (d delimsStringReader) Recv() (string, error) {
	return d.src.ReadUntil(d.delims)
}

type onceStringStream struct {
	Stream[string]
	done atomic.Bool
}

func (s *onceStringStream) Recv() (string, error) {
	if s.done.Swap(true) {
		return "", io.EOF
	}
	var sb strings.Builder
	err := Consume(s.Stream, func(v string) error {
		sb.WriteString(v)
		return nil
	})
	return sb.String(), err
}

// OnceStringStream 等到流都结束了，才将拼接好的字符串发送出来，在消费方想将某个流式突然想变成非流式时有用
func OnceStringStream(stream Stream[string]) Stream[string] {
	return &onceStringStream{Stream: avoidNil(stream)}
}
