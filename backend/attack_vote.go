package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	requests := 50

	// Replace with a valid post ID and a valid JWT token from your local DB
	postID := "da925be6-e094-422f-8301-3e0a6831e77b"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjJlZTU4NzEtMjQzNy00ZTA4LWExOGMtNTY1Y2E5ODBhNzdlIiwia3ljX3N0YXR1cyI6IlZFUklGSUVEIiwiaXNzIjoicHVibGljLXBvbGljeS1mb3J1bSIsImV4cCI6MTc3ODQyNDk5MiwiaWF0IjoxNzc4MzM4NTkyfQ.C6IbWgXHfRlkZi8eJmPyUUIZhzU36zsanLU6ZCwVKFg"
	url := fmt.Sprintf("http://localhost:8080/api/posts/%s/vote", postID)

	fmt.Printf("Firing %d concurrent votes...\n", requests)

	for range requests {
		wg.Go(func() {
			payload := strings.NewReader(`{"vote": 1}`)
			req, _ := http.NewRequest("POST", url, payload)
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err == nil {
				fmt.Printf("Response Status: %d\n", resp.StatusCode)
				resp.Body.Close()
			}
		})
	}

	// Wait for all 50 concurrent requests to finish
	wg.Wait()
	fmt.Println("Attack complete. Check your database!")
}
