package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// dependency injection container
type application struct {
	jwtSecret    []byte
	db           *store.Queries
	pool         *pgxpool.Pool
	geminiAPIKey string
	logger       *slog.Logger
}

func main() {

	// 1. INITIALIZE LOGGER
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)

	// 2. LOAD ENVIRONMENT
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found, falling back to system environment variables", slog.String("error", err.Error()))
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Error("JWT_SECRET environment variable is required but not set")
		os.Exit(1)
	}

	// Start OpenTelemetry
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider", slog.String("error", err.Error()))
		}
	}()

	// 3. CONNECT TO DATABASE
	dsn := "postgres://admin:password123@localhost:5432/policy_forum?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("Unable to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("Successfully connect to the PostgreSQL database")

	// 4. API KEYS
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		logger.Error("GEMINI_API_KEY is not set in .env")
		os.Exit(1)
	}

	// 5. DEPENDENCY INJECTION
	app := &application{
		db:           store.New(pool),
		pool:         pool,
		jwtSecret:    []byte(jwtSecret),
		geminiAPIKey: geminiKey,
		logger:       logger,
	}

	// Configure the HTTP server with strict timeout

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server
	app.logger.Info("database connection pool established")
	app.logger.Info("Starting server", slog.String("post", srv.Addr))

	if err := srv.ListenAndServe(); err != nil {
		app.logger.Error("server encountered a fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}

}
