package main

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (app *application) streamPostStatusHandler(w http.ResponseWriter, r *http.Request) {
	postIdStr := r.PathValue("postID")
	postID, err := uuid.Parse(postIdStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid post ID"))
		return
	}

	// SSE Headers, crucial for keeping the connection open
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		app.serverErrorResponse(w, r, errors.New("streaming unsupported by client"))
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			// user close tab or navigate away
			return
		case <-ticker.C:
			// check the database for the status of this job
			jobRow, checkErr := app.db.CheckSummaryJobStatus(ctx, postID.String())
			if checkErr != nil {
				if errors.Is(checkErr, pgx.ErrNoRows) {
					msg := "data: {\"status\":\"NONE\"}\n\n"
					_, err = w.Write([]byte(msg))
					if err != nil {
						app.LogError(ctx, "failed to write to sse stream", slog.String("error", err.Error()))
					}
					flusher.Flush()
					return
				}
				app.LogError(ctx, "failed to check job status", slog.String("error", checkErr.Error()))
				return
			}

			if jobRow.Status == "PENDING" || jobRow.Status == "PROCESSING" {
				if time.Since(jobRow.CreatedAt) > 3*time.Minute {
					app.LogWarn(ctx, "zombie job detected, marking as failed", slog.String("post_id", postID.String()))
					msg := "data: {\"status\":\"FAILED\"}\n\n"
					_, err = w.Write([]byte(msg))
					if err != nil {
						app.LogError(ctx, "failed to write to sse stream", slog.String("error", err.Error()))
					}
					flusher.Flush()
					return
				}
				continue
			}

			// if job is completed or failed, push the event to next js
			switch jobRow.Status {
			case "COMPLETED":
				// SSE required exact format
				msg := "data: {\"status\":\"COMPLETED\"}\n\n"
				_, err = w.Write([]byte(msg))
				if err != nil {
					app.LogError(ctx, "failed to write to sse stream", slog.String("error", err.Error()))
					return
				}
				flusher.Flush()
				// close connection
				return
			case "FAILED":
				msg := "data: {\"status\":\"FAILED\"}\n\n"
				_, err = w.Write([]byte(msg))
				if err != nil {
					app.LogError(ctx, "failed to write to sse stream", slog.String("error", err.Error()))
					return
				}
				flusher.Flush()
				// close connection
				return
			}

		}
	}
}
