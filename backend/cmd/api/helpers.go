package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type envelope map[string]any

type PaginationRequest struct {
	Limit  int
	Cursor time.Time
	Sort   string
	Offset int32
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func (app *application) parsePagination(r *http.Request) (PaginationRequest, error) {
	// default
	req := PaginationRequest{
		Limit:  20,
		Sort:   "desc",
		Offset: 0,
	}

	// limit parsing
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return req, errors.New("limit query parameter must be an integer")
		}
		if l < 1 || l > 100 {
			return req, errors.New("limit query parameter must be within 1 and 100")
		}
		req.Limit = l
	}

	// cursor parsing
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05.999999Z07:00",
			"2006-01-02T15:04:05.999999", // Common Postgres format (no Z)
			"2006-01-02 15:04:05.999999", // Space instead of T
		}

		var parsed time.Time
		var err error
		for _, layout := range layouts {
			parsed, err = time.Parse(layout, cursorStr)
			if err == nil {
				req.Cursor = parsed.UTC()
				break
			}
		}
		// if failed all layout, it is a bad request
		if err != nil {
			app.logger.Error("pagination error: failed to parse cursor", slog.String("error", err.Error()))
			// return clean error for frontend
			return req, errors.New("cursor query parameter is not a valid timestamp")
		}
	}

	// offset parsing
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			return req, errors.New("offset query paramater must be an integer")
		}
		if o < 0 {
			return req, errors.New("offset query paramater cannot be negative")
		}
		req.Offset = int32(o)
	}

	return req, nil
}

func (app *application) LogInfo(ctx context.Context, msg string, args ...any) {
	traceID, ok := ctx.Value(traceIDKey).(string)
	if ok {
		args = append(args, slog.String("trace_id", traceID))
	}
	app.logger.InfoContext(ctx, msg, args...)
}

func (app *application) LogError(ctx context.Context, msg string, args ...any) {
	traceID, ok := ctx.Value(traceIDKey).(string)
	if ok {
		args = append(args, slog.String("trace_id", traceID))
	}
	app.logger.ErrorContext(ctx, msg, args...)
}

func (app *application) LogWarn(ctx context.Context, msg string, args ...any) {
	traceID, ok := ctx.Value(traceIDKey).(string)
	if ok {
		args = append(args, slog.String("trace_id", traceID))
	}
	app.logger.WarnContext(ctx, msg, args...)
}
