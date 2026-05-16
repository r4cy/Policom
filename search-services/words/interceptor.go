package main

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		trace_id := ""
		md, ok := metadata.FromIncomingContext(ctx); 
		if ok {
			values := md.Get("trace_id")
			if len(values) > 0 {
				trace_id = values[0]
			}
		}
		
		logger := log.With("trace_id", trace_id)
		logger.InfoContext(ctx, "grpc request", "method", info.FullMethod)
		return handler(ctx, req)
	}
}