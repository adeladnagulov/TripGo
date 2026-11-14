package currentlyinqueue

import (
	"sync"

	"github.com/google/uuid"
)

type InQueue struct {
	mu    sync.RWMutex
	items map[uuid.UUID]struct{}
}

func NewInQueue() *InQueue {
	return &InQueue{
		items: make(map[uuid.UUID]struct{}),
	}
}

func (q *InQueue) TryAdd(tripId uuid.UUID) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.items[tripId]; ok {
		return false
	}
	q.items[tripId] = struct{}{}
	return true
}

func (q *InQueue) Remove(tripId uuid.UUID) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.items, tripId)
}

func (q *InQueue) Contains(tripId uuid.UUID) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	_, ok := q.items[tripId]
	return ok
}

func (q *InQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}
