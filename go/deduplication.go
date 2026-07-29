package main

import "strings"

func deduplicateJobs(jobs []JobListing) []JobListing {
	var uniqueJobs []JobListing
	seenURLs := make(map[string]bool)
	// seenComposites := make(map[string]bool)

	for _, job := range jobs {
		// grab just the link, no url params
		cleanURL := cleanURL(job.Link)

		if seenURLs[cleanURL] {
			continue
		}

		seenURLs[cleanURL] = true
		uniqueJobs = append(uniqueJobs, job)
	}

	return uniqueJobs
}

func cleanURL(raw string) string {
	clean := strings.TrimSpace(raw)
	clean = strings.Split(clean, "?")[0]
	clean = strings.Split(clean, "#")[0]
	return strings.ToLower(clean)
}
