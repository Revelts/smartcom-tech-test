package repository

import (
	"context"
	"sync"

	"github.com/smartcom/integration-platform/services/middleware/internal/domain"
)

type EventQueue struct {
	queue  chan domain.Event
	mu     sync.RWMutex
	closed bool
}

const DefaultQueueSize = 1000

func NewEventQueue(size int) (q *EventQueue) {
	if size <= 0 {
		size = DefaultQueueSize
	}

	q = &EventQueue{
		queue:  make(chan domain.Event, size),
		closed: false,
	}
	return
}

func (q *EventQueue) Enqueue(ctx context.Context, event domain.Event) (err error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		err = ErrQueueClosed
		return
	}

	select {
	case q.queue <- event:
	case <-ctx.Done():
		err = ctx.Err()
		return
	}

	return
}

func (q *EventQueue) Dequeue(ctx context.Context) (event domain.Event, ok bool) {
	select {
	case event, ok = <-q.queue:
		return
	case <-ctx.Done():
		ok = false
		return
	}
}

func (q *EventQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.closed {
		q.closed = true
		close(q.queue)
	}
}

func (q *EventQueue) Len() (length int) {
	length = len(q.queue)
	return
}
