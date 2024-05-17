package main

import (
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"metrics/server/pb/pv"
	"metrics/server/version"
	"time"
)

func (s *server) Export(ctx context.Context,
	req *pb.ExportMetricsServiceRequest) (*pb.ExportMetricsServiceResponse, error) {
	logger.Info("Received request: %v", zap.Any("request", req))
	//log.Printf("Received request: %v", req)

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

		// Add request to lastErrorRequests cache
		s.addToErrorCache(req)
	} else {
		// Add request to lastSuccessfulRequests cache
		s.addToSuccessCache(req)
	}

	return response, nil
}

// addToSuccessCache adds a successful request to the lastSuccessfulRequests cache.
func (s *server) addToSuccessCache(req *pb.ExportMetricsServiceRequest) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if len(s.lastSuccessfulRequests) >= 10 {
		// Remove oldest request
		s.lastSuccessfulRequests = s.lastSuccessfulRequests[1:]
	}
	// Append new request with timestamp
	s.lastSuccessfulRequests = append(s.lastSuccessfulRequests, &cachedRequest{
		Request:   req,
		Timestamp: time.Now(),
	})
}

// addToErrorCache adds an error request to the lastErrorRequests cache.
func (s *server) addToErrorCache(req *pb.ExportMetricsServiceRequest) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if len(s.lastErrorRequests) >= 10 {
		// Remove oldest request
		s.lastErrorRequests = s.lastErrorRequests[1:]
	}
	// Append new request with timestamp
	s.lastErrorRequests = append(s.lastErrorRequests, &cachedRequest{
		Request:   req,
		Timestamp: time.Now(),
	})
}

func (s *server) GetVersion(context.Context, *emptypb.Empty) (*pv.VersionResponse, error) {
	log.Printf("Received request:")
	commitSha, timestamp := version.BuildVersion()
	return &pv.VersionResponse{
		BuildTimestamp: timestamp,
		GitCommitSha:   commitSha,
	}, nil
}
