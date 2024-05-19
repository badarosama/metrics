package main

import (
	"testing"
	"time"
)

func TestCircularQueueOverwrite(t *testing.T) {
	// Create a new circular queue with a fixed size of 10
	queue := NewCircularQueue(10)

	// Enqueue 15 elements with timestamps
	for i := 0; i < 15; i++ {
		request := CachedRequest{
			Request:   nil, // Dummy request
			Timestamp: time.Now(),
		}
		queue.Enqueue(request)
	}

	// Check the contents of the queue
	expectedLength := 10
	if len(queue.queue) != expectedLength {
		t.Errorf("Expected queue length %d, got %d", expectedLength, len(queue.queue))
	}

	// Check that the timestamps in the queue are in the expected order (oldest overwritten)
	oldestTimestamp := queue.queue[0].Timestamp
	for i := 1; i < len(queue.queue); i++ {
		if queue.queue[i].Timestamp.Before(oldestTimestamp) {
			oldestTimestamp = queue.queue[i].Timestamp
		}
	}

	// Now compare the timestamps with the oldestTimestamp
	for i := 1; i < len(queue.queue); i++ {
		if queue.queue[i].Timestamp.Before(oldestTimestamp) {
			t.Errorf("Elements in the queue are not overwritten properly")
		}
	}
}
