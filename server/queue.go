package main

import (
	"fmt"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"sync"
	"time"
)

type CachedRequest struct {
	Request   *pb.ExportMetricsServiceRequest
	Timestamp time.Time
}

type CircularQueue struct {
	queue []CachedRequest
	size  int
	head  int
	tail  int
	mutex sync.Mutex
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

// PrintFirst prints the first element in the queue
func (q *CircularQueue) PrintFirst() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.head == q.tail {
		fmt.Println("Queue is empty")
		return
	}
	fmt.Printf("First element: %+v\n", q.queue[q.head])
}

// PrintLast prints the last element in the queue
func (q *CircularQueue) PrintLast() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.head == q.tail {
		fmt.Println("Queue is empty")
		return
	}
	lastIndex := (q.tail - 1 + q.size) % q.size
	fmt.Printf("Last element: %+v\n", q.queue[lastIndex])
}

// PrintAll prints all elements in the queue in order
func (q *CircularQueue) PrintAll() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.head == q.tail {
		fmt.Println("Queue is empty")
		return
	}
	fmt.Println("All elements in order:")
	i := q.head
	for {
		fmt.Printf("%+v\n", q.queue[i])
		if i == q.tail-1 {
			break
		}
		i = (i + 1) % q.size
	}
}
