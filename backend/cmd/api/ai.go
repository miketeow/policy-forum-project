package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
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
				return cleanCategory
			}

			app.LogWarn(ctx, "AI returned invalid category format", slog.String("returned_category", cleanCategory))
			return "OTHER"
		}

		break // If we get here, the response was successfully parsed but empty
	}

	app.LogError(ctx, "Gemini API categorization failed completely", slog.Int("max_retries", maxRetries))
	return "OTHER"

}
