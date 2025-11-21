package streams

import (
	"sync"
	"time"
)

type streamWithTracer[T any] struct {
	Stream[T]
	span Span
	mu   sync.Mutex

	collect      []T
	startAt      *time.Time
	firstTokenAt *time.Time
	stop         bool
}

type Span interface {
	SetOutput(output any)
	Finish()
}

func WithTracer[T any](stream Stream[T], span Span) Stream[T] {
	return &streamWithTracer[T]{
		Stream: avoidNil(stream),
		span:   span,
	}
}

func (s *streamWithTracer[T]) finish(err error) {
	stopAt := time.Now()
	cost := stopAt.Sub(*s.startAt)
	s.span.SetOutput(map[string]any{
		"error":          err,
		"start_at":       s.startAt.Format(time.DateTime),
		"first_token_at": s.firstTokenAt.Format(time.DateTime),
		"end_at":         stopAt.Format(time.DateTime),
		"cost":           cost.String(),
		"frames":         s.collect,
		"frames_count":   len(s.collect),
	})
	s.span.Finish()
}

func (s *streamWithTracer[T]) Recv() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.startAt == nil {
		now := time.Now()
		s.startAt = &now
	}

	c, err := s.Stream.Recv()

	if s.firstTokenAt == nil {
		now := time.Now()
		s.firstTokenAt = &now
	}

	if s.stop {
		return c, err
	}

	if err != nil {
		s.stop = true
		s.finish(err)
	} else {
		s.collect = append(s.collect, c)
	}
	return c, err
}
