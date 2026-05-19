package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"policy-forum-backend/internal/store"
	"policy-forum-backend/internal/worker"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/genai"
)

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

var validCategories = map[string]bool{
	"INFRASTRUCTURE": true,
	"ECONOMY":        true,
	"HEALTHCARE":     true,
	"EDUCATION":      true,
	"ENVIRONMENT":    true,
	"SAFETY":         true,
	"OTHER":          true,
}

func (app *application) categorizeWithAI(ctx context.Context, title, content string) string {
	tracer := otel.Tracer("ai-service")
	ctx, span := tracer.Start(ctx, "Gemini_Categorization")
	defer span.End()

	span.SetAttributes(
		attribute.String("post.title", title),
		attribute.Int("post.content_length", len(content)),
	)

	prompt := fmt.Sprintf(`You are a highly accurate automated categorization engine for a public policy forum.
			Read the user's post title and content below.

			You must classify the post into EXACTLY ONE of the following categories based on these rules:
			- INFRASTRUCTURE: Roads, public transit, zoning, housing, internet, urban planning.
			- ECONOMY: Jobs, taxes, small businesses, city budgets, inflation.
			- HEALTHCARE: Hospitals, public health, sanitation, mental health.
			- EDUCATION: Schools, teachers, libraries, youth programs.
			- ENVIRONMENT: Parks, pollution, waste management, climate, energy.
			- SAFETY: Police, fire services, emergency response, crime, traffic safety.
			- OTHER: Use ONLY if the post has absolutely zero relation to the above categories.

			RULES:
			Respond with ONLY the exact uppercase word of the category.
			Do not include any punctuation, explanations, or markdown.

			Title: %s
			Content: %s
		`, title, content)

	// build the JSON payload
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		app.LogError(ctx, "Failed to marshal Gemini request", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal JSON")
		return "OTHER"
	}

	// build the http request
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-lite-latest:generateContent?key=" + app.geminiAPIKey

	// use a 10-second timeout
	client := &http.Client{Timeout: 30 * time.Second}

	maxRetries := 3
	for i := 1; i <= maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			app.LogError(ctx, "failed to create HTTP request", slog.String("error", err.Error()))
			return "OTHER"
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			waitTime := time.Duration(i*2) * time.Second
			app.LogWarn(ctx, "network error communicating with gemini",
				slog.Int("attempt", i),
				slog.String("error", err.Error()),
				slog.String("retry_in", waitTime.String()))
			time.Sleep(waitTime)
			continue
		}

		// If Google is busy (503) or rate-limiting us (429), sleep and retry
		if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			waitTime := time.Duration(i*2) * time.Second
			app.LogWarn(ctx, "gemini api rate limited or busy",
				slog.Int("status_code", resp.StatusCode),
				slog.Int("attempt", i),
				slog.String("retry_in", waitTime.String()))
			time.Sleep(waitTime)
			continue
		}

		// If it's a hard error (like 400 Bad Request), fail immediately
		if resp.StatusCode != http.StatusOK {
			var errBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errBody)
			app.LogError(ctx, "gemini api returned hard http error",
				slog.Int("status_code", resp.StatusCode),
				slog.Any("api_error_body", errBody))
			resp.Body.Close()

			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d from Gemini", resp.StatusCode))
			return "OTHER"
		}

		// SUCCESS! Decode the response
		var apiResp geminiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			app.LogError(ctx, "failed to decode gemini JSON response", slog.String("error", err.Error()))
			resp.Body.Close()
			return "OTHER"
		}
		resp.Body.Close()

		// extract the text and clean it
		if len(apiResp.Candidates) > 0 && len(apiResp.Candidates[0].Content.Parts) > 0 {
			rawText := apiResp.Candidates[0].Content.Parts[0].Text
			app.LogInfo(ctx, "received successful AI categorization", slog.String("raw_response", rawText))
			cleanCategory := strings.TrimSpace(strings.ToUpper(rawText))
			cleanCategory = strings.TrimRight(cleanCategory, ".")

			// FIX: The Whitelist Firewall
			if validCategories[cleanCategory] {
				span.SetAttributes(
					attribute.String("ai.resolved_category", cleanCategory),
					attribute.Int("ai.retry_count", i),
				)
				span.SetStatus(codes.Ok, "Categorization successful")
				return cleanCategory
			}

			app.LogWarn(ctx, "AI returned invalid category format", slog.String("returned_category", cleanCategory))
			span.SetAttributes(
				attribute.String("ai.resolved_category", "OTHER"),
				attribute.String("ai.invalid_raw_response", cleanCategory),
				attribute.Int("ai.retry_count", i),
			)
			return "OTHER"
		}

		break // If we get here, the response was successfully parsed but empty
	}

	app.LogError(ctx, "Gemini API categorization failed completely", slog.Int("max_retries", maxRetries))
	// 4. FINAL FAILURE STATE
	span.SetStatus(codes.Error, "Max retries exhausted")
	span.SetAttributes(
		attribute.String("ai.resolved_category", "OTHER"), // Defaulted
		attribute.Int("ai.retry_count", maxRetries),
	)
	return "OTHER"

}

