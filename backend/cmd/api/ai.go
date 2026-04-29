package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

func (app *application) categorizeWithAI(title, content string) string {
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
	jsonData, _ := json.Marshal(reqBody)

	// build the http request
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-lite-latest:generateContent?key=" + app.geminiAPIKey

	// use a 10-second timeout
	client := &http.Client{Timeout: 30 * time.Second}

	maxRetries := 3
	for i := 1; i <= maxRetries; i++ {
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))

		// 2. FIX RETRY LOGIC: If network timeout, sleep and retry! Do NOT return!
		if err != nil {
			waitTime := time.Duration(i*2) * time.Second
			log.Printf("⚠️ Network error on attempt %d: %v. Retrying in %v...", i, err, waitTime)
			time.Sleep(waitTime)
			continue
		}

		// If Google is busy (503) or rate-limiting us (429), sleep and retry
		if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			waitTime := time.Duration(i*2) * time.Second
			log.Printf("⚠️ Gemini is busy (Status %d). Retrying in %v...", resp.StatusCode, waitTime)
			time.Sleep(waitTime)
			continue
		}

		// If it's a hard error (like 400 Bad Request), fail immediately
		if resp.StatusCode != http.StatusOK {
			var errBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errBody)
			log.Printf("❌ Gemini API HTTP Error %d: %v", resp.StatusCode, errBody)
			resp.Body.Close()
			return "OTHER"
		}

		// SUCCESS! Decode the response
		var apiResp geminiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			log.Printf("❌ Gemini API JSON Decode Error: %v", err)
			resp.Body.Close()
			return "OTHER"
		}
		resp.Body.Close()

		// extract the text and clean it
		if len(apiResp.Candidates) > 0 && len(apiResp.Candidates[0].Content.Parts) > 0 {
			rawText := apiResp.Candidates[0].Content.Parts[0].Text
			log.Printf("🧠 Raw AI Response: %q", rawText)
			cleanCategory := strings.TrimSpace(strings.ToUpper(rawText))
			return cleanCategory
		}

		break // If we get here, the response was successfully parsed but empty
	}

	log.Printf("❌ Gemini API failed after %d retries", maxRetries)
	return "OTHER"

}
