package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTraceIDFromContextEmpty(t *testing.T) {
	traceID := TraceIDFromContext(context.Background())

	require.Empty(t, traceID)
}

func TestContextWithTraceID(t *testing.T) {
	ctx := ContextWithTraceID(context.Background(), "trace-123")

	traceID := TraceIDFromContext(ctx)

	require.Equal(t, "trace-123", traceID)
}

func TestTraceIDFromContextWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), traceIDKey, 123)

	traceID := TraceIDFromContext(ctx)

	require.Empty(t, traceID)
}
