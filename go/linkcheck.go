package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var linkClient = &http.Client{
	Timeout: 5 * time.Second,
}

func checkLink(urlStr string) bool {

	if !strings.Contains(strings.ToLower(urlStr), "zapply.jobs") {
		return true
	}

	resp, err := linkClient.Get(urlStr)

	// prevent accidental delete idk we'll do this later
	if err != nil {
		// fmt.Printf("[linkcheck] keep (network): %s (%v)\n", urlStr, err)
		return true
	}

	// go will only reuse connection if response body was read to the eof
	// 
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	finalURL := resp.Request.URL.String()
	trimmed := strings.ToLower(strings.TrimRight(finalURL, "/"))
	trimmed = strings.SplitN(trimmed, "?", 2)[0]
	if trimmed == "https://zapply.jobs/jobs" || trimmed == "http://zapply.jobs/jobs" {
		fmt.Printf("[linkcheck] drop: %s -> %s\n", urlStr, finalURL)
		return false
	}

	return true
}

func keepLiveJobs(jobs []JobListing) []JobListing {
	var (
		mu    sync.Mutex
		alive []JobListing
		wg    sync.WaitGroup
	)
	workers := make(chan struct{}, 20)

	for _, job := range jobs {
		wg.Add(1)
		go func(job JobListing) {
			defer wg.Done()

			workers <- struct{}{}
			keep := checkLink(job.Link)
			<-workers

			if !keep {
				return
			}

			mu.Lock()
			alive = append(alive, job)
			mu.Unlock()
		}(job)
	}

	wg.Wait()
	return alive
}
