package main

import (
	"context"
	"errors"
	"net/http"
	"policy-forum-backend/internal/auth"
	"strings"

	"github.com/google/uuid"
)

// create custom type for context key to prevent name collision in memory
type contextKey string

const userIDKey = contextKey("userID")

func (app *application) extractUserIDFromAuthHeader(r *http.Request) (uuid.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return uuid.Nil, errors.New("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil, errors.New("malformed authorization header")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.VerifyToken(app.jwtSecret, tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

func (app *application) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := app.extractUserIDFromAuthHeader(r)
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// injection: put uuid safely into the context
		ctx := context.WithValue(r.Context(), userIDKey, userID)

		// allow user to pass through with actual handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (app *application) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := app.extractUserIDFromAuthHeader(r)
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// injection: put uuid safely into the context
		ctx := context.WithValue(r.Context(), userIDKey, userID)

		// allow user to pass through with actual handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
