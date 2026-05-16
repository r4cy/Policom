package core

import "context"

type contextKey string

const traceIDKey contextKey = "trace_id"

func TraceIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(traceIDKey).(string); ok {
		return val
	}
	return ""
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey, traceID)
}
