package main

import (
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestCircularQueue_Enqueue(t *testing.T) {
	queue := NewCircularQueue(5)

	now := time.Now()
	for i := 0; i < 5; i++ {
		queue.Enqueue(CachedRequest{
			Request:   nil,
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	// Check if the internal queue has 5 elements with the correct timestamps
	for i := 0; i < 5; i++ {
		req := (*CachedRequest)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&queue.queue[i]))))
		if req.Timestamp != now.Add(time.Duration(i)*time.Second) {
			t.Errorf("Expected Timestamp %v at index %d, got %v", now.Add(time.Duration(i)*time.Second), i, req.Timestamp)
		}
	}
}

func TestCircularQueue_EnqueueOverflow(t *testing.T) {
	queue := NewCircularQueue(3)

	now := time.Now()
	for i := 0; i < 4; i++ {
		queue.Enqueue(CachedRequest{
			Request:   nil,
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	expectedTimes := []time.Time{
		now.Add(3 * time.Second),
		now.Add(1 * time.Second),
		now.Add(2 * time.Second),
	}
	// Assert that the queue contains the expected timestamps in the correct order after overflow
	for i := 0; i < len(expectedTimes); i++ {
		req := (*CachedRequest)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&queue.queue[i]))))
		if !req.Timestamp.Equal(expectedTimes[i]) {
			t.Errorf("Expected timestamp %s at index %d, got %s", expectedTimes[i], i, req.Timestamp)
		}
	}
}
