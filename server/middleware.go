package main

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
