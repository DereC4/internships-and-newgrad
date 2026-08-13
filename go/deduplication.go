package main

import (
	"regexp"
	"strings"
)

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
	clean = strings.ToLower(clean)

	if strings.Contains(clean, "gh_jid=") {
		parts := strings.Split(clean, "gh_jid=")
		if len(parts) > 1 {
			jid := strings.Split(parts[1], "&")[0]
			return "greenhouse-job-id:" + jid
		}
	} else if strings.Contains(clean, "token=") && strings.Contains(clean, "greenhouse") {
		parts := strings.Split(clean, "token=")
		if len(parts) > 1 {
			token := strings.Split(parts[1], "&")[0]
			return "greenhouse-job-id:" + token
		}
	} else if strings.Contains(clean, "pid=") && strings.Contains(clean, "microsoft.com") {
		parts := strings.Split(clean, "pid=")
		if len(parts) > 1 {
			pid := strings.Split(parts[1], "&")[0]
			return "microsoft-job-id:" + pid
		}
	}

	clean = strings.Split(clean, "?")[0]
	clean = strings.Split(clean, "#")[0]

	clean = strings.ReplaceAll(clean, "://www.", "://")

	clean = strings.TrimSuffix(clean, "/")
	clean = strings.TrimSuffix(clean, "/apply")
	clean = strings.TrimSuffix(clean, "/application")
	clean = strings.TrimSuffix(clean, "/detail")
	clean = strings.TrimSuffix(clean, "/")

	if strings.Contains(clean, "myworkdayjobs.com") {
		re := regexp.MustCompile(`(myworkdayjobs\.com)/[a-z]{2}(?:-[a-z]{2})?/`)
		clean = re.ReplaceAllString(clean, "$1/")
	}

	if strings.Contains(clean, "greenhouse.io") {
		clean = strings.ReplaceAll(clean, "job-boards.greenhouse.io", "boards.greenhouse.io")
		clean = strings.ReplaceAll(clean, "job-boards.eu.greenhouse.io", "boards.greenhouse.io")
	}

	return clean
}
