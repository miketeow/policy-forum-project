package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"policy-forum-backend/internal/auth"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// dependency injection container
type application struct {
	jwtSecret    []byte
	db           *store.Queries
	pool         *pgxpool.Pool
	geminiAPIKey string
	logger       *log.Logger
}

func main() {

	// 1. INITIALIZE LOGGER
	logger := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// 2. LOAD ENVIRONMENT
	if err := godotenv.Load(); err != nil {
		logger.Println("No .env file found, falling back to system environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required but not set")
	}

	// 3. CONNECT TO DATABASE
	dsn := "postgres://admin:password123@localhost:5432/policy_forum?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	logger.Println("Successfully connect to the PostgreSQL database")

	// 4. API KEYS
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		logger.Fatal("GEMINI_API_KEY is not set in .env")
	}

	// 5. DEPENDENCY INJECTION
	app := &application{
		db:           store.New(pool),
		pool:         pool,
		jwtSecret:    []byte(jwtSecret),
		geminiAPIKey: geminiKey,
		logger:       logger,
	}

	// =========================================================================
	// 6. ROUTER & MIDDLEWARE REGISTRATION
	// =========================================================================
	mux := http.NewServeMux()

	// --- SYSTEM & GLOBAL ROUTES ---
	mux.HandleFunc("GET /health", app.healthCheckHandler)
	mux.HandleFunc("GET /api/search", app.searchHandler)

	// --- AUTHENTICATION ---
	mux.HandleFunc("POST /api/auth/register", app.registerHandler)
	mux.HandleFunc("POST /api/auth/login", app.loginHandler)
	mux.HandleFunc("POST /api/auth/logout", app.requireAuth(app.logoutHandler))

	// --- USER PROFILE & DASHBOARD
	mux.HandleFunc("GET /api/users/me", app.requireAuth(app.getUserProfileHandler))
	mux.HandleFunc("GET /api/users/me/posts", app.requireAuth(app.getUserPostsHandler))
	mux.HandleFunc("GET /api/users/me/comments", app.requireAuth(app.getUserCommentsHandler))
	mux.HandleFunc("GET /api/users/me/upvoted/posts", app.requireAuth(app.getUserUpvotedPostsHandler))
	mux.HandleFunc("GET /api/users/me/upvoted/comments", app.requireAuth(app.getUserUpvotedCommentsHandler))

	// --- POSTS (Core Resource)
	// Public / Optional Auth (Read operation)
	mux.HandleFunc("GET /api/posts", app.optionalAuth(app.listPostHandler))
	mux.HandleFunc("GET /api/posts/{postId}", app.optionalAuth(app.getPostHandler))
	mux.HandleFunc("GET /api/posts/{postId}/comments", app.optionalAuth(app.getCommentsHandler))

	// Protected (Write operation)
	mux.HandleFunc("POST /api/posts", app.requireAuth(app.createPostHandler))
	mux.HandleFunc("PUT /api/posts/{postId}", app.requireAuth(app.updatePostHandler))
	mux.HandleFunc("DELETE /api/posts/{postId}", app.requireAuth(app.deletePostHandler))
	mux.HandleFunc("POST /api/posts/{postId}/vote", app.requireAuth(app.votePostHandler))
	mux.HandleFunc("POST /api/posts/{postId}/comments", app.requireAuth(app.createCommentHandler))

	// -- COMMENTS (Independent Actions)
	// Protected (Write operation)
	mux.HandleFunc("PUT /api/comments/{commentId}", app.requireAuth(app.updateCommentHandler))
	mux.HandleFunc("DELETE /api/comments/{commentId}", app.requireAuth(app.deleteCommentHandler))
	mux.HandleFunc("POST /api/comments/{commentId}/vote", app.requireAuth(app.voteCommentHandler))

	handlerWithCORS := corsMiddleware(mux)

	// Configure the HTTP server with strict timeout

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handlerWithCORS,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server
	app.logger.Printf("database connection pool established")
	app.logger.Printf("Starting server on port %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		app.logger.Fatal(err)
	}

}

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "OK - System is running\n")
}

// middleware to intercepts every incoming request, add CORS headers, and pass the request to the next handler

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// tell browser to allow requests from this origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		// tell browser to allow these HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// tell browser to allow these headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanedErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanedErr)
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

	app.logger.Printf("SUCCESS: User %s created with ID %s", user.Name, user.ID)

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "Acoount created successfully"})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanErr)
		return
	}

	user, err := app.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		app.logger.Printf("Login failed: Email not found (%s)", req.Email)
		app.invalidCredentialsResponse(w, r)
		return
	}

	// compare plain text password with the hashed password
	match, err := auth.ComparePasswordAndHash(req.Password, user.HashedPassword)

	if err != nil || !match {
		app.logger.Printf("Login failed: Password mismatch for user %s", user.Email)
		app.invalidCredentialsResponse(w, r)
		return
	}

	tokenString, err := auth.GenerateToken(app.jwtSecret, user.ID, user.KycStatus)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to generate JWT token for user %s: %w", user.ID, err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	app.logger.Printf("SUCCESS: User %s logged in successfully and received a token", user.Email)

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Login successful", "token": tokenString})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from contex")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	// fetch user profile from database
	user, err := app.db.GetUserByID(r.Context(), userID)
	if err != nil {
		// differentiate between not found and database error
		if errors.Is(err, pgx.ErrNoRows) {
			app.notFoundResponse(w, r)
			return
		}
		wrappedErr := fmt.Errorf("failed to fetch user profile: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user})
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
