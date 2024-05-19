package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"time"
)

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
