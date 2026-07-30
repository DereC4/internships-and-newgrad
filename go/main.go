package main

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

// You cannot use the short declaration operator := with constants.
const (
	jobTableStart = "<!-- A DerexXD certified divider lives here -->"
	jobTableEnd   = "<!-- End of DerexXD divider 1 -->"
	badgesStart   = "<!-- BADGES_START_DEREXXD -->"
	badgesEnd     = "<!-- BADGES_END_DEREXXD -->"
)

func cleanHTML(val string) string {
	val = strings.ReplaceAll(val, "<br>", " ")
	val = strings.ReplaceAll(val, "<br/>", " ")
	val = strings.ReplaceAll(val, "</br>", " ")

	for {
		start := strings.Index(val, "<")
		if start == -1 {
			break
		}

		end := strings.Index(val[start:], ">")
		if end == -1 {
			break
		}

		val = val[:start] + val[start+end+1:]
	}

	val = strings.ReplaceAll(val, "  ", " ")
	return strings.TrimSpace(val)
}

func parseSimplify(rawHTML string) []JobListing {
	var jobs []JobListing
	var lastCompany string

	_, afterTbody, foundStart := strings.Cut(rawHTML, "<tbody>")
	tbodyContent, _, foundEnd := strings.Cut(afterTbody, "</tbody>")
	if !foundStart || !foundEnd {
		return nil
	}

	rows := strings.Split(tbodyContent, "<tr>")

	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		row = strings.ReplaceAll(row, "<td align=\"center\">", "<td>")

		cols := strings.Split(row, "<td>")
		if len(cols) < 6 {
			continue
		}

		appCell := cols[4]

		if strings.Contains(appCell, "🔒") {
			continue
		}

		var appURL string
		if strings.Contains(appCell, "href=\"") {
			_, afterHref, _ := strings.Cut(appCell, "href=\"")
			appURL, _, _ = strings.Cut(afterHref, "\"")
		}

		if appURL == "" {
			continue
		}

		company := cleanHTML(cols[1])
		role := cleanHTML(cols[2])
		location := cleanHTML(cols[3])
		age := cleanHTML(cols[5])

		if company == "↳" || company == "" {
			company = lastCompany
		} else {
			lastCompany = company
		}

		jobs = append(jobs, JobListing{
			Company:  company,
			Role:     role,
			Location: location,
			Link:     appURL,
			Age:      age,
			NewGrad:  false,
		})
	}

	return jobs
}

func convertVanshDate(dateStr string) string {
	if dateStr == "" {
		return "N/A"
	}
	parsed, err := time.Parse("Jan 02 2006", dateStr+" 2026")
	if err != nil {
		return dateStr
	}

	days := int(time.Since(parsed).Hours() / 24)
	if days < 0 {
		days = 0
	}
	if days >= 30 {
		return fmt.Sprintf("%dmo", days/30)
	}
	return fmt.Sprintf("%dd", days)
}

func parseVansh(rawMarkdown string) []JobListing {
	var jobs []JobListing
	var lastCompany string

	// discard the before return string, we do not need that
	_, afterStart, foundStart := strings.Cut(rawMarkdown, "<!-- Please leave a one line gap between this and the table TABLE_START (DO NOT CHANGE THIS LINE) -->")
	tableContent, _, foundEnd := strings.Cut(afterStart, "[⬆️ Back to Top ⬆️]")

	if !foundStart || !foundEnd {
		return nil
	}

	lines := strings.Split(tableContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") {
			continue
		}

		cols := strings.Split(line, "|")
		if len(cols) < 6 {
			continue
		}

		// cols[0] = "" (prefix)
		// cols[1] = Company
		// cols[2] = Role
		// cols[3] = Location
		// cols[4] = Application / Link
		// cols[5] = Date Posted
		appCell := cols[4]

		if strings.Contains(appCell, "🔒") {
			continue
		}

		var appURL string
		if strings.Contains(appCell, "href=\"") {
			_, afterHref, _ := strings.Cut(appCell, "href=\"")
			appURL, _, _ = strings.Cut(afterHref, "\"")
		}

		if appURL == "" {
			continue
		}

		company := cleanHTML(cols[1])
		role := cleanHTML(cols[2])
		location := cleanHTML(cols[3])
		datePosted := convertVanshDate(cleanHTML(cols[5]))

		if company == "↳" || company == "" {
			company = lastCompany
		} else {
			lastCompany = company
		}

		jobs = append(jobs, JobListing{
			Company:  company,
			Role:     role,
			Location: location,
			Link:     appURL,
			Age:      datePosted,
			NewGrad:  false,
		})
	}

	return jobs
}

