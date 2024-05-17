package main

import (
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"metrics/server/pb/pv"
)

func (s *server) Export(ctx context.Context,
	req *pb.ExportMetricsServiceRequest) (*pb.ExportMetricsServiceResponse, error) {
	log.Printf("Received request: %v", req)

	hasErrors := false
	var errorMessage string
	rejectedDataPoints := 0

	// Helper function to check if a metric is empty
	isEmptyMetric := func(metric *v1.Metric) bool {
		return metric.Name == "" || metric.Description == "" || metric.Unit == "" || metric.Data == nil
	}

	for _, resourceMetrics := range req.ResourceMetrics {
		for _, scopeMetrics := range resourceMetrics.ScopeMetrics {
			for _, metric := range scopeMetrics.Metrics {
				if isEmptyMetric(metric) { // Example error condition
					hasErrors = true
					rejectedDataPoints++
					errorMessage = "Found nil metric"
				}
			}
		}
	}

	response := &pb.ExportMetricsServiceResponse{}
	if hasErrors {
		log.Printf("hasErrors: %v", hasErrors)
		// Build response with errors
		response = &pb.ExportMetricsServiceResponse{
			PartialSuccess: &pb.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(rejectedDataPoints),
				ErrorMessage:       errorMessage,
			},
		}
	}

	return response, nil
}

func (s *server) GetVersion(context.Context, *emptypb.Empty) (*pv.VersionResponse, error) {
	log.Printf("Received request:")
	return &pv.VersionResponse{
		BuildTime: int32(buildTime),
		GitCommit: gitCommit,
	}, nil
}
