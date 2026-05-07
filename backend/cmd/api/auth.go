package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"policy-forum-backend/internal/auth"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func (app *application) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanedErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanedErr)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	// hash the password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to hash password for user registration: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	// prepare the database transfer object
	now := time.Now().UTC()
	args := store.CreateUserParams{
		ID:             uuid.New(),
		Name:           req.Name,
		Email:          req.Email,
		HashedPassword: hashedPassword,
		KycStatus:      "UNVERIFIED",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	user, err := app.db.CreateUser(r.Context(), args)
	if err != nil {
		// declare a variable to hold the postgres specific error
		var pgErr *pgconn.PgError

		// look inside the error chain, extract the details
		if errors.As(err, &pgErr) {
			// check if the error code is 23505 (unique constraint violation)
			if pgErr.Code == "23505" {
				// this is user's fault, sent a safe 409 conflict message
				cleanErr := errors.New("a user with this email address already exists")
				app.errorResponse(w, r, http.StatusConflict, cleanErr.Error())
				return
			}
		}
		// if it was not unique constraint violation, sent server error
		wrappedErr := fmt.Errorf("failed to create user for user registration: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	app.LogInfo(r.Context(), "user created successfully", slog.String("user_name", user.Name), slog.String("user_id", user.ID.String()))

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "Acoount created successfully"})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanErr)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	user, err := app.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		app.LogWarn(r.Context(), "login failed due to email not found", slog.String("user_email", req.Email))

		app.invalidCredentialsResponse(w, r)
		return
	}

	// compare plain text password with the hashed password
	match, err := auth.ComparePasswordAndHash(req.Password, user.HashedPassword)

	if err != nil || !match {
		app.LogWarn(r.Context(), "login failed due to password mismatch", slog.String("user_email", user.Email))
		app.invalidCredentialsResponse(w, r)
		return
	}

	tokenString, err := auth.GenerateToken(app.jwtSecret, user.ID, user.KycStatus)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to generate JWT token for user %s: %w", user.ID, err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}
	app.LogInfo(r.Context(), "user login successfully", slog.String("user_email", user.Email))
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Login successful", "token": tokenString})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	app.writeJSON(w, http.StatusOK, envelope{"message": "Logout successfully"})
}
