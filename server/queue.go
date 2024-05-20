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

// CircularQueue is a thread-safe circular queue that holds CachedRequest elements
type CircularQueue struct {
	queue []CachedRequest // The queue slice holding the CachedRequests
	size  int             // The size of the queue
	head  int             // The index of the head of the queue
	tail  int             // The index of the tail of the queue
	mutex sync.Mutex      // A mutex to ensure thread-safety
}

func NewCircularQueue(size int) *CircularQueue {
	return &CircularQueue{
		queue: make([]CachedRequest, size),
		size:  size,
	}
}

// Enqueue adds a new CachedRequest to the queue
// If the queue is full, it will overwrite the oldest element
func (q *CircularQueue) Enqueue(request CachedRequest) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// Adjust head if the queue is full
	if (q.tail+1)%q.size == q.head {
		// Queue is full, dequeue one element
		q.head = (q.head + 1) % q.size
	}

	q.queue[q.tail] = request
	q.tail = (q.tail + 1) % q.size
}
