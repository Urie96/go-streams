package streams

import "sync"

type Node[T any] struct {
	value T
	next  *Node[T]
}

type ConcurrentQueue[T any] struct {
	head *Node[T]
	tail *Node[T]
	mu   sync.RWMutex
}

// Enqueue 入队操作
func (q *ConcurrentQueue[T]) Push(value T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	newNode := &Node[T]{value: value, next: nil}
	if q.tail == nil {
		q.head = newNode
		q.tail = newNode
	} else {
		q.tail.next = newNode
		q.tail = newNode
	}
}

// Dequeue 出队操作
func (q *ConcurrentQueue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T

	if q.head == nil {
		return zero, false
	}

	value := q.head.value
	q.head = q.head.next

	if q.head == nil {
		q.tail = nil
	}

	return value, true
}
