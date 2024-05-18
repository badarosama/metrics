// circular_queue.go
package main

import (
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"sync"
	"time"
)

type CachedRequest struct {
	Request   *pb.ExportMetricsServiceRequest
	Timestamp time.Time
}

type CircularQueue struct {
	queue      []CachedRequest
	size       int
	head, tail int
	mutex      sync.Mutex
}

func NewCircularQueue(size int) *CircularQueue {
	return &CircularQueue{
		queue: make([]CachedRequest, size),
		size:  size,
	}
}

func (q *CircularQueue) Enqueue(request CachedRequest) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if (q.tail+1)%q.size == q.head {
		// Queue is full, dequeue one element
		q.head = (q.head + 1) % q.size
	}

	q.queue[q.tail] = request
	q.tail = (q.tail + 1) % q.size
}

func (q *CircularQueue) Dequeue() *CachedRequest {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.head == q.tail {
		// Queue is empty
		return nil
	}

	request := q.queue[q.head]
	q.head = (q.head + 1) % q.size
	return &request
}
