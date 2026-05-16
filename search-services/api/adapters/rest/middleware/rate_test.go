package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareRate(t *testing.T) {
	t.Run("without problem", TestMiddlewareRateGood)
	t.Run("rate return bad status", TestMiddlewareBad)
}

func TestMiddlewareRateGood(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/isearch?phrase='hello'", nil)
	rec := httptest.NewRecorder()

	handler := Rate(stubHandler, 100)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "should without problem")
}

func TestMiddlewareBad(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/isearch?phrase='hello'", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := Rate(stubHandler, 100)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code, "should return with problem")
}