func parseSpeedyApply(rawMarkdown string) []JobListing {
	var jobs []JobListing

	sections := []struct{ start, end string }{
		{"<!-- TABLE_FAANG_START -->", "<!-- TABLE_FAANG_END -->"},
		{"<!-- TABLE_QUANT_START -->", "<!-- TABLE_QUANT_END -->"},
		{"<!-- TABLE_START -->", "<!-- TABLE_END -->"},
	}

	for _, sec := range sections {
		_, afterStart, foundStart := strings.Cut(rawMarkdown, sec.start)
		if !foundStart {
			continue
		}

		tableContent, _, foundEnd := strings.Cut(afterStart, sec.end)
		if !foundEnd {
			continue
		}

		for _, line := range strings.Split(tableContent, "\n") {
			line = strings.TrimSpace(line)

			if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") || strings.Contains(line, "Company") {
				continue
			}

			cols := strings.Split(line, "|")
			if len(cols) < 7 {
				continue
			}

			var appCell, ageCell string

			if len(cols) >= 8 {
				appCell = cols[5]
				ageCell = cols[6]
			} else {
				// cols[1]=Company, cols[2]=Role, cols[3]=Location, cols[4]=Link, cols[5]=Age
				appCell = cols[4]
				ageCell = cols[5]
			}

			if strings.Contains(appCell, "🔒") {
				continue
			}

			_, afterHref, foundHref := strings.Cut(appCell, "href=\"")
			if !foundHref {
				continue
			}
			appURL, _, _ := strings.Cut(afterHref, "\"")

			jobs = append(jobs, JobListing{
				Company:  cleanHTML(cols[1]),
				Role:     cleanHTML(cols[2]),
				Location: cleanHTML(cols[3]),
				Link:     appURL,
				Age:      cleanHTML(ageCell),
				NewGrad:  false,
			})
		}
	}

	return jobs
}

func ageToDays(ageStr string) int {
	clean := strings.TrimSpace(ageStr)

	if strings.HasSuffix(clean, "d") {
		numStr := strings.TrimSuffix(clean, "d")
		if val, err := strconv.Atoi(numStr); err == nil {
			return val
		}
	} else if strings.HasSuffix(clean, "mo") {
		numStr := strings.TrimSuffix(clean, "mo")
		if val, err := strconv.Atoi(numStr); err == nil {
			return val * 30 // Approximate 1 month = 30 days
		}
	}

	return 9999 // Push unknown or "N/A" formats to the very bottom of the table
}

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
		fmt.Println("-------------------------------")

		if strings.Contains(fetchedURL, "SimplifyJobs") {
			fmt.Println("Processing Simplify Repo...")

			// basically HashMap <String, String> categories
			// syntax: map [keytype] valuetype
			categories := map[string]string{
				"Software Engineering": "## 💻 Software Engineering Internship Roles",
				"Product Management":   "## 📱 Product Management Internship Roles",
				"Data Science / AI":    "## 🤖 Data Science, AI & Machine Learning Internship Roles",
				"Quant Finance":        "## 📈 Quantitative Finance Internship Roles",
			}

			for catName, header := range categories {
				_, afterHeader, foundHeader := strings.Cut(results, header)
				if !foundHeader {
					continue
				}

				tableBlock, _, _ := strings.Cut(afterHeader, "</table>")
				tableBlock = tableBlock + "</table>"
				parsedJobs := parseSimplify(tableBlock)

				fmt.Printf("-> Found %d jobs under category: [%s]\n", len(parsedJobs), catName)

				// for _, job := range parsedJobs {
				// 	fmt.Printf("   🏢 %s | 💼 %s | 📍 %s | 🔗 %s\n", job.Company, job.Role, job.Location, job.Link)
				// }

				fmt.Println()

				totalJobs = append(totalJobs, parsedJobs...)

			}
		} else if strings.Contains(fetchedURL, "vanshb03") {
			fmt.Println("Processing Vansh Repo...")
			jobs := parseVansh(results)
			fmt.Printf("-> Found %d active jobs in Vansh Repo\n", len(jobs))

			// for _, job := range jobs {
			// 	fmt.Printf("   🏢 %s | 💼 %s | 📍 %s | 🔗 %s\n", job.Company, job.Role, job.Location, job.Link)
			// }

			totalJobs = append(totalJobs, jobs...)
		} else if strings.Contains(fetchedURL, "speedyapply") {
			fmt.Println("Processing SpeedyApply Repo...")
			jobs := parseSpeedyApply(results)
			fmt.Printf("-> Found %d active jobs in SpeedyApply Repo\n", len(jobs))
			totalJobs = append(totalJobs, jobs...)
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
