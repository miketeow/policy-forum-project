package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"policy-forum-backend/internal/auth"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// dependency injection container
type application struct {
	jwtSecret []byte
	db        *store.Queries
}

func main() {
	// load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to system environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required but not set")
	}

	// connect to database using pgxpool for concurrency safety
	dsn := "postgres://admin:password123@localhost:5432/policy_forum?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	log.Println("Successfully connect to the PostgreSQL database")

	// Initialize the application atruct with the sqlc-generated store
	app := &application{
		db:        store.New(pool),
		jwtSecret: []byte(jwtSecret),
	}

	// Create private router
	mux := http.NewServeMux()

	// Register health check handler with GET prefix
	mux.HandleFunc("GET /health", app.healthCheckHandler)
	mux.HandleFunc("POST /api/auth/register", app.registerHandler)
	mux.HandleFunc("POST /api/auth/login", app.loginHandler)
	mux.HandleFunc("POST /api/auth/logout", app.logoutHandler)
	mux.HandleFunc("GET /api/users/me", app.requireAuth(app.getUserProfileHandler))

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
	log.Printf("Starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK - System is running\n"))
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

		// move on to next handler
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
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// hash the password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	// log to prove connection worked
	// log.Printf("SUCCESS: Receive register attempt from name, %s, email: %s, password: %s\n", req.Name, req.Email, req.Password)
	// log.Printf("Hashed password successfully, here is the hashed password: %s\n", hashedPassword)

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
		log.Printf("Database error: %v", err)
		// in production, check for unique constraint violation (e.g., email already exists)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: User %s created with ID %s", user.Name, user.ID)

	// send json response back to nextjs frontend
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account created successfully",
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := app.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// compare plain text password with the hashed password
	match, err := auth.ComparePasswordAndHash(req.Password, user.HashedPassword)

	if err != nil || !match {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := auth.GenerateToken(app.jwtSecret, user.ID, user.KycStatus)
	if err != nil {

		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: User %s logged in successfully and received a token", user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"token":   tokenString,
	})
}

func (app *application) getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Critical Error: User ID missing from context", http.StatusInternalServerError)
		return
	}

	// fetch user profile from database
	user, err := app.db.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"message": "Logout successfully",
	}
	json.NewEncoder(w).Encode(response)
}
