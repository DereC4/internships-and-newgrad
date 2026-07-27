package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type JobListing struct {
	Company  string
	Role     string
	Location string
	Link     string
	Age      string
}

// You cannot use the short declaration operator := with constants.
const (
	jobTableStart = "<!-- A DerexXD certified divider lives here -->"
	jobTableEnd   = "<!-- End of DerexXD divider 1 -->"
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
		})
	}

	return jobs
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
		datePosted := cleanHTML(cols[5])

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
		})
	}

	return jobs
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
	}

	resultsChannel := make(chan string)
	// make a channel type so we can talk to main
	// loop through urls and start a thread for each one
	fmt.Println("Starting fetches")
	for _, url := range urls {
		// underscore is so we ignore the index, discard it
		go dogWorker(url, resultsChannel)
	}

	// you have to open the file before reading from channel
	file, err := os.OpenFile("testing.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	var totalJobs []JobListing

	for i := 0; i < len(urls); i++ {
		rawPayload := <-resultsChannel
		// separate the url from results
		fetchedURL, results, _ := strings.Cut(rawPayload, "|derexXD certified separator|")
		// channels will get consumed when you read them all one by one, so our two for loop approach was writing nothing
		fmt.Printf("--- Document Received #%d from %s ---\n", i+1, fetchedURL)
		// fmt.Println(results)
		fmt.Println("-------------------------------")

		// separator := fmt.Sprintf("\n\n# --- Document Received #%d ---\n\n", i+1)

		// if _, err := file.WriteString(separator); err != nil {
		// 	fmt.Printf("Error writing separator to file: %v\n", err)
		// }

		// if _, err := file.WriteString(results); err != nil {
		// 	fmt.Printf("Error writing content to file: %v\n", err)
		// }

		// fmt.Printf("Saved document #%d to combined_output.md\n", i+1)

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
		}
	}

	file.WriteString("| Company | Role | Location |\n")
	file.WriteString("| --- | --- | --- |\n")

	for _, job := range totalJobs {
		roleLink := fmt.Sprintf("[%s](%s)", job.Role, job.Link)
		formattedLine := fmt.Sprintf("%s | %s | %s | \n", job.Company, roleLink, job.Location)

		fmt.Println(formattedLine)

		if _, err := file.WriteString(formattedLine); err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
		}
	}

}
