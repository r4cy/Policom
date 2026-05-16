package middleware

import (
	"context"
	"net/http"
	"strings"

	"yadro.com/course/api/core"
)

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (core.User, error)
}

func Auth(next http.HandlerFunc, verifier TokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "You don't sent token", http.StatusUnauthorized)
			return
		}

		token := strings.SplitN(authHeader, " ", 2)
		if len(token) != 2 || token[0] != "Bearer" {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user, err := verifier.Verify(r.Context(), token[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), core.UserKey, user)
		next(w, r.WithContext(ctx))
	}
}
