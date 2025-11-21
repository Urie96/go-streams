package streams

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type streamWithLog[T any] struct {
	Stream[T]
	key string
	log func(info string)
	mu  sync.Mutex

	collect []T
	startAt *time.Time
	stop    bool
}

func WithLog[T any](stream Stream[T], key string, log func(info string)) Stream[T] {
	return &streamWithLog[T]{
		Stream: avoidNil(stream),
		key:    key,
		log:    log,
	}
}

func (s *streamWithLog[T]) Recv() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.startAt == nil {
		now := time.Now()
		s.startAt = &now
		s.log(fmt.Sprintf("[StreamLog] stream %s start to consume", s.key))
	}

	c, err := s.Stream.Recv()

	if s.stop {
		return c, err
	}

	if err == io.EOF {
		cost := time.Now().Sub(*s.startAt)
		s.log(fmt.Sprintf("[StreamLog] stream %s consume completed, cost %s, %d items: %s", s.key, cost.String(), len(s.collect), s.allToString()))
		s.stop = true
	} else if err != nil {
		s.log(fmt.Sprintf("[StreamLog] stream %s consume error: %v", s.key, err))
		s.stop = true
	} else {
		s.collect = append(s.collect, c)
	}
	return c, err
}

func (s *streamWithLog[T]) allToString() string {
	switch v := any(s.collect).(type) {
	case []string:
		return strings.Join(v, " ")
	default:
		b, err := json.Marshal(s.collect)
		if err != nil {
			return err.Error()
		}
		return string(b)
	}
}
