package main

import (
	"fmt"
	"net/http"
	"time"
)

func checkLink(urlStr string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Head(urlStr)
	if err != nil {
		fmt.Println("Network error or timeout")
		return true
	}
	defer resp.Body.Close()

	fmt.Printf("Final Status Code: %d\n", resp.StatusCode)
	if resp.StatusCode == 404 {
		fmt.Println("Job is dead (404 Not Found)")
		return false
	}

	finalURL := resp.Request.URL.String()
	fmt.Printf("Final Destination: %s\n", finalURL)

	if finalURL == "https://zapply.jobs/jobs" || finalURL == "https://zapply.jobs/jobs/" {
		fmt.Println("Job is dead (Redirected to default Zapply board)")
		return false
	}

	return true
}
