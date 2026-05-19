package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AIService interface {
	GenerateSummary(ctx context.Context, title, content string) (string, error)
	GenerateCategoryReport(ctx context.Context, category string, payloadBytes []byte) (string, error)
}
type Worker struct {
	db     *store.Queries
	logger *slog.Logger
	ai     AIService
}

func New(db *store.Queries, logger *slog.Logger, ai AIService) *Worker {
	return &Worker{
		db:     db,
		logger: logger,
		ai:     ai,
	}
}

// Start the infinite loop, until context is cancelled
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("Background worker started")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Background worker shutting down")
			return
		case <-ticker.C:
			w.processNextJob(ctx)
		}
	}
}

func (w *Worker) processNextJob(ctx context.Context) {
	// ask postgres for a job (use skip locked)
	now := time.Now().UTC()
	job, err := w.db.DequeueJob(ctx, now)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// no pending job, do nothing
			return
		}
		w.logger.Error("Failed to dequeue job", slog.String("error", err.Error()))
		return
	}

	w.logger.Info("Processing job", slog.String("job_id", job.ID.String()), slog.String("type", job.JobType))

	var processErr error

	// route the job to the correct logic based on job type
	switch job.JobType {
	case "EXEC_SUMMARY":
		processErr = w.handleExecSummary(ctx, job.Payload)
	case "CATEGORY_REPORT":
		processErr = w.handleCategoryReport(ctx, job.Payload)
	default:
		processErr = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	finishTime := time.Now().UTC()

	if processErr != nil {
		w.logger.Error("Job Failed", slog.String("job_id", job.ID.String()), slog.String("error", processErr.Error()))
		err := w.db.FailJob(ctx, store.FailJobParams{
			ID: job.ID,
			ErrorMessage: pgtype.Text{
				String: processErr.Error(),
				Valid:  true,
			},
			UpdatedAt: finishTime,
		})
		if err != nil {
			w.logger.Error("Critical: Failed to save FAILED status to database", slog.String("error", err.Error()))
		}
	} else {
		err := w.db.CompleteJob(ctx, store.CompleteJobParams{
			ID:        job.ID,
			UpdatedAt: finishTime,
		})
		if err != nil {
			w.logger.Error("Critical: Failed to save COMPLETED status to database", slog.String("error", err.Error()))
		} else {
			w.logger.Info("Job completed successfully", slog.String("job_id", job.ID.String()))
		}
	}
}

func (w *Worker) handleExecSummary(ctx context.Context, payloadBytes []byte) error {
	var payload ExecSummaryPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	post, err := w.db.GetPostForWorker(ctx, payload.PostID)
	if err != nil {
		return fmt.Errorf("failed to fetch post for summary: %w", err)
	}

	w.logger.Info("Generating AI Summary via Gemini...", slog.String("post_id", payload.PostID.String()))

	summaryText, err := w.ai.GenerateSummary(ctx, post.Title, post.Content)
	if err != nil {
		return fmt.Errorf("gemini api failed: %w", err)
	}
	finishTime := time.Now().UTC()

	err = w.db.UpdatePostSummary(ctx, store.UpdatePostSummaryParams{
		ID:        post.ID,
		Summary:   pgtype.Text{String: summaryText, Valid: true},
		UpdatedAt: finishTime,
	})
	if err != nil {
		return fmt.Errorf("failed to save summary to db: %w", err)
	}
	return nil
}

type CategoryReportPayload struct {
	Category string `json:"category"`
}

func (w *Worker) handleCategoryReport(ctx context.Context, jobPayloadBytes []byte) error {
	var payload CategoryReportPayload
	if err := json.Unmarshal(jobPayloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to parse category report payload: %w", err)
	}

	w.logger.Info("Fetching top posts and comments for category", slog.String("category", payload.Category))

	posts, err := w.db.GetTopPostsWithComments(ctx, store.PostCategory(payload.Category))
	if err != nil {
		return fmt.Errorf("failed to fetch category data: %w", err)
	}

	if len(posts) == 0 {
		w.logger.Info("No data found for category, skipping report", slog.String("category", payload.Category))
		return nil
	}

	// serialize the slice of db structs directly to JSON bytes
	promptDataBytes, err := json.Marshal(posts)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt data: %w", err)
	}

	w.logger.Info("Generating AI category Report via Gemini...", slog.String("category", payload.Category))

	jsonReportString, err := w.ai.GenerateCategoryReport(ctx, payload.Category, promptDataBytes)
	if err != nil {
		return fmt.Errorf("gemini api failed: %w", err)
	}

	finishTime := time.Now().UTC()

	err = w.db.SaveCategoryReport(ctx, store.SaveCategoryReportParams{
		ID:          uuid.New(),
		Category:    store.PostCategory(payload.Category),
		Report:      []byte(jsonReportString),
		GeneratedAt: finishTime,
	})

	if err != nil {
		return fmt.Errorf("failed to save category report to db: %w", err)
	}

	return nil

}
