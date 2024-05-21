package main

import (
	"testing"
	"time"
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

	if len(queue.queue) != 5 {
		t.Errorf("Expected queue length 5, got %d", len(queue.queue))
	}

	for i := 0; i < 5; i++ {
		if queue.queue[i].Timestamp != now.Add(time.Duration(i)*time.Second) {
			t.Errorf("Expected Timestamp %v at index %d, got %v", now.Add(time.Duration(i)*time.Second), i, queue.queue[i].Timestamp)
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
	// Assert that the queue contains the expected timestamps
	for i := 0; i < len(expectedTimes); i++ {
		if !queue.queue[i].Timestamp.Equal(expectedTimes[i]) {
			t.Errorf("Expected timestamp %s at index %d, got %s", expectedTimes[i], i, queue.queue[i].Timestamp)
		}
	}
}
