package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"yadro.com/course/api/core"
)

func TestMiddlewareRole(t *testing.T) {
	t.Run("without user in context", TestMiddlewareRoleWithoutUser)
	t.Run("wrong role", TestMiddlewareRoleWrongRole)
	t.Run("correct role", TestMiddlewareRoleCorrectRole)
}

func TestMiddlewareRoleWithoutUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/db/stats", nil)
	rec := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireRole(core.RoleAdmin)(next)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.False(t, called)
}

func TestMiddlewareRoleWrongRole(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/db/stats", nil)

	user := core.User{
		ID:   "user-1",
		Role: core.RoleUser,
	}

	ctx := context.WithValue(req.Context(), core.UserKey, user)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireRole(core.RoleAdmin)(next)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.False(t, called)
}

func TestMiddlewareRoleCorrectRole(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/db/stats", nil)

	user := core.User{
		ID:   "admin-1",
		Role: core.RoleAdmin,
	}

	ctx := context.WithValue(req.Context(), core.UserKey, user)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireRole(core.RoleAdmin)(next)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, called)
}
