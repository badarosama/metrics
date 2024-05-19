package main

import (
	"context"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestExport(t *testing.T) {
	s := &server{
		logger:                 &zap.Logger{},
		lastErrorRequests:      NewCircularQueue(10),
		lastSuccessfulRequests: NewCircularQueue(10),
	}
	ctx := context.Background()

	tests := []struct {
		name             string
		request          *pb.ExportMetricsServiceRequest
		wantErrors       bool
		wantRejected     int64
		wantErrorMessage string
	}{
		{
			name: "Valid Request",
			request: &pb.ExportMetricsServiceRequest{
				ResourceMetrics: []*v1.ResourceMetrics{
					{
						ScopeMetrics: []*v1.ScopeMetrics{
							{
								Metrics: []*v1.Metric{
									{Name: "metric1", Description: "desc1", Unit: "unit1", Data: &v1.Metric_Sum{}},
								},
							},
						},
					},
				},
			},
			wantErrors: false,
		},
		{
			name: "Empty Metric",
			request: &pb.ExportMetricsServiceRequest{
				ResourceMetrics: []*v1.ResourceMetrics{
					{
						ScopeMetrics: []*v1.ScopeMetrics{
							{
								Metrics: []*v1.Metric{
									{Name: "", Description: "", Unit: "", Data: nil},
								},
							},
						},
					},
				},
			},
			wantErrors:       true,
			wantRejected:     1,
			wantErrorMessage: "Found nil metric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Export(ctx, tt.request)
			if err != nil {
				t.Fatalf("Export() error = %v", err)
			}

			if (resp.PartialSuccess != nil) != tt.wantErrors {
				t.Errorf("Export() wantErrors = %v, got %v", tt.wantErrors, resp.PartialSuccess != nil)
			}

			if tt.wantErrors {
				if resp.PartialSuccess.RejectedDataPoints != tt.wantRejected {
					t.Errorf("Export() rejectedDataPoints = %v, want %v", resp.PartialSuccess.RejectedDataPoints, tt.wantRejected)
				}

				if resp.PartialSuccess.ErrorMessage != tt.wantErrorMessage {
					t.Errorf("Export() errorMessage = %v, want %v", resp.PartialSuccess.ErrorMessage, tt.wantErrorMessage)
				}
			}
		})
	}
}

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
	queue := NewCircularQueue(5)

	now := time.Now()
	for i := 0; i < 7; i++ {
		queue.Enqueue(CachedRequest{
			Request:   nil,
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	expectedTimes := []time.Time{
		now.Add(2 * time.Second),
		now.Add(3 * time.Second),
		now.Add(4 * time.Second),
		now.Add(5 * time.Second),
		now.Add(6 * time.Second),
	}

	for i, expected := range expectedTimes {
		if queue.queue[(queue.head+i)%queue.size].Timestamp != expected {
			t.Errorf("Expected Timestamp %v at index %d, got %v", expected, i, queue.queue[(queue.head+i)%queue.size].Timestamp)
		}
	}
}

func TestCircularQueue_PrintFirst(t *testing.T) {
	queue := NewCircularQueue(5)

	now := time.Now()
	queue.Enqueue(CachedRequest{
		Request:   nil,
		Timestamp: now,
	})

	// Simply call the PrintFirst method and ensure it doesn't panic.
	queue.PrintFirst()
}

func TestCircularQueue_PrintLast(t *testing.T) {
	queue := NewCircularQueue(5)

	now := time.Now()
	queue.Enqueue(CachedRequest{
		Request:   nil,
		Timestamp: now,
	})

	// Simply call the PrintLast method and ensure it doesn't panic.
	queue.PrintLast()
}

func TestCircularQueue_PrintAll(t *testing.T) {
	queue := NewCircularQueue(5)

	now := time.Now()
	for i := 0; i < 5; i++ {
		queue.Enqueue(CachedRequest{
			Request:   nil,
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	// Simply call the PrintAll method and ensure it doesn't panic.
	queue.PrintAll()
}