func (app *application) GenerateSummary(ctx context.Context, title, content string) (string, error) {

	// configure model parameter
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.2)),
	}

	prompt := fmt.Sprintf(`You are a neutral, professional analyst summarizing community feedback for a public policy forum.
Your goal is to extract the signal from the noise.

Read the following forum post:
Title: %s
Content:
%s

Provide a concise, single-paragraph executive summary based STRICTLY on these rules:
1. Identify the core issue, grievance, or topic being raised.
2. Identify the overall sentiment (e.g., frustrated, concerned, inquiring, supportive).
3. IF the user proposes a specific solution, include it. IF they are just complaining or asking a question, DO NOT invent a solution.
4. Ignore irrelevant rants, insults, or off-topic noise.

Return ONLY the summary paragraph. Do not include introductory phrases.`, title, content)

	resp, err := app.aiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("gemini network execution failed: %w", err)
	}

	finalSummary := strings.TrimSpace(resp.Text())

	if finalSummary == "" {
		return "", fmt.Errorf("gemini generated an empty summary string")
	}

	return finalSummary, nil
}

func (app *application) triggerSummaryHandler(w http.ResponseWriter, r *http.Request) {
	postIdStr := r.PathValue("postID")
	postID, err := uuid.Parse(postIdStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid post ID"))
		return
	}

	jobID := uuid.New()
	payload := worker.ExecSummaryPayload{PostID: postID}
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		app.serverErrorResponse(w, r, marshalErr)
		return
	}

	now := time.Now().UTC()
	enqueueErr := app.db.EnqueueJob(r.Context(), store.EnqueueJobParams{
		ID:        jobID,
		JobType:   "EXEC_SUMMARY",
		Payload:   payloadBytes,
		Status:    "PENDING",
		CreatedAt: now,
		UpdatedAt: now,
	})

	if enqueueErr != nil {
		app.serverErrorResponse(w, r, enqueueErr)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"message": "summary generation queued"})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GenerateCategoryReport(ctx context.Context, category string, promptDataBytes []byte) (string, error) {
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"trend_summary": {
				Type:        genai.TypeString,
				Description: "A concise executive summary of the primary concern in this category",
			},
			"overall_sentiment": {
				Type:        genai.TypeString,
				Description: "Must be exactly POSITIVE, NEGATIVE, or DIVIDED",
			},
			"actionable_insight": {
				Type:        genai.TypeString,
				Description: "Policy recommendation based on the top voted comment or the post",
			},
			"key_themes": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
				Description: "An array of 2 or 4 short phrases highlighting recurring themes",
			},
		},
		Required: []string{"trend_summary", "overall_sentiment", "actionable_insight", "key_themes"},
	}

	config := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr(float32(0.1)),
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	systemInstruction := fmt.Sprintf(`You are a lead data analyst for a state government.
Analyze the provided JSON containing the top upvoted civic forum posts and comments for the %s category.
Extract the core public sentiment and provide actionable intelligence.`, category)

	prompt := fmt.Sprintf("%s\n\nData:\n%s", systemInstruction, string(promptDataBytes))

	resp, err := app.aiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("gemini network execution failed: %w", err)
	}

	jsonResponse := strings.TrimSpace(resp.Text())
	if jsonResponse == "" {
		return "", fmt.Errorf("gemini generated an empty response")
	}

	return jsonResponse, nil
}

func (app *application) triggerCategoryReportHandler(w http.ResponseWriter, r *http.Request) {
	categoryParam := r.PathValue("category")
	cleanCategory := strings.ToUpper(strings.TrimSpace(categoryParam))

	if !validCategories[cleanCategory] {
		app.badRequestResponse(w, r, errors.New("invalid category specified"))
		return
	}

	payload := worker.CategoryReportPayload{
		Category: cleanCategory,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		app.serverErrorResponse(w, r, marshalErr)
		return
	}

	jobID := uuid.New()
	now := time.Now().UTC()

	enqueueErr := app.db.EnqueueJob(r.Context(), store.EnqueueJobParams{
		ID:        jobID,
		JobType:   "CATEGORY_REPORT",
		Payload:   payloadBytes,
		Status:    "PENDING",
		CreatedAt: now,
		UpdatedAt: now,
	})

	if enqueueErr != nil {
		app.serverErrorResponse(w, r, enqueueErr)
		return
	}

	app.logger.Info("Category report generation queued", slog.String("category", cleanCategory), slog.String("job_id", jobID.String()))

	err := app.writeJSON(w, http.StatusAccepted, envelope{
		"message": "Category report generation queued",
		"job_id":  jobID,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getCategoryReportHandler(w http.ResponseWriter, r *http.Request) {
	categoryParam := r.PathValue("category")
	cleanCategory := strings.ToUpper(strings.TrimSpace(categoryParam))

	if !validCategories[cleanCategory] {
		app.badRequestResponse(w, r, errors.New("invalid category specified"))
		return
	}

	reportRow, err := app.db.GetLatestCategoryReport(r.Context(), store.PostCategory(cleanCategory))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	response := struct {
		ID          string          `json:"id"`
		Category    string          `json:"category"`
		Report      json.RawMessage `json:"report"`
		GeneratedAt time.Time       `json:"generated_at"`
	}{
		ID:          reportRow.ID.String(),
		Category:    string(reportRow.Category),
		Report:      json.RawMessage(reportRow.Report),
		GeneratedAt: reportRow.GeneratedAt,
	}

	err = app.writeJSON(w, http.StatusOK, response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getReportStatusHandler(w http.ResponseWriter, r *http.Request) {
	categoryParam := strings.ToUpper(strings.TrimSpace(r.PathValue("category")))

	if !validCategories[categoryParam] {
		app.badRequestResponse(w, r, errors.New("invalid category"))
		return
	}

	isPending, err := app.db.CheckPendingReportJob(r.Context(), []byte(categoryParam))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"is_pending": isPending})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
