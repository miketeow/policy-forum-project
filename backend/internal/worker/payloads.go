package worker

import "github.com/google/uuid"

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
