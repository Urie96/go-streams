package streams

import "sync"

type forkStream[T any] struct {
	cache *streamCache[T]
	index int
}

type streamCache[T any] struct {
	src Stream[T]
	buf []T
	err error
	mu  sync.RWMutex
}

func newCache[T any](src Stream[T]) *streamCache[T] {
	return &streamCache[T]{
		src: avoidNil(src),
	}
}

func (c *streamCache[T]) tryGet(i int) (bool, T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 已有足够数据或已结束
	if i < len(c.buf) {
		return true, c.buf[i], nil
	}
	var zero T
	if c.err != nil {
		return true, zero, c.err
	}
	return false, zero, nil
}

func (c *streamCache[T]) blockLoad(i int) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果等待锁的期间，其他协程已经拉到数据了
	if i < len(c.buf) {
		return c.buf[i], nil
	}
	if c.err != nil {
		var zero T
		return zero, c.err
	}

	// 到这里说明当前协程是第一个拿到锁的，那我们读一条并缓存
	val, err := c.src.Recv()
	if err == nil {
		c.buf = append(c.buf, val)
	} else {
		c.err = err
	}
	return val, err
}

func (c *streamCache[T]) Get(i int) (T, error) {
	if ok, val, err := c.tryGet(i); ok {
		return val, err
	}
	return c.blockLoad(i)
}

func (t *forkStream[T]) Recv() (T, error) {
	index := t.index
	t.index++
	return t.cache.Get(index)
}

// TeeReader 将一条流复制为两条流，需要注意入参的流不可再消费
func TeeReader[T any](src Stream[T]) (Stream[T], Stream[T]) {
	ss := Fork(src, 2)
	return ss[0], ss[1]
}

// Fork 将一条流复制为多条流，需要注意入参的流不可再消费
func Fork[T any](src Stream[T], copies int) []Stream[T] {
	src = avoidNil(src)
	res := make([]Stream[T], copies)
	if unwrapped, ok := src.(*forkStream[T]); ok && unwrapped != nil { // 性能优化：如果入参是forkStream，则直接复用缓存
		cache := unwrapped.cache
		for i := range res {
			res[i] = &forkStream[T]{cache: cache, index: unwrapped.index}
		}
	} else {
		cache := newCache(src)
		for i := range res {
			res[i] = &forkStream[T]{cache: cache, index: 0}
		}
	}

	return res
}

func Demux[T any](src Stream[T], classifier func(T) string, labelsRange []string) map[string]Stream[T] {
	demuxRes := make(map[string]Stream[T], len(labelsRange)+1)
	copyStreams := Fork(src, len(labelsRange)+1)
	for i, label := range labelsRange {
		demuxRes[label] = Filter(copyStreams[i], func(t T) bool {
			return classifier(t) == label
		})
	}
	demuxRes[""] = Filter(copyStreams[len(labelsRange)], func(t T) bool {
		return classifier(t) == ""
	})
	return demuxRes
}
