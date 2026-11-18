package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

// You cannot use the short declaration operator := with constants.
const (
	jobTableStart     = "<!-- A DerexXD certified divider lives here -->"
	jobTableEnd       = "<!-- End of DerexXD divider 1 -->"
	badgesStart       = "<!-- BADGES_START_DEREXXD -->"
	badgesEnd         = "<!-- BADGES_END_DEREXXD -->"
	newGradTableStart = "<!-- NEWGRAD_TABLE_START_DEREXXD -->"
	newGradTableEnd   = "<!-- NEWGRAD_TABLE_END_DEREXXD -->"
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
		"https://raw.githubusercontent.com/vanshb03/Summer2027-Internships/dev/.github/scripts/listings.json",
		"https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/refs/heads/dev/README.md",
		"https://raw.githubusercontent.com/speedyapply/2027-SWE-College-Jobs/refs/heads/main/README.md",
		"https://raw.githubusercontent.com/sndsh404/summer-2027-internships/refs/heads/main/README.md",
		"https://raw.githubusercontent.com/zapplyjobs/Internships-2027/refs/heads/main/README.md",
		"https://raw.githubusercontent.com/zshah101/Automated-List-Of-Summer-2027-and-Fall-2026-Tech-Internships/refs/heads/main/README.md",
		"https://raw.githubusercontent.com/SimplifyJobs/New-Grad-Positions/refs/heads/dev/README.md",
		"https://raw.githubusercontent.com/zapplyjobs/New-Grad-Software-Engineering-Jobs-2027/refs/heads/main/README.md",
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

		if strings.Contains(fetchedURL, "SimplifyJobs") && isNewGradSource(fetchedURL) {
			fmt.Println("Processing Simplify New Grad Repo...")
			jobs := parseSimplifyCategories(results, true)
			fmt.Printf("-> Found %d active jobs in Simplify New Grad Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "SimplifyJobs") {
			fmt.Println("Processing Simplify Repo...")
			jobs := parseSimplifyCategories(results, false)
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
		} else if strings.Contains(fetchedURL, "zapplyjobs") && isNewGradSource(fetchedURL) {
			fmt.Println("Processing zapplyjobs New Grad Repo...")
			jobs := parseZapply(results, true)
			fmt.Printf("-> Found %d active jobs in zapplyjobs New Grad Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "zapplyjobs") {
			fmt.Println("Processing zapplyjobs Repo...")
			jobs := parseZapply(results, false)
			fmt.Printf("-> Found %d active jobs in zapplyjobs Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "zshah101") {
			fmt.Println("Processing zshah101 Repo...")
			jobs := parseZShah(results)
			fmt.Printf("-> Found %d active jobs in zshah101 Repo\n", len(jobs))
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

	for i := range totalJobs {
		r := strings.ToLower(totalJobs[i].Role)
		if strings.Contains(r, "intern") || strings.Contains(r, "co-op") || strings.Contains(r, "coop") {
			totalJobs[i].NewGrad = false
		}
		totalJobs[i].Category = classifyRole(totalJobs[i].Role, totalJobs[i].Category)
	}

	// final output to README.md goes here
	var internshipJobs []JobListing
	for _, job := range totalJobs {
		if !job.NewGrad {
			internshipJobs = append(internshipJobs, job)
		}
	}

	uniqueJobs := keepLiveJobs(deduplicateJobs(internshipJobs))
	slices.SortFunc(uniqueJobs, compareJobListings)

	// update the fancy badges
	companySet := make(map[string]bool)
	for _, job := range uniqueJobs {
		companySet[job.Company] = true
	}
	uniqueCompanyCount := len(companySet)

	var badges strings.Builder
	badges.WriteString(fmt.Sprintf("![Job Listings](https://img.shields.io/badge/Aggregated_Jobs-%d-brightgreen?style=flat&logo=briefcase)\n", len(internshipJobs)))
	badges.WriteString(fmt.Sprintf("![Unique Listings](https://img.shields.io/badge/Unique_Listings-%d-brightgreen?style=flat&logo=briefcase)\n", len(uniqueJobs)))
	badges.WriteString(fmt.Sprintf("![Companies](https://img.shields.io/badge/Companies-%d-blue?style=flat&logo=building)\n", uniqueCompanyCount))

	writeTrackFiles(uniqueJobs, "internships", "README.md", "Internships")
	internshipMarkdown := renderMainMarkdown(uniqueJobs, false)

	content, _ := os.ReadFile("../README.md")
	updatedReadme := string(content)

	beforeTable, afterTableStart, _ := strings.Cut(updatedReadme, jobTableStart)
	_, afterTableEnd, _ := strings.Cut(afterTableStart, jobTableEnd)
	updatedReadme = beforeTable + jobTableStart + "\n\n" + internshipMarkdown + "\n" + jobTableEnd + afterTableEnd

	beforeBadges, afterBadgesStart, _ := strings.Cut(updatedReadme, badgesStart)
	_, afterBadgesEnd, _ := strings.Cut(afterBadgesStart, badgesEnd)
	updatedReadme = beforeBadges + badgesStart + "\n" + badges.String() + badgesEnd + afterBadgesEnd

	os.WriteFile("../README.md", []byte(updatedReadme), 0644)

	var newGradJobs []JobListing
	for _, job := range totalJobs {
		if job.NewGrad {
			newGradJobs = append(newGradJobs, job)
		}
	}

	uniqueNewGradJobs := keepLiveJobs(deduplicateJobs(newGradJobs))
	slices.SortFunc(uniqueNewGradJobs, compareJobListings)

	writeTrackFiles(uniqueNewGradJobs, "newgrad", "NEWGRAD.md", "New Grad Roles")
	newGradMarkdown := renderMainMarkdown(uniqueNewGradJobs, true)

	newGradContent, _ := os.ReadFile("../NEWGRAD.md")
	updatedNewGrad := string(newGradContent)

	beforeNewGradTable, afterNewGradTableStart, _ := strings.Cut(updatedNewGrad, newGradTableStart)
	_, afterNewGradTableEnd, _ := strings.Cut(afterNewGradTableStart, newGradTableEnd)
	updatedNewGrad = beforeNewGradTable + newGradTableStart + "\n\n" + newGradMarkdown + "\n" + newGradTableEnd + afterNewGradTableEnd

	os.WriteFile("../NEWGRAD.md", []byte(updatedNewGrad), 0644)

}

type categoryTrack struct {
	key      string
	title    string
	filename string
}

var coreTracks = []categoryTrack{
	{key: "SWE", title: "Software Engineering", filename: "SWE.md"},
	{key: "Quant", title: "Quantitative Finance", filename: "QUANT.md"},
	{key: "AI_ML", title: "AI & Machine Learning", filename: "AI_ML.md"},
	{key: "Data", title: "Data Science & Analytics", filename: "DATA.md"},
}

func writeTrackFiles(jobs []JobListing, subfolder, parentDoc, label string) {
	os.MkdirAll("../tracks/"+subfolder, 0755)

	categoryMap := make(map[string][]JobListing)
	for _, job := range jobs {
		categoryMap[job.Category] = append(categoryMap[job.Category], job)
	}

	for _, track := range coreTracks {
		catJobs := categoryMap[track.key]
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("# %s %s\n\n", track.title, label))
		sb.WriteString(fmt.Sprintf("> 🔙 [Back to All %s](../../%s)\n\n", label, parentDoc))
		sb.WriteString(fmt.Sprintf("Total Listings: **%d**\n\n", len(catJobs)))
		sb.WriteString("| Company | Role | Location | Age |\n")
		sb.WriteString("| --- | --- | --- | --- |\n")

		hasZeroDayJobs := false
		for _, job := range catJobs {
			if job.Age == "0d" {
				hasZeroDayJobs = true
				break
			}
		}

		dividerWritten := false
		for _, job := range catJobs {
			if hasZeroDayJobs && !dividerWritten && job.Age != "0d" {
				sb.WriteString("\n<!-- DerexXD 0d divider -->\n\n")
				sb.WriteString("| Company | Role | Location | Age |\n")
				sb.WriteString("| --- | --- | --- | --- |\n")
				dividerWritten = true
			}

			roleLink := fmt.Sprintf("[%s](%s)", job.Role, job.Link)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", job.Company, roleLink, job.Location, job.Age))
		}

		os.WriteFile(fmt.Sprintf("../tracks/%s/%s", subfolder, track.filename), []byte(sb.String()), 0644)
	}
}

func renderMainMarkdown(jobs []JobListing, isNewGrad bool) string {
	categoryMap := make(map[string][]JobListing)
	for _, job := range jobs {
		categoryMap[job.Category] = append(categoryMap[job.Category], job)
	}

	subfolder := "internships"
	if isNewGrad {
		subfolder = "newgrad"
	}

	var sb strings.Builder

	if isNewGrad {
		sb.WriteString("> 📍 **Browsing: 🎓 New Grad Listings** &nbsp;|&nbsp; 🔄 **Switch to: [💼 Internships](README.md)**\n\n")
	} else {
		sb.WriteString("> 📍 **Browsing: 💼 Internships** &nbsp;|&nbsp; 🔄 **Switch to: [🎓 New Grad Listings](NEWGRAD.md)**\n\n")
	}

	sb.WriteString("### Filter by Track\n")
	for _, track := range coreTracks {
		count := len(categoryMap[track.key])
		sb.WriteString(fmt.Sprintf("- [%s (%d)](tracks/%s/%s)\n", track.title, count, subfolder, track.filename))
	}
	sb.WriteString("\n---\n\n")

	sb.WriteString("| Company | Role | Location | Age |\n")
	sb.WriteString("| --- | --- | --- | --- |\n")

	hasZeroDayJobs := false
	for _, job := range jobs {
		if job.Age == "0d" {
			hasZeroDayJobs = true
			break
		}
	}

	dividerWritten := false
	for _, job := range jobs {
		if hasZeroDayJobs && !dividerWritten && job.Age != "0d" {
			sb.WriteString("\n<!-- DerexXD 0d divider -->\n\n")
			sb.WriteString("| Company | Role | Location | Age |\n")
			sb.WriteString("| --- | --- | --- | --- |\n")
			dividerWritten = true
		}

		roleLink := fmt.Sprintf("[%s](%s)", job.Role, job.Link)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", job.Company, roleLink, job.Location, job.Age))
	}

	return sb.String()
}
