package main

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"time"
)

// UnaryInterceptor is a gRPC unary server interceptor that tracks request latencies.
func UnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)

		// Calculate request latency
		latency := time.Since(startTime)
		// Log the latency
		logger.Info("Request processed",
			zap.String("method", info.FullMethod),
			zap.Duration("latency", latency),
		)

		return resp, err
	}
}

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

	// Record the metrics
	requestCount.WithLabelValues(info.FullMethod, clientAddr, code).Inc()
	requestDuration.WithLabelValues(info.FullMethod).Observe(duration)

	return resp, err
}
