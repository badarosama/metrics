package main

import (
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"golang.org/x/net/context"
	"testing"
)

func TestExport(t *testing.T) {
	s := &server{}
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
