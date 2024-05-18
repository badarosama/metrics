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

type Node struct {
	value CachedRequest
	next  *Node
	prev  *Node
}

type LinkedList struct {
	head  *Node
	tail  *Node
	size  int
	cap   int
	mutex sync.Mutex
}

func NewLinkedList(cap int) *LinkedList {
	return &LinkedList{
		cap: cap,
	}
}

func (l *LinkedList) Append(value CachedRequest) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	newNode := &Node{value: value}

	if l.tail != nil {
		l.tail.next = newNode
		newNode.prev = l.tail
		l.tail = newNode
	} else {
		l.head = newNode
		l.tail = newNode
	}

	if l.size == l.cap {
		l.head = l.head.next
		if l.head != nil {
			l.head.prev = nil
		}
	} else {
		l.size++
	}
}

func (l *LinkedList) GetAll() []CachedRequest {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var requests []CachedRequest
	current := l.head
	for current != nil {
		requests = append(requests, current.value)
		current = current.next
	}
	return requests
}
