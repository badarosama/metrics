package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

func TestCircularQueue(t *testing.T) {
	size := 3
	queue := NewCircularQueue(size)

	// Test Enqueue
	request1 := CachedRequest{Request: &pb.ExportMetricsServiceRequest{}, Timestamp: time.Now()}
	queue.Enqueue(request1)
	assert.Equal(t, 1, queue.tail, "Tail should be updated after enqueueing an element")
	assert.Equal(t, request1, queue.queue[0], "Enqueued element should be at the head of the queue")

	// Test Enqueue when queue is full
	request2 := CachedRequest{Request: &pb.ExportMetricsServiceRequest{}, Timestamp: time.Now()}
	request3 := CachedRequest{Request: &pb.ExportMetricsServiceRequest{}, Timestamp: time.Now()}
	queue.Enqueue(request2)
	queue.Enqueue(request3)
	assert.Equal(t, 0, queue.head, "Head should be updated after dequeueing an element due to full queue")
	assert.Nil(t, queue.queue[0].Request, "Oldest element should be nil after enqueueing new elements when queue is full")

	// Test PrintFirst and PrintLast
	queue.PrintFirst()
	queue.PrintLast()

	// Test PrintAll
	queue.PrintAll()
}
