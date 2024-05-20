package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"time"
)

// UnaryInterceptorPrometheus is a gRPC unary interceptor that collects metrics
// related to incoming unary RPC requests and records them using Prometheus.
//
// Parameters:
//   - ctx: The context.Context object representing the context of the RPC call.
//   - req: The request message sent by the client.
//   - info: Information about the RPC call, including method name and other metadata.
//   - handler: The gRPC unary handler function that processes the request.
//
// Returns:
//   - interface{}: The response message returned by the handler function.
//   - error: An error encountered during request processing, if any.
//
// This interceptor retrieves peer information from the context, including the client's address.
// It measures the duration of the RPC call processing and records metrics using Prometheus.
// The recorded metrics include request count and duration, labeled with method name,
// client address, and status code.
//
// Example usage:
//
//	// Register the interceptor with your gRPC server.
//	s := grpc.NewServer(
//	    grpc.UnaryInterceptor(UnaryInterceptorPrometheus),
//	)
//
//	// Start your gRPC server.
//	if err := s.Serve(lis); err != nil {
//	    log.Fatalf("failed to serve: %v", err)
//	}
//
//	// Ensure that your Prometheus metrics are registered for scraping.
//	prometheus.MustRegister(requestCount)
//	prometheus.MustRegister(requestDuration)
//
// This function is typically used as a gRPC server interceptor to monitor and measure
// the performance of unary RPC calls in your gRPC server.
func UnaryInterceptorPrometheus(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	p, _ := peer.FromContext(ctx)
	clientAddr := "unknown"
	if p != nil {
		clientAddr = p.Addr.String()
	}

	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start).Seconds()

	code := status.Code(err).String()

	// https://prometheus.io/docs/prometheus/latest/getting_started/
	// Record the metrics
	requestCount.WithLabelValues(info.FullMethod, clientAddr, code).Inc()
	requestDuration.WithLabelValues(info.FullMethod).Observe(duration)

	return resp, err
}
