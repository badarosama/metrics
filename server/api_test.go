package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestExport(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	s := &server{
		logger:                 logger,
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
			assert.NoError(t, err)

			if tt.wantErrors {
				assert.NotNil(t, resp.PartialSuccess)
				assert.Equal(t, tt.wantRejected, resp.PartialSuccess.RejectedDataPoints)
				assert.Equal(t, tt.wantErrorMessage, resp.PartialSuccess.ErrorMessage)
			} else {
				assert.Nil(t, resp.PartialSuccess)
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
