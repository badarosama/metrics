package main

import (
	"sync/atomic"
	"time"
	"unsafe"

	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

type CachedRequest struct {
	Request   *pb.ExportMetricsServiceRequest
	Timestamp time.Time
}

// CircularQueue is a thread-safe circular queue that holds CachedRequest elements
type CircularQueue struct {
	queue []CachedRequest // The queue slice holding the CachedRequests
	size  int             // The size of the queue
	head  int32           // The index of the head of the queue
	tail  int32           // The index of the tail of the queue
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
	tail := atomic.LoadInt32(&q.tail)
	head := atomic.LoadInt32(&q.head)

	// Adjust head if the queue is full
	if (tail+1)%int32(q.size) == head {
		// Queue is full, dequeue one element
		atomic.StoreInt32(&q.head, (head+1)%int32(q.size))
	}

	// Use unsafe.Pointer to atomically store the request
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&q.queue[tail])), unsafe.Pointer(&request))
	atomic.StoreInt32(&q.tail, (tail+1)%int32(q.size))
}
