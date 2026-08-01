package main

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

// You cannot use the short declaration operator := with constants.
const (
	jobTableStart = "<!-- A DerexXD certified divider lives here -->"
	jobTableEnd   = "<!-- End of DerexXD divider 1 -->"
	badgesStart   = "<!-- BADGES_START_DEREXXD -->"
	badgesEnd     = "<!-- BADGES_END_DEREXXD -->"
)

func dogWorker(url string, ch chan string) {
	// function signature in Go is variableName dataType
	// channels are type safe in Go so you have to define what type a channel takes
	response, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprintf("Error fetching %s: %v", url, err)
		return
	}

	bodyBytes, err := io.ReadAll(response.Body)

	if err != nil {
		ch <- fmt.Sprintf("Error reading body for %s: %v", url, err)
		return
	}

	ch <- url + "|derexXD certified separator|" + string(bodyBytes)

}

func main() {
	urls := []string{
		// this is a slice
		"https://raw.githubusercontent.com/vanshb03/Summer2027-Internships/dev/README.md",
		"https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/refs/heads/dev/README.md",
		"https://raw.githubusercontent.com/speedyapply/2027-SWE-College-Jobs/refs/heads/main/README.md",
		"https://raw.githubusercontent.com/sndsh404/summer-2027-internships/refs/heads/main/README.md",
		"https://raw.githubusercontent.com/zapplyjobs/Internships-2027/refs/heads/main/README.md",
	}

	resultsChannel := make(chan string)

	// make a channel type so we can talk to main
	// loop through urls and start a thread for each one
	fmt.Println("Starting fetches")
	for _, url := range urls {
		// underscore is so we ignore the index, discard it
		go dogWorker(url, resultsChannel)
	}

	var totalJobs []JobListing

	for i := 0; i < len(urls); i++ {
		rawPayload := <-resultsChannel

		// separate the url from results for each part of the payload
		fetchedURL, results, _ := strings.Cut(rawPayload, "|derexXD certified separator|")

		// channels will get consumed when you read them all one by one, so our two for loop approach was writing nothing
		fmt.Printf("--- Document Received #%d from %s ---\n", i+1, fetchedURL)

		if strings.Contains(fetchedURL, "SimplifyJobs") {
			fmt.Println("Processing Simplify Repo...")
			jobs := parseSimplifyCategories(results)
			fmt.Printf("-> Found %d active jobs in Simplify Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "vanshb03") {
			fmt.Println("Processing Vansh Repo...")
			jobs := parseVansh(results)
			fmt.Printf("-> Found %d active jobs in Vansh Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "speedyapply") {
			fmt.Println("Processing SpeedyApply Repo...")
			jobs := parseSpeedyApply(results)
			fmt.Printf("-> Found %d active jobs in SpeedyApply Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "sndsh404") {
			fmt.Println("Processing sndsh404 Repo...")
			jobs := parseSandesh(results)
			fmt.Printf("-> Found %d active jobs in sndsh404 Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "zapplyjobs") {
			fmt.Println("Processing zapplyjobs Repo...")
			jobs := parseSandesh(results)
			fmt.Printf("-> Found %d active jobs in zapplyjobs Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else {
			fmt.Println("[WARN] URL with no available parser found for " + fetchedURL)
		}
	}

	var table strings.Builder

	table.WriteString("| Company | Role | Location | Age |\n")
	table.WriteString("| --- | --- | --- | --- |\n")

	for _, job := range totalJobs {
		roleLink := fmt.Sprintf("[%s](%s)", job.Role, job.Link)
		table.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", job.Company, roleLink, job.Location, job.Age))
	}

	unfilteredContent := table.String()
	if err := os.WriteFile("UNFILTERED.md", []byte(unfilteredContent), 0644); err != nil {
		fmt.Printf("Error writing unfiltered.md: %v\n", err)
	} else {
		fmt.Printf("Saved %d raw jobs to UNFILTERED.md\n", len(totalJobs))
	}

	// final output to README.md goes here
	uniqueJobs := deduplicateJobs(totalJobs)
	table.Reset()
	table.WriteString("| Company | Role | Location | Age |\n")
	table.WriteString("| --- | --- | --- | --- |\n")

	slices.SortFunc(uniqueJobs, func(a, b JobListing) int {
		daysA := ageToDays(a.Age)
		daysB := ageToDays(b.Age)

		if ageComparison := cmp.Compare(daysA, daysB); ageComparison != 0 {
			return ageComparison
		}

		return cmp.Compare(a.Company, b.Company)
	})

	// update the fancy badges
	companySet := make(map[string]bool)
	for _, job := range uniqueJobs {
		companySet[job.Company] = true
	}
	uniqueCompanyCount := len(companySet)

	var badges strings.Builder
	badges.WriteString(fmt.Sprintf("![Job Listings](https://img.shields.io/badge/Total_Scraped-%d-brightgreen?style=flat&logo=briefcase)\n", len(totalJobs)))
	badges.WriteString(fmt.Sprintf("![Companies](https://img.shields.io/badge/Companies-%d-blue?style=flat&logo=building)\n", uniqueCompanyCount))

	for _, job := range uniqueJobs {
		roleLink := fmt.Sprintf("[%s](%s)", job.Role, job.Link)
		table.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", job.Company, roleLink, job.Location, job.Age))
	}

	content, _ := os.ReadFile("../README.md")
	updatedReadme := string(content)

	beforeTable, afterTableStart, _ := strings.Cut(updatedReadme, jobTableStart)
	_, afterTableEnd, _ := strings.Cut(afterTableStart, jobTableEnd)
	updatedReadme = beforeTable + jobTableStart + "\n\n" + table.String() + "\n" + jobTableEnd + afterTableEnd

	beforeBadges, afterBadgesStart, _ := strings.Cut(updatedReadme, badgesStart)
	_, afterBadgesEnd, _ := strings.Cut(afterBadgesStart, badgesEnd)
	updatedReadme = beforeBadges + badgesStart + "\n" + badges.String() + badgesEnd + afterBadgesEnd

	os.WriteFile("../README.md", []byte(updatedReadme), 0644)

}
