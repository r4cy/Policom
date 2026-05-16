package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"yadro.com/course/api/core"
)

var stubHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

type fakeVerifier struct {
    user     core.User
	err      error
	called   bool
	gotToken string
}

func (f *fakeVerifier) Verify(ctx context.Context, token string) (core.User, error)  {
    f.called = true
	f.gotToken = token

	if f.err != nil {
		return core.User{}, f.err
	}

	return f.user, nil
}

func TestMiddlewareAuth(t *testing.T) {
	t.Run("without header", TestMiddlewareAuthWithoutHeader)
	t.Run("error format", TestMiddlewareAuthBadFormat)
	t.Run("invalid token", TestMiddlewareAuthInvalidToken)
	t.Run("valid token", TestMiddlewareAuthValidToken)
}

func TestMiddlewareAuthWithoutHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	verifier := &fakeVerifier{}

	handler := Auth(stubHandler, verifier)
	handler.ServeHTTP(rec, req)
	
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.False(t, verifier.called)
}

func TestMiddlewareAuthBadFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Token abc")
	rec := httptest.NewRecorder()
	verifier := &fakeVerifier{}

	handler := Auth(stubHandler, verifier)
	handler.ServeHTTP(rec, req)
	
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.False(t, verifier.called)
}


func TestMiddlewareAuthInvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	verifier := &fakeVerifier{err: errors.New("bad token")}

	handler := Auth(stubHandler, verifier)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.True(t, verifier.called)
	require.Equal(t, "abc", verifier.gotToken)
}

func TestMiddlewareAuthValidToken(t *testing.T) {
	expectedUser := core.User{
		ID:       "user-1",
		Username: "roman",
		Email:    "roman@example.com",
		Role:     core.RoleAdmin,
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(core.UserKey).(core.User)

		require.True(t, ok)
		require.Equal(t, expectedUser, user)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	verifier := &fakeVerifier{user: expectedUser}

	handler := Auth(next, verifier)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, verifier.called)
	require.Equal(t, "abc", verifier.gotToken)
}
