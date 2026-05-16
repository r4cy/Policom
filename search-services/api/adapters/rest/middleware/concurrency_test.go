package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)


func TestMiddlewareConc(t *testing.T) {
	t.Run("without problem", TestMiddlewareConcGood)
	t.Run("semaphor return bad status", TestMiddlewareConcBad)
}

func TestMiddlewareConcGood(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	rec := httptest.NewRecorder()

	handler := Concurrency(stubHandler, 1)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "should without problem")
}

func TestMiddlewareConcBad(t *testing.T) {
	var block = make(chan struct{})

	var slowHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()

	handler := Concurrency(slowHandler, 1)

	go func ()  {
		handler.ServeHTTP(rec1, req1)
	}()
	time.Sleep(10 * time.Millisecond)

	handler.ServeHTTP(rec2, req2)
	
	require.Equal(t, http.StatusServiceUnavailable, rec2.Code, "should with problem")
	block <- struct{}{}
	defer close(block)
}
