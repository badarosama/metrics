package main

import (
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/emptypb"
	"metrics/server/pb/pv"
	"metrics/server/version"
	"time"
)

// Export is a gRPC method of the MetricsService service that handles the exporting of metrics data.
//
// This method receives an ExportMetricsServiceRequest containing metrics data to be exported.
// It processes the received metrics data, checks for any errors, and generates an appropriate response.
// If there are errors in the received data, it returns a response with details about the errors.
// Otherwise, it returns a response indicating successful processing.
//
// Parameters:
// - ctx: The context.Context for the RPC call.
// - req: The ExportMetricsServiceRequest containing the metrics data to be exported.
//
// Returns:
// - *pb.ExportMetricsServiceResponse: The response containing the result of the export operation.
// - error: An error, if any occurred during processing.
func (s *server) Export(ctx context.Context,
	req *pb.ExportMetricsServiceRequest) (*pb.ExportMetricsServiceResponse, error) {
	//s.logger.Info("Received request", zap.Any("request", req))

	hasErrors := false
	var errorMessage string
	rejectedDataPoints := 0

	// Helper function to check if a metric is empty:
	// Other checks can also be added for error checking
	// based on requirements.
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
		// Build response with errors
		response = &pb.ExportMetricsServiceResponse{
			PartialSuccess: &pb.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(rejectedDataPoints),
				ErrorMessage:       errorMessage,
			},
		}

		s.lastErrorRequests.Enqueue(CachedRequest{
			Request:   req,
			Timestamp: time.Now(),
		})
	} else {
		s.lastSuccessfulRequests.Enqueue(CachedRequest{
			Request:   req,
			Timestamp: time.Now(),
		})
	}

	return response, nil
}

// GetVersion retrieves the current version information. This method takes no parameters and returns a VersionResponse
// message containing version information such as the build timestamp and Git commit SHA.
func (s *server) GetVersion(context.Context, *emptypb.Empty) (*pv.VersionResponse, error) {
	commitSha, timestamp := version.BuildVersion()
	s.logger.Info("Received request GetVersion", zap.String("commitSha", commitSha))

	return &pv.VersionResponse{
		BuildTimestamp: timestamp,
		GitCommitSha:   commitSha,
	}, nil
}
