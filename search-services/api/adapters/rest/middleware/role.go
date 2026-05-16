package middleware

import (
	"net/http"

	"yadro.com/course/api/core"
)

func RequireRole(role core.UserRole) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(core.UserKey).(core.User)
			if !ok || user.Role != role {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return 
			}
			next(w, r)
		}
	}
}