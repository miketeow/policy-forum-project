package main

import (
	"context"
	"log"
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

func (app *application) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// // TRACER 1: Did the header arrive at the Go server?
		// log.Printf("🔍 OPTIONAL AUTH: Received header length: %d", len(authHeader))

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Try to verify the token
			claims, err := auth.VerifyToken(app.jwtSecret, tokenString)

			if err != nil {
				// TRACER 2: The token arrived, but verification FAILED!
				log.Printf("❌ OPTIONAL AUTH: Token verification failed: %v", err)
			} else {
				// TRACER 3: Complete Success!
				// log.Printf("✅ OPTIONAL AUTH: Token verified for user: %v", claims.UserID)
				ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
				r = r.WithContext(ctx)
			}
		} else if authHeader != "" {
			// TRACER 4: Header arrived, but the word "Bearer " was missing
			log.Printf("⚠️ OPTIONAL AUTH: Header exists but doesn't start with 'Bearer '")
		}

		next.ServeHTTP(w, r)
	}
}
