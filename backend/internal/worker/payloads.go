package worker

import (
	"encoding/json"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusPending    JobStatus = "PENDING"
	StatusProcessing JobStatus = "PROCESSING"
	StatusCompleted  JobStatus = "COMPLETED"
	StatusFailed     JobStatus = "FAILED"
)

type JobType string

const (
	TypeExecSummary       JobType = "EXEC_SUMMARY"
	TypeSentimentAnalysis JobType = "SENTIMENT_ANALYSIS"
)

type ExecSummaryPayload struct {
	PostID uuid.UUID `json:"post_id"`
}

type SentimentAnalysisPayload struct {
	PostID uuid.UUID `json:"post_id"`
	Limit  int       `json:"limit"`
}

type AICategoryPromptData struct {
	Category string          `json:"category"`
	Posts    []AIPostContext `json:"posts"`
}

type AIPostContext struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Content     string          `json:"content"`
	Score       int             `json:"score"`
	TopComments json.RawMessage `json:"top_comments"`
}

type CategorySentimentReport struct {
	TrendSummary      string   `json:"trend_summary"`
	OverallSentiment  string   `json:"overall_sentiment"`
	ActionableInsight string   `json:"actionable_insight"`
	KeyThemes         []string `json:"key_themes"`
}
