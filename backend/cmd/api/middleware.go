package main

import (
	"context"
	"net/http"
	"policy-forum-backend/internal/auth"
	"strings"
)

// create custom type for context key to prevent name collision in memory
type contextKey string

const userIDKey = contextKey("userID")

func (app *application) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		// extract the token string
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// verify it
		claims, err := auth.VerifyToken(app.jwtSecret, tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// injection: put uuid safely into the context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)

		// create a new request with the updatedcontext
		req := r.WithContext(ctx)

		// allow user to pass through with actual handler
		next.ServeHTTP(w, req)
	}
}
